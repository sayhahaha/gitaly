//go:build !gitaly_test_sha256

package praefect

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v15/client"
	"gitlab.com/gitlab-org/gitaly/v15/internal/cache"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/gittest"
	gconfig "gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service"
	gitaly_metadata "gitlab.com/gitlab-org/gitaly/v15/internal/metadata"
	"gitlab.com/gitlab-org/gitaly/v15/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/v15/internal/middleware/metadatahandler"
	"gitlab.com/gitlab-org/gitaly/v15/internal/praefect/commonerr"
	"gitlab.com/gitlab-org/gitaly/v15/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v15/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v15/internal/praefect/grpc-proxy/proxy"
	"gitlab.com/gitlab-org/gitaly/v15/internal/praefect/nodes"
	"gitlab.com/gitlab-org/gitaly/v15/internal/praefect/protoregistry"
	"gitlab.com/gitlab-org/gitaly/v15/internal/praefect/transactions"
	"gitlab.com/gitlab-org/gitaly/v15/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper/promtest"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper/testdb"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v15/internal/transaction/txinfo"
	"gitlab.com/gitlab-org/gitaly/v15/internal/transaction/voting"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
	"gitlab.com/gitlab-org/labkit/correlation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var testLogger = logrus.New()

func init() {
	testLogger.SetOutput(io.Discard)
}

func TestStreamDirectorReadOnlyEnforcement(t *testing.T) {
	t.Parallel()
	db := testdb.New(t)
	for _, tc := range []struct {
		desc     string
		readOnly bool
	}{
		{desc: "writable", readOnly: false},
		{desc: "read-only", readOnly: true},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			db.TruncateAll(t)

			const (
				virtualStorage = "test-virtual-storage"
				relativePath   = "test-repository"
				storage        = "test-storage"
			)
			conf := config.Config{
				VirtualStorages: []*config.VirtualStorage{
					{
						Name: virtualStorage,
						Nodes: []*config.Node{
							{
								Address: "tcp://gitaly-primary.example.com",
								Storage: storage,
							},
						},
					},
				},
			}
			ctx := testhelper.Context(t)

			rs := datastore.MockRepositoryStore{
				GetConsistentStoragesFunc: func(context.Context, string, string) (string, map[string]struct{}, error) {
					if tc.readOnly {
						return "", map[string]struct{}{storage + "-other": {}}, nil
					}
					return "", map[string]struct{}{storage: {}}, nil
				},
			}

			coordinator := NewCoordinator(
				datastore.NewPostgresReplicationEventQueue(db),
				rs,
				NewNodeManagerRouter(&nodes.MockManager{GetShardFunc: func(vs string) (nodes.Shard, error) {
					require.Equal(t, virtualStorage, vs)
					return nodes.Shard{
						Primary: &nodes.MockNode{GetStorageMethod: func() string {
							return storage
						}},
					}, nil
				}}, rs),
				transactions.NewManager(conf),
				conf,
				protoregistry.GitalyProtoPreregistered,
			)

			frame, err := proto.Marshal(&gitalypb.DeleteRefsRequest{
				Repository: &gitalypb.Repository{
					StorageName:  virtualStorage,
					RelativePath: relativePath,
				},
				Refs: [][]byte{
					[]byte("refs/heads/does-not-exist"),
				},
			})
			require.NoError(t, err)

			_, err = coordinator.StreamDirector(ctx, "/gitaly.RefService/DeleteRefs", &mockPeeker{frame: frame})
			if tc.readOnly {
				testhelper.RequireGrpcError(t, ErrRepositoryReadOnly, err)
				testhelper.RequireGrpcCode(t, err, codes.FailedPrecondition)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStreamDirectorMutator(t *testing.T) {
	t.Parallel()
	gitalySocket0, gitalySocket1 := testhelper.GetTemporaryGitalySocketFileName(t), testhelper.GetTemporaryGitalySocketFileName(t)
	testhelper.NewServerWithHealth(t, gitalySocket0)
	testhelper.NewServerWithHealth(t, gitalySocket1)

	primaryAddress, secondaryAddress := "unix://"+gitalySocket0, "unix://"+gitalySocket1
	primaryNode := &config.Node{Address: primaryAddress, Storage: "praefect-internal-1"}
	secondaryNode := &config.Node{Address: secondaryAddress, Storage: "praefect-internal-2"}
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name:  "praefect",
				Nodes: []*config.Node{primaryNode, secondaryNode},
			},
		},
	}
	db := testdb.New(t)

	targetRepo := gitalypb.Repository{
		StorageName:  "praefect",
		RelativePath: "/path/to/hashed/storage",
	}
	ctx := testhelper.Context(t)

	txMgr := transactions.NewManager(conf)

	nodeSet, err := DialNodes(ctx, conf.VirtualStorages, protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	defer nodeSet.Close()

	for _, tc := range []struct {
		desc             string
		repositoryExists bool
		error            error
	}{
		{
			desc:             "succcessful",
			repositoryExists: true,
		},
		{
			desc:  "repository not found",
			error: structerr.NewNotFound("%w", fmt.Errorf("mutator call: route repository mutator: %w", fmt.Errorf("get repository id: %w", commonerr.NewRepositoryNotFoundError(targetRepo.StorageName, targetRepo.RelativePath)))),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			tx := db.Begin(t)
			defer tx.Rollback(t)

			rs := datastore.NewPostgresRepositoryStore(tx, conf.StorageNames())

			if tc.repositoryExists {
				require.NoError(t, rs.CreateRepository(ctx, 1, targetRepo.StorageName, targetRepo.RelativePath, targetRepo.RelativePath, primaryNode.Storage, []string{secondaryNode.Storage}, nil, true, true))
			}

			testdb.SetHealthyNodes(t, ctx, tx, map[string]map[string][]string{"praefect": conf.StorageNames()})
			queueInterceptor := datastore.NewReplicationEventQueueInterceptor(datastore.NewPostgresReplicationEventQueue(db))
			queueInterceptor.OnEnqueue(func(ctx context.Context, event datastore.ReplicationEvent, queue datastore.ReplicationEventQueue) (datastore.ReplicationEvent, error) {
				assert.True(t, len(queueInterceptor.GetEnqueued()) < 2, "expected only one event to be created")
				return queue.Enqueue(ctx, event)
			})

			coordinator := NewCoordinator(
				queueInterceptor,
				rs,
				NewPerRepositoryRouter(
					nodeSet.Connections(),
					nodes.NewPerRepositoryElector(tx),
					StaticHealthChecker(conf.StorageNames()),
					NewLockedRandom(rand.New(rand.NewSource(0))),
					rs,
					datastore.NewAssignmentStore(tx, conf.StorageNames()),
					rs,
					nil,
				),
				txMgr,
				conf,
				protoregistry.GitalyProtoPreregistered,
			)

			frame, err := proto.Marshal(&gitalypb.FetchIntoObjectPoolRequest{
				Origin:     &targetRepo,
				ObjectPool: &gitalypb.ObjectPool{Repository: &targetRepo},
			})
			require.NoError(t, err)

			require.NoError(t, err)

			fullMethod := "/gitaly.ObjectPoolService/FetchIntoObjectPool"

			peeker := &mockPeeker{frame}
			streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
			if tc.error != nil {
				require.Equal(t, tc.error, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, primaryAddress, streamParams.Primary().Conn.Target())

			mi, err := coordinator.registry.LookupMethod(fullMethod)
			require.NoError(t, err)

			m, err := mi.UnmarshalRequestProto(streamParams.Primary().Msg)
			require.NoError(t, err)

			rewrittenTargetRepo, err := mi.TargetRepo(m)
			require.NoError(t, err)
			require.Equal(t, "praefect-internal-1", rewrittenTargetRepo.GetStorageName(), "stream director should have rewritten the storage name")

			// this call creates new events in the queue and simulates usual flow of the update operation
			require.NoError(t, streamParams.RequestFinalizer())

			// wait until event persisted (async operation)
			require.NoError(t, queueInterceptor.Wait(time.Minute, func(i *datastore.ReplicationEventQueueInterceptor) bool {
				return len(i.GetEnqueuedResult()) == 1
			}))

			events, err := queueInterceptor.Dequeue(ctx, "praefect", "praefect-internal-2", 10)
			require.NoError(t, err)
			require.Len(t, events, 1)

			expectedEvent := datastore.ReplicationEvent{
				ID:        1,
				State:     datastore.JobStateInProgress,
				Attempt:   2,
				LockID:    "praefect|praefect-internal-2|/path/to/hashed/storage",
				CreatedAt: events[0].CreatedAt,
				UpdatedAt: events[0].UpdatedAt,
				Job: datastore.ReplicationJob{
					RepositoryID:      1,
					Change:            datastore.UpdateRepo,
					VirtualStorage:    conf.VirtualStorages[0].Name,
					RelativePath:      targetRepo.RelativePath,
					TargetNodeStorage: secondaryNode.Storage,
					SourceNodeStorage: primaryNode.Storage,
				},
				Meta: datastore.Params{metadatahandler.CorrelationIDKey: "my-correlation-id"},
			}
			require.Equal(t, expectedEvent, events[0], "ensure replication job created by stream director is correct")
		})
	}
}

func TestStreamDirectorMutator_StopTransaction(t *testing.T) {
	t.Parallel()
	socket := testhelper.GetTemporaryGitalySocketFileName(t)
	testhelper.NewServerWithHealth(t, socket)

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "praefect",
				Nodes: []*config.Node{
					{Address: "unix://" + socket, Storage: "primary"},
					{Address: "unix://" + socket, Storage: "secondary"},
				},
			},
		},
	}

	repo := gitalypb.Repository{
		StorageName:  "praefect",
		RelativePath: "/path/to/hashed/storage",
	}

	nodeMgr, err := nodes.NewManager(testhelper.NewDiscardingLogEntry(t), conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Hour)
	defer nodeMgr.Stop()
	ctx := testhelper.Context(t)

	shard, err := nodeMgr.GetShard(ctx, conf.VirtualStorages[0].Name)
	require.NoError(t, err)

	for _, name := range []string{"primary", "secondary"} {
		node, err := shard.GetNode(name)
		require.NoError(t, err)
		waitNodeToChangeHealthStatus(t, ctx, node, true)
	}

	rs := datastore.MockRepositoryStore{
		GetConsistentStoragesFunc: func(ctx context.Context, virtualStorage, relativePath string) (string, map[string]struct{}, error) {
			return relativePath, map[string]struct{}{"primary": {}, "secondary": {}}, nil
		},
	}

	txMgr := transactions.NewManager(conf)

	coordinator := NewCoordinator(
		datastore.NewPostgresReplicationEventQueue(testdb.New(t)),
		rs,
		NewNodeManagerRouter(nodeMgr, rs),
		txMgr,
		conf,
		protoregistry.GitalyProtoPreregistered,
	)

	fullMethod := "/gitaly.SmartHTTPService/PostReceivePack"

	frame, err := proto.Marshal(&gitalypb.PostReceivePackRequest{
		Repository: &repo,
	})
	require.NoError(t, err)
	peeker := &mockPeeker{frame}

	streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
	require.NoError(t, err)

	txCtx := peer.NewContext(streamParams.Primary().Ctx, &peer.Peer{})
	transaction, err := txinfo.TransactionFromContext(txCtx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	var syncWG sync.WaitGroup

	wg.Add(2)
	syncWG.Add(2)

	go func() {
		defer wg.Done()

		vote := voting.VoteFromData([]byte("vote"))
		err := txMgr.VoteTransaction(ctx, transaction.ID, "primary", vote)
		require.NoError(t, err)

		// Assure that at least one vote was agreed on.
		syncWG.Done()
		syncWG.Wait()

		require.NoError(t, txMgr.StopTransaction(ctx, transaction.ID))
	}()

	go func() {
		defer wg.Done()

		vote := voting.VoteFromData([]byte("vote"))
		err := txMgr.VoteTransaction(ctx, transaction.ID, "secondary", vote)
		require.NoError(t, err)

		// Assure that at least one vote was agreed on.
		syncWG.Done()
		syncWG.Wait()

		err = txMgr.VoteTransaction(ctx, transaction.ID, "secondary", vote)
		assert.True(t, errors.Is(err, transactions.ErrTransactionStopped))
	}()

	wg.Wait()

	err = streamParams.RequestFinalizer()
	require.NoError(t, err)
}

func TestStreamDirectorMutator_SecondaryErrorHandling(t *testing.T) {
	t.Parallel()
	testhelper.NewFeatureSets(featureflag.NodeErrorCancelsVoter).Run(t, testStreamDirectorMutatorSecondaryErrorHandling)
}

func testStreamDirectorMutatorSecondaryErrorHandling(t *testing.T, ctx context.Context) {
	ctx, ctxCancel := context.WithCancel(ctx)

	socket := testhelper.GetTemporaryGitalySocketFileName(t)
	testhelper.NewServerWithHealth(t, socket)

	primaryNode := &config.Node{Address: "unix://" + socket, Storage: "praefect-internal-1"}
	secondaryNode1 := &config.Node{Address: "unix://" + socket, Storage: "praefect-internal-2"}
	secondaryNode2 := &config.Node{Address: "unix://" + socket, Storage: "praefect-internal-3"}

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name:  "praefect",
				Nodes: []*config.Node{primaryNode, secondaryNode1, secondaryNode2},
			},
		},
	}

	repo := gitalypb.Repository{
		StorageName:  "praefect",
		RelativePath: "/path/to/hashed/storage",
	}

	nodeMgr, err := nodes.NewManager(testhelper.NewDiscardingLogEntry(t), conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Hour)
	defer nodeMgr.Stop()

	shard, err := nodeMgr.GetShard(ctx, conf.VirtualStorages[0].Name)
	require.NoError(t, err)

	for _, name := range []string{"praefect-internal-1", "praefect-internal-2", "praefect-internal-3"} {
		node, err := shard.GetNode(name)
		require.NoError(t, err)
		waitNodeToChangeHealthStatus(t, ctx, node, true)
	}

	rs := datastore.MockRepositoryStore{
		GetConsistentStoragesFunc: func(ctx context.Context, virtualStorage, relativePath string) (string, map[string]struct{}, error) {
			return relativePath, map[string]struct{}{"praefect-internal-1": {}, "praefect-internal-2": {}, "praefect-internal-3": {}}, nil
		},
	}

	txMgr := transactions.NewManager(conf)

	coordinator := NewCoordinator(
		datastore.NewPostgresReplicationEventQueue(testdb.New(t)),
		rs,
		NewNodeManagerRouter(nodeMgr, rs),
		txMgr,
		conf,
		protoregistry.GitalyProtoPreregistered,
	)

	fullMethod := "/gitaly.SmartHTTPService/PostReceivePack"

	frame, err := proto.Marshal(&gitalypb.PostReceivePackRequest{
		Repository: &repo,
	})
	require.NoError(t, err)
	peeker := &mockPeeker{frame}

	streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
	require.NoError(t, err)

	txCtx := peer.NewContext(streamParams.Primary().Ctx, &peer.Peer{})
	transaction, err := txinfo.TransactionFromContext(txCtx)
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		vote := voting.VoteFromData([]byte("vote"))
		err := txMgr.VoteTransaction(ctx, transaction.ID, "praefect-internal-1", vote)
		if featureflag.NodeErrorCancelsVoter.IsEnabled(ctx) {
			require.ErrorIs(t, err, transactions.ErrTransactionFailed)
		} else {
			require.EqualError(t, err, "context canceled")
		}
	}()

	for _, secondary := range streamParams.Secondaries() {
		wg.Add(1)
		secondary := secondary
		go func() {
			err := secondary.ErrHandler(errors.New("node RPC failure"))
			require.NoError(t, err)

			wg.Done()
		}()
	}

	// If the feature flag is not enabled the context must be canceled to unblock voters.
	if !featureflag.NodeErrorCancelsVoter.IsEnabled(ctx) {
		ctxCancel()
	}

	wg.Wait()

	err = streamParams.RequestFinalizer()
	require.NoError(t, err)
	ctxCancel()
}

func TestStreamDirectorMutator_ReplicateRepository(t *testing.T) {
	ctx := testhelper.Context(t)

	socket := testhelper.GetTemporaryGitalySocketFileName(t)
	testhelper.NewServerWithHealth(t, socket)

	// Setup config with two virtual storages.
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name:  "praefect-1",
				Nodes: []*config.Node{{Address: "unix://" + socket, Storage: "praefect-internal-1"}},
			},
			{
				Name:  "praefect-2",
				Nodes: []*config.Node{{Address: "unix://" + socket, Storage: "praefect-internal-2"}},
			},
		},
	}

	nodeMgr, err := nodes.NewManager(testhelper.NewDiscardingLogEntry(t), conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Hour)
	defer nodeMgr.Stop()

	incrementGenerationInvoked := false
	rs := datastore.MockRepositoryStore{
		GetConsistentStoragesFunc: func(ctx context.Context, virtualStorage, relativePath string) (string, map[string]struct{}, error) {
			return relativePath, map[string]struct{}{"praefect-internal-2": {}}, nil
		},
		CreateRepositoryFunc: func(ctx context.Context, repositoryID int64, virtualStorage, relativePath, replicaPath, primary string, updatedSecondaries, outdatedSecondaries []string, storePrimary, storeAssignments bool) error {
			require.Fail(t, "CreateRepository should not be called")
			return nil
		},
		IncrementGenerationFunc: func(ctx context.Context, repositoryID int64, primary string, secondaries []string) error {
			incrementGenerationInvoked = true
			return nil
		},
	}

	router := mockRouter{
		// Simulate scenario where target repository already exists and error is returned.
		routeRepositoryCreation: func(ctx context.Context, virtualStorage, relativePath, additionalRepoRelativePath string) (RepositoryMutatorRoute, error) {
			return RepositoryMutatorRoute{}, fmt.Errorf("reserve repository id: %w", commonerr.ErrRepositoryAlreadyExists)
		},
		// Pass through normally to handle route creation.
		routeRepositoryMutator: func(ctx context.Context, virtualStorage, relativePath, additionalRepoRelativePath string) (RepositoryMutatorRoute, error) {
			return NewNodeManagerRouter(nodeMgr, rs).RouteRepositoryMutator(ctx, virtualStorage, relativePath, additionalRepoRelativePath)
		},
	}

	coordinator := NewCoordinator(
		&datastore.MockReplicationEventQueue{},
		rs,
		router,
		transactions.NewManager(conf),
		conf,
		protoregistry.GitalyProtoPreregistered,
	)

	fullMethod := "/gitaly.RepositoryService/ReplicateRepository"

	frame, err := proto.Marshal(&gitalypb.ReplicateRepositoryRequest{
		Repository: &gitalypb.Repository{
			StorageName:  "praefect-2",
			RelativePath: "/path/to/hashed/storage",
		},
		Source: &gitalypb.Repository{
			StorageName:  "praefect-1",
			RelativePath: "/path/to/hashed/storage",
		},
	})
	require.NoError(t, err)
	peeker := &mockPeeker{frame}

	// Validate that stream parameters can be constructed successfully for
	// `ReplicateRepository` when the target repository already exists.
	streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
	require.NoError(t, err)

	// Validate that `CreateRepository()` is not invoked and `IncrementGeneration()`
	// is when target repository already exists.
	err = streamParams.RequestFinalizer()
	require.NoError(t, err)
	require.True(t, incrementGenerationInvoked)
}

func TestStreamDirector_maintenance(t *testing.T) {
	t.Parallel()

	node1 := &config.Node{
		Address: "unix://" + testhelper.GetTemporaryGitalySocketFileName(t),
		Storage: "praefect-internal-1",
	}

	node2 := &config.Node{
		Address: "unix://" + testhelper.GetTemporaryGitalySocketFileName(t),
		Storage: "praefect-internal-2",
	}

	cfg := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name:  "praefect",
				Nodes: []*config.Node{node1, node2},
			},
		},
	}

	db := testdb.New(t)

	repo := gitalypb.Repository{
		StorageName:  "praefect",
		RelativePath: "/path/to/hashed/storage",
	}

	ctx := testhelper.Context(t)

	nodeSet, err := DialNodes(ctx, cfg.VirtualStorages, protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	defer nodeSet.Close()

	tx := db.Begin(t)
	defer tx.Rollback(t)

	rs := datastore.NewPostgresRepositoryStore(tx, cfg.StorageNames())
	require.NoError(t, rs.CreateRepository(ctx, 1, repo.StorageName, repo.RelativePath,
		repo.RelativePath, node1.Storage, []string{node2.Storage}, nil, true, true))

	testdb.SetHealthyNodes(t, ctx, tx, map[string]map[string][]string{"praefect": cfg.StorageNames()})

	queueInterceptor := datastore.NewReplicationEventQueueInterceptor(datastore.NewPostgresReplicationEventQueue(tx))

	coordinator := NewCoordinator(
		queueInterceptor,
		rs,
		NewPerRepositoryRouter(
			nodeSet.Connections(),
			nodes.NewPerRepositoryElector(tx),
			StaticHealthChecker(cfg.StorageNames()),
			NewLockedRandom(rand.New(rand.NewSource(0))),
			rs,
			datastore.NewAssignmentStore(tx, cfg.StorageNames()),
			rs,
			nil,
		),
		nil,
		cfg,
		protoregistry.GitalyProtoPreregistered,
	)

	message, err := proto.Marshal(&gitalypb.OptimizeRepositoryRequest{
		Repository: &repo,
	})
	require.NoError(t, err)

	methodInfo, err := protoregistry.GitalyProtoPreregistered.LookupMethod("/gitaly.RepositoryService/OptimizeRepository")
	require.NoError(t, err)

	for _, tc := range []struct {
		desc         string
		primaryErr   error
		secondaryErr error
		expectedErr  error
	}{
		{
			desc: "successful",
		},
		{
			desc:        "primary returns an error",
			primaryErr:  structerr.NewNotFound("primary error"),
			expectedErr: structerr.NewNotFound("primary error"),
		},
		{
			desc:         "secondary returns an error",
			secondaryErr: structerr.NewNotFound("secondary error"),
			expectedErr:  structerr.NewNotFound("secondary error"),
		},
		{
			desc:         "primary error preferred",
			primaryErr:   structerr.NewNotFound("primary error"),
			secondaryErr: structerr.NewNotFound("secondary error"),
			expectedErr:  structerr.NewNotFound("primary error"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			// Behaviour between the new and old routing are sufficiently different that
			// it doesn't make sense to try and cram both tests in here.
			queueInterceptor.OnEnqueue(func(context.Context, datastore.ReplicationEvent, datastore.ReplicationEventQueue) (datastore.ReplicationEvent, error) {
				require.FailNow(t, "no replication jobs should have been created")
				return datastore.ReplicationEvent{}, fmt.Errorf("unexpected call")
			})

			streamParams, err := coordinator.StreamDirector(ctx, methodInfo.FullMethodName(), &mockPeeker{message})
			require.NoError(t, err)

			var targetNodes []string
			destinationByAddress := make(map[string]proxy.Destination)
			for _, node := range append(streamParams.Secondaries(), streamParams.Primary()) {
				targetNodes = append(targetNodes, node.Conn.Target())
				destinationByAddress[node.Conn.Target()] = node
			}

			// Assert that both nodes are part of the stream parameters. Because the
			// order is not deterministic (we randomly shuffle primary and secondary
			// nodes) we only assert that elements match, not that any of both nodes has
			// a specific role.
			require.ElementsMatch(t, []string{
				node1.Address,
				node2.Address,
			}, targetNodes)

			// Assert that the target repositories were rewritten as expected for all of
			// the nodes.
			for _, nodeCfg := range []*config.Node{node1, node2} {
				destination, ok := destinationByAddress[nodeCfg.Address]
				require.True(t, ok)

				request, err := methodInfo.UnmarshalRequestProto(destination.Msg)
				require.NoError(t, err)

				rewrittenTargetRepo, err := methodInfo.TargetRepo(request)
				require.NoError(t, err)
				require.Equal(t, nodeCfg.Storage, rewrittenTargetRepo.GetStorageName(), "stream director should have rewritten the storage name")
			}

			// We now inject errors into the streams by manually executing the error
			// handlers. This simulates that the RPCs have been proxied and may or may
			// not have returned an error.
			//
			// Note that these functions should not return an error: this is a strict
			// requirement because otherwise failures returned by the primary would
			// abort the operation on secondaries.
			require.Nil(t, streamParams.Primary().ErrHandler(tc.primaryErr))
			require.Nil(t, streamParams.Secondaries()[0].ErrHandler(tc.secondaryErr))

			// The request finalizer should then see the errors as injected above.
			require.Equal(t, tc.expectedErr, streamParams.RequestFinalizer())
		})
	}
}

func TestStreamDirector_maintenanceRPCs(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)

	primaryStorage := "internal-gitaly-0"
	primaryServer, primarySocketPath := runMockMaintenanceServer(
		t, testcfg.Build(t, testcfg.WithStorages(primaryStorage)),
	)

	secondaryStorage := "internal-gitaly-1"
	secondaryServer, secondarySocketPath := runMockMaintenanceServer(t, testcfg.Build(
		t, testcfg.WithStorages(secondaryStorage)),
	)

	cc, _, cleanup := RunPraefectServer(t, ctx, config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "default",
				Nodes: []*config.Node{
					{
						Storage: primaryStorage,
						Address: primarySocketPath,
					},
					{
						Storage: secondaryStorage,
						Address: secondarySocketPath,
					},
				},
			},
		},
	}, BuildOptions{})
	defer cleanup()

	repository := &gitalypb.Repository{
		StorageName:  "default",
		RelativePath: gittest.NewRepositoryName(t),
	}
	primaryRepository := &gitalypb.Repository{
		StorageName:  primaryStorage,
		RelativePath: repository.RelativePath,
	}
	secondaryRepository := &gitalypb.Repository{
		StorageName:  secondaryStorage,
		RelativePath: repository.RelativePath,
	}

	repositoryClient := gitalypb.NewRepositoryServiceClient(cc)
	refClient := gitalypb.NewRefServiceClient(cc)

	for _, tc := range []struct {
		desc                     string
		maintenanceFunc          func(t *testing.T)
		expectedPrimaryRequest   proto.Message
		expectedSecondaryRequest proto.Message
	}{
		{
			desc: "GarbageCollect",
			maintenanceFunc: func(t *testing.T) {
				//nolint:staticcheck
				_, err := repositoryClient.GarbageCollect(ctx, &gitalypb.GarbageCollectRequest{
					Repository:   repository,
					CreateBitmap: true,
					Prune:        true,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.GarbageCollectRequest{
				Repository:   primaryRepository,
				CreateBitmap: true,
				Prune:        true,
			},
			expectedSecondaryRequest: &gitalypb.GarbageCollectRequest{
				Repository:   secondaryRepository,
				CreateBitmap: true,
				Prune:        true,
			},
		},
		{
			desc: "RepackFull",
			maintenanceFunc: func(t *testing.T) {
				//nolint:staticcheck
				_, err := repositoryClient.RepackFull(ctx, &gitalypb.RepackFullRequest{
					Repository:   repository,
					CreateBitmap: true,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.RepackFullRequest{
				Repository:   primaryRepository,
				CreateBitmap: true,
			},
			expectedSecondaryRequest: &gitalypb.RepackFullRequest{
				Repository:   secondaryRepository,
				CreateBitmap: true,
			},
		},
		{
			desc: "RepackIncremental",
			maintenanceFunc: func(t *testing.T) {
				//nolint:staticcheck
				_, err := repositoryClient.RepackIncremental(ctx, &gitalypb.RepackIncrementalRequest{
					Repository: repository,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.RepackIncrementalRequest{
				Repository: primaryRepository,
			},
			expectedSecondaryRequest: &gitalypb.RepackIncrementalRequest{
				Repository: secondaryRepository,
			},
		},
		{
			desc: "Cleanup",
			maintenanceFunc: func(t *testing.T) {
				//nolint:staticcheck
				_, err := repositoryClient.Cleanup(ctx, &gitalypb.CleanupRequest{
					Repository: repository,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.CleanupRequest{
				Repository: primaryRepository,
			},
			expectedSecondaryRequest: &gitalypb.CleanupRequest{
				Repository: secondaryRepository,
			},
		},
		{
			desc: "WriteCommitGraph",
			maintenanceFunc: func(t *testing.T) {
				//nolint:staticcheck
				_, err := repositoryClient.WriteCommitGraph(ctx, &gitalypb.WriteCommitGraphRequest{
					Repository: repository,
					// This is not a valid split strategy, but we currently only support a
					// single default split strategy with value 0. So we just test with an
					// invalid split strategy to check that a non-default value gets properly
					// replicated.
					SplitStrategy: 1,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.WriteCommitGraphRequest{
				Repository:    primaryRepository,
				SplitStrategy: 1,
			},
			expectedSecondaryRequest: &gitalypb.WriteCommitGraphRequest{
				Repository:    secondaryRepository,
				SplitStrategy: 1,
			},
		},
		{
			desc: "MidxRepack",
			maintenanceFunc: func(t *testing.T) {
				//nolint:staticcheck
				_, err := repositoryClient.MidxRepack(ctx, &gitalypb.MidxRepackRequest{
					Repository: repository,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.MidxRepackRequest{
				Repository: primaryRepository,
			},
			expectedSecondaryRequest: &gitalypb.MidxRepackRequest{
				Repository: secondaryRepository,
			},
		},
		{
			desc: "OptimizeRepository",
			maintenanceFunc: func(t *testing.T) {
				_, err := repositoryClient.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
					Repository: repository,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.OptimizeRepositoryRequest{
				Repository: primaryRepository,
			},
			expectedSecondaryRequest: &gitalypb.OptimizeRepositoryRequest{
				Repository: secondaryRepository,
			},
		},
		{
			desc: "PruneUnreachableObjects",
			maintenanceFunc: func(t *testing.T) {
				_, err := repositoryClient.PruneUnreachableObjects(ctx, &gitalypb.PruneUnreachableObjectsRequest{
					Repository: repository,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.PruneUnreachableObjectsRequest{
				Repository: primaryRepository,
			},
			expectedSecondaryRequest: &gitalypb.PruneUnreachableObjectsRequest{
				Repository: secondaryRepository,
			},
		},
		{
			desc: "PackRefs",
			maintenanceFunc: func(t *testing.T) {
				//nolint:staticcheck
				_, err := refClient.PackRefs(ctx, &gitalypb.PackRefsRequest{
					Repository: repository,
				})
				require.NoError(t, err)
			},
			expectedPrimaryRequest: &gitalypb.PackRefsRequest{
				Repository: primaryRepository,
			},
			expectedSecondaryRequest: &gitalypb.PackRefsRequest{
				Repository: secondaryRepository,
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			tc.maintenanceFunc(t)
			testhelper.ProtoEqual(t, tc.expectedPrimaryRequest, <-primaryServer.requestCh)
			testhelper.ProtoEqual(t, tc.expectedSecondaryRequest, <-secondaryServer.requestCh)
		})
	}
}

type mockMaintenanceServer struct {
	requestCh chan proto.Message
	gitalypb.UnimplementedRepositoryServiceServer
	gitalypb.UnimplementedRefServiceServer
}

func runMockMaintenanceServer(t *testing.T, cfg gconfig.Cfg) (*mockMaintenanceServer, string) {
	server := &mockMaintenanceServer{
		requestCh: make(chan proto.Message, 1),
	}

	addr := testserver.RunGitalyServer(t, cfg, nil, func(srv *grpc.Server, deps *service.Dependencies) {
		gitalypb.RegisterRepositoryServiceServer(srv, server)
		gitalypb.RegisterRefServiceServer(srv, server)
	}, testserver.WithDisablePraefect())

	return server, addr
}

func (m *mockMaintenanceServer) GarbageCollect(ctx context.Context, in *gitalypb.GarbageCollectRequest) (*gitalypb.GarbageCollectResponse, error) {
	m.requestCh <- in
	return &gitalypb.GarbageCollectResponse{}, nil
}

func (m *mockMaintenanceServer) RepackFull(ctx context.Context, in *gitalypb.RepackFullRequest) (*gitalypb.RepackFullResponse, error) {
	m.requestCh <- in
	return &gitalypb.RepackFullResponse{}, nil
}

func (m *mockMaintenanceServer) RepackIncremental(ctx context.Context, in *gitalypb.RepackIncrementalRequest) (*gitalypb.RepackIncrementalResponse, error) {
	m.requestCh <- in
	return &gitalypb.RepackIncrementalResponse{}, nil
}

func (m *mockMaintenanceServer) Cleanup(ctx context.Context, in *gitalypb.CleanupRequest) (*gitalypb.CleanupResponse, error) {
	m.requestCh <- in
	return &gitalypb.CleanupResponse{}, nil
}

func (m *mockMaintenanceServer) WriteCommitGraph(ctx context.Context, in *gitalypb.WriteCommitGraphRequest) (*gitalypb.WriteCommitGraphResponse, error) {
	m.requestCh <- in
	return &gitalypb.WriteCommitGraphResponse{}, nil
}

func (m *mockMaintenanceServer) MidxRepack(ctx context.Context, in *gitalypb.MidxRepackRequest) (*gitalypb.MidxRepackResponse, error) {
	m.requestCh <- in
	return &gitalypb.MidxRepackResponse{}, nil
}

func (m *mockMaintenanceServer) OptimizeRepository(ctx context.Context, in *gitalypb.OptimizeRepositoryRequest) (*gitalypb.OptimizeRepositoryResponse, error) {
	m.requestCh <- in
	return &gitalypb.OptimizeRepositoryResponse{}, nil
}

func (m *mockMaintenanceServer) PruneUnreachableObjects(ctx context.Context, in *gitalypb.PruneUnreachableObjectsRequest) (*gitalypb.PruneUnreachableObjectsResponse, error) {
	m.requestCh <- in
	return &gitalypb.PruneUnreachableObjectsResponse{}, nil
}

func (m *mockMaintenanceServer) PackRefs(ctx context.Context, in *gitalypb.PackRefsRequest) (*gitalypb.PackRefsResponse, error) {
	m.requestCh <- in
	return &gitalypb.PackRefsResponse{}, nil
}

type mockRouter struct {
	Router
	routeRepositoryAccessorFunc func(ctx context.Context, virtualStorage, relativePath string, forcePrimary bool) (RepositoryAccessorRoute, error)
	routeRepositoryCreation     func(ctx context.Context, virtualStorage, relativePath, additionalRepoRelativePath string) (RepositoryMutatorRoute, error)
	routeRepositoryMutator      func(ctx context.Context, virtualStorage, relativePath, additionalRepoRelativePath string) (RepositoryMutatorRoute, error)
}

func (m mockRouter) RouteRepositoryAccessor(ctx context.Context, virtualStorage, relativePath string, forcePrimary bool) (RepositoryAccessorRoute, error) {
	return m.routeRepositoryAccessorFunc(ctx, virtualStorage, relativePath, forcePrimary)
}

func (m mockRouter) RouteRepositoryCreation(ctx context.Context, virtualStorage, relativePath, additionalRepoRelativePath string) (RepositoryMutatorRoute, error) {
	return m.routeRepositoryCreation(ctx, virtualStorage, relativePath, additionalRepoRelativePath)
}

func (m mockRouter) RouteRepositoryMutator(ctx context.Context, virtualStorage, relativePath, additionalRepoRelativePath string) (RepositoryMutatorRoute, error) {
	return m.routeRepositoryMutator(ctx, virtualStorage, relativePath, additionalRepoRelativePath)
}

func TestStreamDirectorAccessor(t *testing.T) {
	t.Parallel()
	gitalySocket := testhelper.GetTemporaryGitalySocketFileName(t)
	testhelper.NewServerWithHealth(t, gitalySocket)

	gitalyAddress := "unix://" + gitalySocket
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "praefect",
				Nodes: []*config.Node{
					{
						Address: gitalyAddress,
						Storage: "praefect-internal-1",
					},
				},
			},
		},
	}

	queue := datastore.NewPostgresReplicationEventQueue(testdb.New(t))

	targetRepo := gitalypb.Repository{
		StorageName:  "praefect",
		RelativePath: "/path/to/hashed/storage",
	}
	ctx := testhelper.Context(t)

	entry := testhelper.NewDiscardingLogEntry(t)
	rs := datastore.MockRepositoryStore{}

	nodeMgr, err := nodes.NewManager(entry, conf, nil, rs, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Minute)
	defer nodeMgr.Stop()

	txMgr := transactions.NewManager(conf)

	for _, tc := range []struct {
		desc   string
		router Router
		error  error
	}{
		{
			desc:   "success",
			router: NewNodeManagerRouter(nodeMgr, rs),
		},
		{
			desc: "repository not found",
			router: mockRouter{
				routeRepositoryAccessorFunc: func(_ context.Context, virtualStorage, relativePath string, _ bool) (RepositoryAccessorRoute, error) {
					return RepositoryAccessorRoute{}, commonerr.NewRepositoryNotFoundError(virtualStorage, relativePath)
				},
			},
			error: structerr.NewNotFound("%w", fmt.Errorf("accessor call: route repository accessor: %w", commonerr.NewRepositoryNotFoundError(targetRepo.StorageName, targetRepo.RelativePath))),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			coordinator := NewCoordinator(
				queue,
				rs,
				tc.router,
				txMgr,
				conf,
				protoregistry.GitalyProtoPreregistered,
			)

			frame, err := proto.Marshal(&gitalypb.FindAllBranchesRequest{Repository: &targetRepo})
			require.NoError(t, err)

			fullMethod := "/gitaly.RefService/FindAllBranches"

			peeker := &mockPeeker{frame: frame}
			streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
			if tc.error != nil {
				require.Equal(t, tc.error, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, gitalyAddress, streamParams.Primary().Conn.Target())

			mi, err := coordinator.registry.LookupMethod(fullMethod)
			require.NoError(t, err)
			require.Equal(t, protoregistry.ScopeRepository, mi.Scope, "method must be repository scoped")
			require.Equal(t, protoregistry.OpAccessor, mi.Operation, "method must be an accessor")

			m, err := mi.UnmarshalRequestProto(streamParams.Primary().Msg)
			require.NoError(t, err)

			rewrittenTargetRepo, err := mi.TargetRepo(m)
			require.NoError(t, err)
			require.Equal(t, "praefect-internal-1", rewrittenTargetRepo.GetStorageName(), "stream director should have rewritten the storage name")
		})
	}
}

func TestCoordinatorStreamDirector_distributesReads(t *testing.T) {
	t.Parallel()
	gitalySocket0, gitalySocket1 := testhelper.GetTemporaryGitalySocketFileName(t), testhelper.GetTemporaryGitalySocketFileName(t)
	primaryHealthSrv := testhelper.NewServerWithHealth(t, gitalySocket0)
	healthSrv := testhelper.NewServerWithHealth(t, gitalySocket1)

	primaryNodeConf := config.Node{
		Address: "unix://" + gitalySocket0,
		Storage: "gitaly-1",
	}

	secondaryNodeConf := config.Node{
		Address: "unix://" + gitalySocket1,
		Storage: "gitaly-2",
	}
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name:  "praefect",
				Nodes: []*config.Node{&primaryNodeConf, &secondaryNodeConf},
			},
		},
		Failover: config.Failover{
			Enabled:          true,
			ElectionStrategy: "local",
		},
	}

	queue := datastore.NewPostgresReplicationEventQueue(testdb.New(t))

	targetRepo := gitalypb.Repository{
		StorageName:  "praefect",
		RelativePath: "/path/to/hashed/storage",
	}
	ctx := testhelper.Context(t)

	entry := testhelper.NewDiscardingLogEntry(t)

	repoStore := datastore.MockRepositoryStore{
		GetConsistentStoragesFunc: func(ctx context.Context, virtualStorage, relativePath string) (string, map[string]struct{}, error) {
			return relativePath, map[string]struct{}{primaryNodeConf.Storage: {}, secondaryNodeConf.Storage: {}}, nil
		},
	}

	nodeMgr, err := nodes.NewManager(entry, conf, nil, repoStore, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Minute)
	defer nodeMgr.Stop()

	txMgr := transactions.NewManager(conf)

	coordinator := NewCoordinator(
		queue,
		repoStore,
		NewNodeManagerRouter(nodeMgr, repoStore),
		txMgr,
		conf,
		protoregistry.GitalyProtoPreregistered,
	)

	t.Run("forwards accessor operations", func(t *testing.T) {
		var primaryChosen int
		var secondaryChosen int

		for i := 0; i < 16; i++ {
			frame, err := proto.Marshal(&gitalypb.FindAllBranchesRequest{Repository: &targetRepo})
			require.NoError(t, err)

			fullMethod := "/gitaly.RefService/FindAllBranches"

			peeker := &mockPeeker{frame: frame}

			streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
			require.NoError(t, err)
			require.Contains(t, []string{primaryNodeConf.Address, secondaryNodeConf.Address}, streamParams.Primary().Conn.Target(), "must be redirected to primary or secondary")

			var nodeConf config.Node
			switch streamParams.Primary().Conn.Target() {
			case primaryNodeConf.Address:
				nodeConf = primaryNodeConf
				primaryChosen++
			case secondaryNodeConf.Address:
				nodeConf = secondaryNodeConf
				secondaryChosen++
			}

			mi, err := coordinator.registry.LookupMethod(fullMethod)
			require.NoError(t, err)
			require.Equal(t, protoregistry.OpAccessor, mi.Operation, "method must be an accessor")

			m, err := protoMessage(mi, streamParams.Primary().Msg)
			require.NoError(t, err)

			rewrittenTargetRepo, err := mi.TargetRepo(m)
			require.NoError(t, err)
			require.Equal(t, nodeConf.Storage, rewrittenTargetRepo.GetStorageName(), "stream director must rewrite the storage name")
		}

		require.NotZero(t, primaryChosen, "primary should have been chosen at least once")
		require.NotZero(t, secondaryChosen, "secondary should have been chosen at least once")
	})

	t.Run("forwards accessor to primary if force-routing", func(t *testing.T) {
		var primaryChosen int
		var secondaryChosen int

		for i := 0; i < 16; i++ {
			frame, err := proto.Marshal(&gitalypb.FindAllBranchesRequest{Repository: &targetRepo})
			require.NoError(t, err)

			fullMethod := "/gitaly.RefService/FindAllBranches"

			peeker := &mockPeeker{frame: frame}

			ctx := correlation.ContextWithCorrelation(ctx, "my-correlation-id")
			ctx = testhelper.MergeIncomingMetadata(ctx, metadata.Pairs(routeRepositoryAccessorPolicy, routeRepositoryAccessorPolicyPrimaryOnly))

			streamParams, err := coordinator.StreamDirector(ctx, fullMethod, peeker)
			require.NoError(t, err)
			require.Contains(t, []string{primaryNodeConf.Address, secondaryNodeConf.Address}, streamParams.Primary().Conn.Target(), "must be redirected to primary or secondary")

			var nodeConf config.Node
			switch streamParams.Primary().Conn.Target() {
			case primaryNodeConf.Address:
				nodeConf = primaryNodeConf
				primaryChosen++
			case secondaryNodeConf.Address:
				nodeConf = secondaryNodeConf
				secondaryChosen++
			}

			mi, err := coordinator.registry.LookupMethod(fullMethod)
			require.NoError(t, err)
			require.Equal(t, protoregistry.OpAccessor, mi.Operation, "method must be an accessor")

			m, err := protoMessage(mi, streamParams.Primary().Msg)
			require.NoError(t, err)

			rewrittenTargetRepo, err := mi.TargetRepo(m)
			require.NoError(t, err)
			require.Equal(t, nodeConf.Storage, rewrittenTargetRepo.GetStorageName(), "stream director must rewrite the storage name")
		}

		require.Equal(t, 16, primaryChosen, "primary should have always been chosen")
		require.Zero(t, secondaryChosen, "secondary should never have been chosen")
	})

	t.Run("forwards accessor to primary for primary-only RPCs", func(t *testing.T) {
		var primaryChosen int
		var secondaryChosen int

		for i := 0; i < 16; i++ {
			frame, err := proto.Marshal(&gitalypb.GetObjectDirectorySizeRequest{Repository: &targetRepo})
			require.NoError(t, err)

			fullMethod := "/gitaly.RepositoryService/GetObjectDirectorySize"

			peeker := &mockPeeker{frame: frame}
			ctx := testhelper.Context(t)

			streamParams, err := coordinator.StreamDirector(ctx, fullMethod, peeker)
			require.NoError(t, err)
			require.Contains(t, []string{primaryNodeConf.Address, secondaryNodeConf.Address}, streamParams.Primary().Conn.Target(), "must be redirected to primary or secondary")

			var nodeConf config.Node
			switch streamParams.Primary().Conn.Target() {
			case primaryNodeConf.Address:
				nodeConf = primaryNodeConf
				primaryChosen++
			case secondaryNodeConf.Address:
				nodeConf = secondaryNodeConf
				secondaryChosen++
			}

			mi, err := coordinator.registry.LookupMethod(fullMethod)
			require.NoError(t, err)
			require.Equal(t, protoregistry.OpAccessor, mi.Operation, "method must be an accessor")

			m, err := protoMessage(mi, streamParams.Primary().Msg)
			require.NoError(t, err)

			rewrittenTargetRepo, err := mi.TargetRepo(m)
			require.NoError(t, err)
			require.Equal(t, nodeConf.Storage, rewrittenTargetRepo.GetStorageName(), "stream director must rewrite the storage name")
		}

		require.Equal(t, 16, primaryChosen, "primary should have always been chosen")
		require.Zero(t, secondaryChosen, "secondary should never have been chosen")
	})

	t.Run("forwards accessor operations only to healthy nodes", func(t *testing.T) {
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		shard, err := nodeMgr.GetShard(ctx, conf.VirtualStorages[0].Name)
		require.NoError(t, err)

		gitaly1, err := shard.GetNode(secondaryNodeConf.Storage)
		require.NoError(t, err)
		waitNodeToChangeHealthStatus(t, ctx, gitaly1, false)
		defer func() {
			healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
			waitNodeToChangeHealthStatus(t, ctx, gitaly1, true)
		}()

		frame, err := proto.Marshal(&gitalypb.FindAllBranchesRequest{Repository: &targetRepo})
		require.NoError(t, err)

		fullMethod := "/gitaly.RefService/FindAllBranches"

		peeker := &mockPeeker{frame: frame}
		streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
		require.NoError(t, err)
		require.Equal(t, primaryNodeConf.Address, streamParams.Primary().Conn.Target(), "must be redirected to primary")

		mi, err := coordinator.registry.LookupMethod(fullMethod)
		require.NoError(t, err)
		require.Equal(t, protoregistry.OpAccessor, mi.Operation, "method must be an accessor")

		m, err := protoMessage(mi, streamParams.Primary().Msg)
		require.NoError(t, err)

		rewrittenTargetRepo, err := mi.TargetRepo(m)
		require.NoError(t, err)
		require.Equal(t, "gitaly-1", rewrittenTargetRepo.GetStorageName(), "stream director must rewrite the storage name")
	})

	t.Run("fails if force-routing to unhealthy primary", func(t *testing.T) {
		primaryHealthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		shard, err := nodeMgr.GetShard(ctx, conf.VirtualStorages[0].Name)
		require.NoError(t, err)

		primaryGitaly, err := shard.GetNode(primaryNodeConf.Storage)
		require.NoError(t, err)
		waitNodeToChangeHealthStatus(t, ctx, primaryGitaly, false)
		defer func() {
			primaryHealthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
			waitNodeToChangeHealthStatus(t, ctx, primaryGitaly, true)
		}()

		frame, err := proto.Marshal(&gitalypb.FindAllBranchesRequest{Repository: &targetRepo})
		require.NoError(t, err)

		fullMethod := "/gitaly.RefService/FindAllBranches"

		ctx := correlation.ContextWithCorrelation(ctx, "my-correlation-id")
		ctx = testhelper.MergeIncomingMetadata(ctx, metadata.Pairs(routeRepositoryAccessorPolicy, routeRepositoryAccessorPolicyPrimaryOnly))

		peeker := &mockPeeker{frame: frame}
		_, err = coordinator.StreamDirector(ctx, fullMethod, peeker)
		require.True(t, errors.Is(err, nodes.ErrPrimaryNotHealthy))
	})

	t.Run("doesn't forward mutator operations", func(t *testing.T) {
		frame, err := proto.Marshal(&gitalypb.UserUpdateBranchRequest{Repository: &targetRepo})
		require.NoError(t, err)

		fullMethod := "/gitaly.OperationService/UserUpdateBranch"

		peeker := &mockPeeker{frame: frame}
		streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
		require.NoError(t, err)
		require.Equal(t, primaryNodeConf.Address, streamParams.Primary().Conn.Target(), "must be redirected to primary")

		mi, err := coordinator.registry.LookupMethod(fullMethod)
		require.NoError(t, err)
		require.Equal(t, protoregistry.OpMutator, mi.Operation, "method must be a mutator")

		m, err := protoMessage(mi, streamParams.Primary().Msg)
		require.NoError(t, err)

		rewrittenTargetRepo, err := mi.TargetRepo(m)
		require.NoError(t, err)
		require.Equal(t, "gitaly-1", rewrittenTargetRepo.GetStorageName(), "stream director must rewrite the storage name")
	})
}

func TestRewrittenRepositoryMessage(t *testing.T) {
	buildRequest := func(storageName, relativePath, additionalRelativePath string) *gitalypb.CreateObjectPoolRequest {
		return &gitalypb.CreateObjectPoolRequest{
			ObjectPool: &gitalypb.ObjectPool{
				Repository: &gitalypb.Repository{
					StorageName:  storageName,
					RelativePath: relativePath,
				},
			},
			Origin: &gitalypb.Repository{
				StorageName:  storageName,
				RelativePath: additionalRelativePath,
			},
		}
	}

	originalRequest := buildRequest("original-storage", "original-relative-path", "original-additional-relative-path")

	methodInfo, err := protoregistry.GitalyProtoPreregistered.LookupMethod("/gitaly.ObjectPoolService/CreateObjectPool")
	require.NoError(t, err)

	rewrittenMessageBytes, err := rewrittenRepositoryMessage(methodInfo, originalRequest, "rewritten-storage", "rewritten-relative-path", "rewritten-additional-relative-path")
	require.NoError(t, err)

	var rewrittenMessage gitalypb.CreateObjectPoolRequest
	require.NoError(t, proto.Unmarshal(rewrittenMessageBytes, &rewrittenMessage))

	testhelper.ProtoEqual(t, buildRequest("original-storage", "original-relative-path", "original-additional-relative-path"), originalRequest)
	testhelper.ProtoEqual(t, buildRequest("rewritten-storage", "rewritten-relative-path", "rewritten-additional-relative-path"), &rewrittenMessage)
}

func TestStreamDirector_repo_creation(t *testing.T) {
	t.Parallel()

	db := testdb.New(t)

	for i, tc := range []struct {
		desc              string
		replicationFactor int
		primaryStored     bool
		assignmentsStored bool
	}{
		{
			desc:              "without variable replication factor",
			primaryStored:     true,
			assignmentsStored: false,
		},
		{
			desc:              "with variable replication factor",
			replicationFactor: 3,
			primaryStored:     true,
			assignmentsStored: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			tx := db.Begin(t)
			defer tx.Rollback(t)

			primaryNode := &config.Node{Storage: "praefect-internal-1"}
			healthySecondaryNode := &config.Node{Storage: "praefect-internal-2"}
			unhealthySecondaryNode := &config.Node{Storage: "praefect-internal-3"}
			conf := config.Config{
				Failover: config.Failover{ElectionStrategy: config.ElectionStrategyPerRepository},
				VirtualStorages: []*config.VirtualStorage{
					{
						Name:                     "praefect",
						DefaultReplicationFactor: tc.replicationFactor,
						Nodes:                    []*config.Node{primaryNode, healthySecondaryNode, unhealthySecondaryNode},
					},
				},
			}

			rewrittenStorage := primaryNode.Storage
			targetRepo := gitalypb.Repository{
				StorageName:  "praefect",
				RelativePath: "/path/to/hashed/storage",
			}

			conns := Connections{
				"praefect": {
					primaryNode.Storage:            &grpc.ClientConn{},
					healthySecondaryNode.Storage:   &grpc.ClientConn{},
					unhealthySecondaryNode.Storage: &grpc.ClientConn{},
				},
			}

			primaryConnPointer := fmt.Sprintf("%p", conns["praefect"][primaryNode.Storage])
			secondaryConnPointers := []string{fmt.Sprintf("%p", conns["praefect"][healthySecondaryNode.Storage])}
			rs := datastore.NewPostgresRepositoryStore(tx, conf.StorageNames())
			router := NewPerRepositoryRouter(
				conns,
				nil,
				StaticHealthChecker{"praefect": {primaryNode.Storage, healthySecondaryNode.Storage}},
				mockRandom{
					intnFunc: func(n int) int {
						require.Equal(t, n, 2, "number of primary candidates should match the number of healthy nodes")
						return 0
					},
					shuffleFunc: func(n int, swap func(int, int)) {
						require.Equal(t, n, 2, "number of secondary candidates should match the number of node minus the primary")
					},
				},
				nil,
				nil,
				rs,
				conf.DefaultReplicationFactors(),
			)

			txMgr := transactions.NewManager(conf)
			queueInterceptor := datastore.NewReplicationEventQueueInterceptor(datastore.NewPostgresReplicationEventQueue(tx))

			coordinator := NewCoordinator(
				queueInterceptor,
				rs,
				router,
				txMgr,
				conf,
				protoregistry.GitalyProtoPreregistered,
			)

			frame, err := proto.Marshal(&gitalypb.CreateRepositoryRequest{
				Repository: &targetRepo,
			})
			require.NoError(t, err)

			fullMethod := "/gitaly.RepositoryService/CreateRepository"
			ctx := testhelper.Context(t)

			peeker := &mockPeeker{frame}
			streamParams, err := coordinator.StreamDirector(correlation.ContextWithCorrelation(ctx, "my-correlation-id"), fullMethod, peeker)
			require.NoError(t, err)
			require.Equal(t, primaryConnPointer, fmt.Sprintf("%p", streamParams.Primary().Conn))

			var secondaries []string
			for _, dst := range streamParams.Secondaries() {
				secondaries = append(secondaries, fmt.Sprintf("%p", dst.Conn))
			}
			require.Equal(t, secondaryConnPointers, secondaries, "secondary connections did not match expected")

			mi, err := coordinator.registry.LookupMethod(fullMethod)
			require.NoError(t, err)

			m, err := mi.UnmarshalRequestProto(streamParams.Primary().Msg)
			require.NoError(t, err)

			rewrittenTargetRepo, err := mi.TargetRepo(m)
			require.NoError(t, err)
			require.Equal(t, rewrittenStorage, rewrittenTargetRepo.GetStorageName(), "stream director should have rewritten the storage name")

			vote := voting.VoteFromData([]byte{})
			require.NoError(t, txMgr.VoteTransaction(ctx, 1, "praefect-internal-1", vote))
			require.NoError(t, txMgr.VoteTransaction(ctx, 1, "praefect-internal-2", vote))

			// this call creates new events in the queue and simulates usual flow of the update operation
			err = streamParams.RequestFinalizer()
			require.NoError(t, err)

			// wait until event persisted (async operation)
			require.NoError(t, queueInterceptor.Wait(time.Minute, func(i *datastore.ReplicationEventQueueInterceptor) bool {
				return len(i.GetEnqueuedResult()) == 1
			}))

			var expectedEvents, actualEvents []datastore.ReplicationEvent
			for _, target := range []string{unhealthySecondaryNode.Storage} {
				actual, err := queueInterceptor.Dequeue(ctx, "praefect", target, 10)
				require.NoError(t, err)
				require.Len(t, actual, 1)

				actualEvents = append(actualEvents, actual[0])
				expectedEvents = append(expectedEvents, datastore.ReplicationEvent{
					ID:        actual[0].ID,
					State:     datastore.JobStateInProgress,
					Attempt:   2,
					LockID:    fmt.Sprintf("praefect|%s|/path/to/hashed/storage", target),
					CreatedAt: actual[0].CreatedAt,
					UpdatedAt: actual[0].UpdatedAt,
					Job: datastore.ReplicationJob{
						RepositoryID:      int64(i + 1),
						Change:            datastore.UpdateRepo,
						VirtualStorage:    conf.VirtualStorages[0].Name,
						RelativePath:      targetRepo.RelativePath,
						TargetNodeStorage: target,
						SourceNodeStorage: primaryNode.Storage,
					},
					Meta: datastore.Params{metadatahandler.CorrelationIDKey: "my-correlation-id"},
				})
			}

			require.Equal(t, expectedEvents, actualEvents, "ensure replication job created by stream director is correct")
		})
	}
}

func waitNodeToChangeHealthStatus(t *testing.T, ctx context.Context, node nodes.Node, health bool) {
	t.Helper()

	require.Eventually(t, func() bool {
		if node.IsHealthy() == health {
			return true
		}

		_, err := node.CheckHealth(ctx)
		require.NoError(t, err)
		return false
	}, time.Minute, time.Nanosecond)
}

type mockPeeker struct {
	frame []byte
}

func (m *mockPeeker) Peek() ([]byte, error) {
	return m.frame, nil
}

func (m *mockPeeker) Modify(payload []byte) error {
	m.frame = payload

	return nil
}

func TestAbsentCorrelationID(t *testing.T) {
	t.Parallel()
	gitalySocket0, gitalySocket1 := testhelper.GetTemporaryGitalySocketFileName(t), testhelper.GetTemporaryGitalySocketFileName(t)
	healthSrv0 := testhelper.NewServerWithHealth(t, gitalySocket0)
	healthSrv1 := testhelper.NewServerWithHealth(t, gitalySocket1)
	healthSrv0.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthSrv1.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	primaryAddress, secondaryAddress := "unix://"+gitalySocket0, "unix://"+gitalySocket1
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "praefect",
				Nodes: []*config.Node{
					{
						Address: primaryAddress,
						Storage: "praefect-internal-1",
					},
					{
						Address: secondaryAddress,
						Storage: "praefect-internal-2",
					},
				},
			},
		},
	}

	queueInterceptor := datastore.NewReplicationEventQueueInterceptor(datastore.NewPostgresReplicationEventQueue(testdb.New(t)))
	queueInterceptor.OnEnqueue(func(ctx context.Context, event datastore.ReplicationEvent, queue datastore.ReplicationEventQueue) (datastore.ReplicationEvent, error) {
		assert.True(t, len(queueInterceptor.GetEnqueued()) < 2, "expected only one event to be created")
		return queue.Enqueue(ctx, event)
	})
	targetRepo := gitalypb.Repository{
		StorageName:  "praefect",
		RelativePath: "/path/to/hashed/storage",
	}
	ctx := testhelper.Context(t)

	entry := testhelper.NewDiscardingLogEntry(t)

	nodeMgr, err := nodes.NewManager(entry, conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Hour)
	defer nodeMgr.Stop()

	txMgr := transactions.NewManager(conf)
	rs := datastore.MockRepositoryStore{}

	coordinator := NewCoordinator(
		queueInterceptor,
		rs,
		NewNodeManagerRouter(nodeMgr, rs),
		txMgr,
		conf,
		protoregistry.GitalyProtoPreregistered,
	)

	frame, err := proto.Marshal(&gitalypb.CreateObjectPoolRequest{
		Origin:     &targetRepo,
		ObjectPool: &gitalypb.ObjectPool{Repository: &targetRepo},
	})
	require.NoError(t, err)

	fullMethod := "/gitaly.ObjectPoolService/CreateObjectPool"
	peeker := &mockPeeker{frame}
	streamParams, err := coordinator.StreamDirector(ctx, fullMethod, peeker)
	require.NoError(t, err)
	require.Equal(t, primaryAddress, streamParams.Primary().Conn.Target())

	// must be run as it adds replication events to the queue
	require.NoError(t, streamParams.RequestFinalizer())

	require.NoError(t, queueInterceptor.Wait(time.Minute, func(i *datastore.ReplicationEventQueueInterceptor) bool {
		return len(i.GetEnqueuedResult()) == 1
	}))
	jobs, err := queueInterceptor.Dequeue(ctx, conf.VirtualStorages[0].Name, conf.VirtualStorages[0].Nodes[1].Storage, 1)
	require.NoError(t, err)
	require.Len(t, jobs, 1)

	require.NotZero(t, jobs[0].Meta[metadatahandler.CorrelationIDKey],
		"the coordinator should have generated a random ID")
}

func TestCoordinatorEnqueueFailure(t *testing.T) {
	t.Parallel()
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "praefect",
				Nodes: []*config.Node{
					{
						Address: "unix:///woof",
						Storage: "praefect-internal-1",
					},
					{
						Address: "unix:///meow",
						Storage: "praefect-internal-2",
					},
				},
			},
		},
	}

	queueInterceptor := datastore.NewReplicationEventQueueInterceptor(nil)
	errQ := make(chan error, 1)
	queueInterceptor.OnEnqueue(func(ctx context.Context, event datastore.ReplicationEvent, queue datastore.ReplicationEventQueue) (datastore.ReplicationEvent, error) {
		return datastore.ReplicationEvent{}, <-errQ
	})
	queueInterceptor.OnDequeue(func(context.Context, string, string, int, datastore.ReplicationEventQueue) ([]datastore.ReplicationEvent, error) {
		return nil, nil
	})
	queueInterceptor.OnAcknowledge(func(context.Context, datastore.JobState, []uint64, datastore.ReplicationEventQueue) ([]uint64, error) {
		return nil, nil
	})

	registrar := func(srv *grpc.Server) {
		gitalypb.RegisterRepositoryServiceServer(srv, &mockRepositoryService{
			ReplicateRepositoryFunc: func(context.Context, *gitalypb.ReplicateRepositoryRequest) (*gitalypb.ReplicateRepositoryResponse, error) {
				return &gitalypb.ReplicateRepositoryResponse{}, nil
			},
		})
	}

	ctx := testhelper.Context(t)

	cc, _, cleanup := RunPraefectServer(t, ctx, conf, BuildOptions{
		WithAnnotations: protoregistry.GitalyProtoPreregistered,
		WithQueue:       queueInterceptor,
		WithBackends: WithMockBackends(t, map[string]func(*grpc.Server){
			conf.VirtualStorages[0].Nodes[0].Storage: registrar,
			conf.VirtualStorages[0].Nodes[1].Storage: registrar,
		}),
	})
	defer cleanup()

	mcli := gitalypb.NewRepositoryServiceClient(cc)

	errQ <- nil
	repoReq := &gitalypb.ReplicateRepositoryRequest{
		Repository: &gitalypb.Repository{
			RelativePath: "meow",
			StorageName:  conf.VirtualStorages[0].Name,
		},
	}
	_, err := mcli.ReplicateRepository(ctx, repoReq)
	require.NoError(t, err)

	expectErrMsg := "enqueue failed"
	errQ <- errors.New(expectErrMsg)
	_, err = mcli.ReplicateRepository(ctx, repoReq)
	require.Error(t, err)
	require.Equal(t, err.Error(), "rpc error: code = Internal desc = enqueue replication event: "+expectErrMsg)
}

func TestStreamDirectorStorageScope(t *testing.T) {
	// stubs health-check requests because nodes.NewManager establishes connection on creation
	gitalySocket0, gitalySocket1 := testhelper.GetTemporaryGitalySocketFileName(t), testhelper.GetTemporaryGitalySocketFileName(t)
	testhelper.NewServerWithHealth(t, gitalySocket0)
	testhelper.NewServerWithHealth(t, gitalySocket1)

	primaryAddress, secondaryAddress := "unix://"+gitalySocket0, "unix://"+gitalySocket1
	primaryGitaly := &config.Node{Address: primaryAddress, Storage: "gitaly-1"}
	secondaryGitaly := &config.Node{Address: secondaryAddress, Storage: "gitaly-2"}
	conf := config.Config{
		Failover: config.Failover{Enabled: true},
		VirtualStorages: []*config.VirtualStorage{{
			Name:  "praefect",
			Nodes: []*config.Node{primaryGitaly, secondaryGitaly},
		}},
	}

	rs := datastore.MockRepositoryStore{}

	nodeMgr, err := nodes.NewManager(testhelper.NewDiscardingLogEntry(t), conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil, nil, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Second)
	defer nodeMgr.Stop()
	coordinator := NewCoordinator(
		nil,
		rs,
		NewNodeManagerRouter(nodeMgr, rs),
		nil,
		conf,
		protoregistry.GitalyProtoPreregistered,
	)
	ctx := testhelper.Context(t)

	t.Run("mutator", func(t *testing.T) {
		fullMethod := "/gitaly.NamespaceService/RemoveNamespace"
		requireScopeOperation(t, coordinator.registry, fullMethod, protoregistry.ScopeStorage, protoregistry.OpMutator)

		frame, err := proto.Marshal(&gitalypb.RemoveNamespaceRequest{
			StorageName: conf.VirtualStorages[0].Name,
			Name:        "stub",
		})
		require.NoError(t, err)

		streamParams, err := coordinator.StreamDirector(ctx, fullMethod, &mockPeeker{frame})
		require.NoError(t, err)

		require.Equal(t, primaryAddress, streamParams.Primary().Conn.Target(), "stream director didn't redirect to gitaly storage")

		rewritten := gitalypb.RemoveNamespaceRequest{}
		require.NoError(t, proto.Unmarshal(streamParams.Primary().Msg, &rewritten))
		require.Equal(t, primaryGitaly.Storage, rewritten.StorageName, "stream director didn't rewrite storage")
	})

	t.Run("accessor", func(t *testing.T) {
		fullMethod := "/gitaly.NamespaceService/NamespaceExists"
		requireScopeOperation(t, coordinator.registry, fullMethod, protoregistry.ScopeStorage, protoregistry.OpAccessor)

		frame, err := proto.Marshal(&gitalypb.NamespaceExistsRequest{
			StorageName: conf.VirtualStorages[0].Name,
			Name:        "stub",
		})
		require.NoError(t, err)

		streamParams, err := coordinator.StreamDirector(ctx, fullMethod, &mockPeeker{frame})
		require.NoError(t, err)

		require.Equal(t, primaryAddress, streamParams.Primary().Conn.Target(), "stream director didn't redirect to gitaly storage")

		rewritten := gitalypb.RemoveNamespaceRequest{}
		require.NoError(t, proto.Unmarshal(streamParams.Primary().Msg, &rewritten))
		require.Equal(t, primaryGitaly.Storage, rewritten.StorageName, "stream director didn't rewrite storage")
	})
}

func TestStreamDirectorStorageScopeError(t *testing.T) {
	ctx := testhelper.Context(t)

	t.Run("no storage provided", func(t *testing.T) {
		mgr := &nodes.MockManager{
			GetShardFunc: func(s string) (nodes.Shard, error) {
				require.FailNow(t, "validation of input was not executed")
				return nodes.Shard{}, assert.AnError
			},
		}

		rs := datastore.MockRepositoryStore{}
		coordinator := NewCoordinator(
			nil,
			rs,
			NewNodeManagerRouter(mgr, rs),
			nil,
			config.Config{},
			protoregistry.GitalyProtoPreregistered,
		)

		frame, err := proto.Marshal(&gitalypb.RemoveNamespaceRequest{StorageName: "", Name: "stub"})
		require.NoError(t, err)

		_, err = coordinator.StreamDirector(ctx, "/gitaly.NamespaceService/RemoveNamespace", &mockPeeker{frame})
		require.Error(t, err)
		result, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, result.Code())
		require.Equal(t, "storage scoped: target storage is invalid", result.Message())
	})

	t.Run("unknown storage provided", func(t *testing.T) {
		mgr := &nodes.MockManager{
			GetShardFunc: func(s string) (nodes.Shard, error) {
				require.Equal(t, "fake", s)
				return nodes.Shard{}, nodes.ErrVirtualStorageNotExist
			},
		}

		rs := datastore.MockRepositoryStore{}
		coordinator := NewCoordinator(
			nil,
			rs,
			NewNodeManagerRouter(mgr, rs),
			nil,
			config.Config{},
			protoregistry.GitalyProtoPreregistered,
		)

		frame, err := proto.Marshal(&gitalypb.RemoveNamespaceRequest{StorageName: "fake", Name: "stub"})
		require.NoError(t, err)

		_, err = coordinator.StreamDirector(ctx, "/gitaly.NamespaceService/RemoveNamespace", &mockPeeker{frame})
		require.Error(t, err)
		result, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, result.Code())
		require.Equal(t, "virtual storage does not exist", result.Message())
	})

	t.Run("primary gitaly is not healthy", func(t *testing.T) {
		t.Run("accessor", func(t *testing.T) {
			mgr := &nodes.MockManager{
				GetShardFunc: func(s string) (nodes.Shard, error) {
					require.Equal(t, "fake", s)
					return nodes.Shard{}, nodes.ErrPrimaryNotHealthy
				},
			}

			rs := datastore.MockRepositoryStore{}
			coordinator := NewCoordinator(
				nil,
				rs,
				NewNodeManagerRouter(mgr, rs),
				nil,
				config.Config{},
				protoregistry.GitalyProtoPreregistered,
			)

			fullMethod := "/gitaly.NamespaceService/NamespaceExists"
			requireScopeOperation(t, coordinator.registry, fullMethod, protoregistry.ScopeStorage, protoregistry.OpAccessor)

			frame, err := proto.Marshal(&gitalypb.NamespaceExistsRequest{StorageName: "fake", Name: "stub"})
			require.NoError(t, err)

			_, err = coordinator.StreamDirector(ctx, fullMethod, &mockPeeker{frame})
			require.Error(t, err)
			result, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, codes.Internal, result.Code())
			require.Equal(t, `accessor storage scoped: route storage accessor "fake": primary gitaly is not healthy`, result.Message())
		})

		t.Run("mutator", func(t *testing.T) {
			mgr := &nodes.MockManager{
				GetShardFunc: func(s string) (nodes.Shard, error) {
					require.Equal(t, "fake", s)
					return nodes.Shard{}, nodes.ErrPrimaryNotHealthy
				},
			}
			rs := datastore.MockRepositoryStore{}
			coordinator := NewCoordinator(
				nil,
				rs,
				NewNodeManagerRouter(mgr, rs),
				nil,
				config.Config{},
				protoregistry.GitalyProtoPreregistered,
			)

			fullMethod := "/gitaly.NamespaceService/RemoveNamespace"
			requireScopeOperation(t, coordinator.registry, fullMethod, protoregistry.ScopeStorage, protoregistry.OpMutator)

			frame, err := proto.Marshal(&gitalypb.RemoveNamespaceRequest{StorageName: "fake", Name: "stub"})
			require.NoError(t, err)

			_, err = coordinator.StreamDirector(ctx, fullMethod, &mockPeeker{frame})
			require.Error(t, err)
			result, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, codes.Internal, result.Code())
			require.Equal(t, `mutator storage scoped: get shard "fake": primary gitaly is not healthy`, result.Message())
		})
	})
}

func TestDisabledTransactionsWithFeatureFlag(t *testing.T) {
	ctx := testhelper.Context(t)

	for rpc, enabledFn := range transactionRPCs {
		if enabledFn(ctx) {
			require.True(t, shouldUseTransaction(ctx, rpc))
			break
		}
	}
}

func requireScopeOperation(t *testing.T, registry *protoregistry.Registry, fullMethod string, scope protoregistry.Scope, op protoregistry.OpType) {
	t.Helper()

	mi, err := registry.LookupMethod(fullMethod)
	require.NoError(t, err)
	require.Equal(t, scope, mi.Scope, "scope doesn't match requested")
	require.Equal(t, op, mi.Operation, "operation type doesn't match requested")
}

type mockOperationServer struct {
	gitalypb.UnimplementedOperationServiceServer
	t      testing.TB
	wg     *sync.WaitGroup
	err    error
	called bool
}

func (s *mockOperationServer) UserCreateBranch(
	context.Context,
	*gitalypb.UserCreateBranchRequest,
) (*gitalypb.UserCreateBranchResponse, error) {
	// We need to wait for all servers to arrive in this RPC. If we don't it could be that for
	// example the primary arrives quicker than the others and directly errors. This would cause
	// stream cancellation, and if the secondaries didn't yet end up in UserCreateBranch, we
	// wouldn't see the function call.
	s.called = true
	s.wg.Done()
	s.wg.Wait()
	return &gitalypb.UserCreateBranchResponse{}, s.err
}

type mockLeaseEnder struct{}

func (e mockLeaseEnder) EndLease(context.Context) error {
	return nil
}

type mockDiskCache struct {
	cache.Cache
}

func (c *mockDiskCache) StartLease(*gitalypb.Repository) (cache.LeaseEnder, error) {
	return mockLeaseEnder{}, nil
}

// TestCoordinator_grpcErrorHandling asserts that we correctly proxy errors in case any of the nodes
// fails. Most importantly, we want to make sure to only ever forward errors from the primary and
// never from the secondaries.
func TestCoordinator_grpcErrorHandling(t *testing.T) {
	t.Parallel()
	praefectConfig := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: testhelper.DefaultStorageName,
			},
		},
	}

	type gitalyNode struct {
		mock            *nodes.MockNode
		operationServer *mockOperationServer
	}

	_, repoProto, _ := testcfg.BuildWithRepo(t)

	for _, tc := range []struct {
		desc        string
		errByNode   map[string]error
		expectedErr error
	}{
		{
			desc: "no errors",
		},
		{
			desc: "primary error gets forwarded",
			errByNode: map[string]error{
				"primary": errors.New("foo"),
			},
			expectedErr: status.Error(codes.Internal, "foo"),
		},
		{
			desc: "secondary error gets ignored",
			errByNode: map[string]error{
				"secondary-1": errors.New("foo"),
			},
		},
		{
			desc: "primary error has precedence",
			errByNode: map[string]error{
				"primary":     errors.New("primary"),
				"secondary-1": errors.New("secondary-1"),
				"secondary-2": errors.New("secondary-2"),
			},
			expectedErr: status.Error(codes.Internal, "primary"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := testhelper.Context(t)

			var wg sync.WaitGroup
			gitalies := make(map[string]gitalyNode)
			for _, gitaly := range []string{"primary", "secondary-1", "secondary-2"} {
				gitaly := gitaly

				cfg := testcfg.Build(t, testcfg.WithStorages(gitaly))

				operationServer := &mockOperationServer{
					t:  t,
					wg: &wg,
				}

				addr := testserver.RunGitalyServer(t, cfg, nil, func(srv *grpc.Server, deps *service.Dependencies) {
					gitalypb.RegisterOperationServiceServer(srv, operationServer)
				}, testserver.WithDiskCache(&mockDiskCache{}), testserver.WithDisablePraefect())

				conn, err := client.DialContext(ctx, addr, []grpc.DialOption{
					grpc.WithDefaultCallOptions(grpc.ForceCodec(proxy.NewCodec())),
				})
				require.NoError(t, err)
				defer conn.Close()

				gitalies[gitaly] = gitalyNode{
					mock: &nodes.MockNode{
						Conn:             conn,
						Healthy:          true,
						GetStorageMethod: func() string { return gitaly },
					},
					operationServer: operationServer,
				}

				praefectConfig.VirtualStorages[0].Nodes = append(praefectConfig.VirtualStorages[0].Nodes, &config.Node{
					Address: addr,
					Storage: gitaly,
				})
			}

			praefectConn, _, cleanup := RunPraefectServer(t, ctx, praefectConfig, BuildOptions{
				// Set up a mock manager which sets up primary/secondaries and pretends that all nodes are
				// healthy. We need fixed roles and unhealthy nodes will not take part in transactions.
				WithNodeMgr: &nodes.MockManager{
					Storage: testhelper.DefaultStorageName,
					GetShardFunc: func(shardName string) (nodes.Shard, error) {
						require.Equal(t, testhelper.DefaultStorageName, shardName)
						return nodes.Shard{
							Primary: gitalies["primary"].mock,
							Secondaries: []nodes.Node{
								gitalies["secondary-1"].mock,
								gitalies["secondary-2"].mock,
							},
						}, nil
					},
				},
				// Set up a mock repsoitory store pretending that all nodes are consistent. Only consistent
				// nodes will take part in transactions.
				WithRepoStore: datastore.MockRepositoryStore{
					GetReplicaPathFunc: func(ctx context.Context, repositoryID int64) (string, error) {
						return repoProto.GetRelativePath(), nil
					},
					GetConsistentStoragesFunc: func(ctx context.Context, virtualStorage, relativePath string) (string, map[string]struct{}, error) {
						return relativePath, map[string]struct{}{"primary": {}, "secondary-1": {}, "secondary-2": {}}, nil
					},
				},
			})
			defer cleanup()

			for name, node := range gitalies {
				wg.Add(1)
				node.operationServer.err = tc.errByNode[name]
				node.operationServer.called = false
			}

			_, err := gitalypb.NewOperationServiceClient(praefectConn).UserCreateBranch(ctx,
				&gitalypb.UserCreateBranchRequest{
					Repository: repoProto,
				})
			testhelper.RequireGrpcError(t, tc.expectedErr, err)

			for _, node := range gitalies {
				require.True(t, node.operationServer.called, "expected gitaly %q to have been called", node.mock.GetStorage())
			}
		})
	}
}

type mockTransaction struct {
	nodeStates      map[string]transactions.VoteResult
	subtransactions int
	didVote         map[string]bool
}

func (t mockTransaction) ID() uint64 {
	return 0
}

func (t mockTransaction) CountSubtransactions() int {
	return t.subtransactions
}

func (t mockTransaction) DidVote(node string) bool {
	return t.didVote[node]
}

func (t mockTransaction) State() (map[string]transactions.VoteResult, error) {
	return t.nodeStates, nil
}

func TestGetUpdatedAndOutdatedSecondaries(t *testing.T) {
	type node struct {
		name  string
		state transactions.VoteResult
		err   error
	}
	ctx := testhelper.Context(t)

	anyErr := errors.New("arbitrary error")
	grpcErr := status.Error(codes.Internal, "arbitrary gRPC error")

	for _, tc := range []struct {
		desc                   string
		primary                node
		secondaries            []node
		replicas               []string
		subtransactions        int
		didVote                map[string]bool
		expectedPrimaryDirtied bool
		expectedOutdated       []string
		expectedUpdated        []string
		expectedMetrics        map[string]int
	}{
		{
			desc: "single committed node",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
		},
		{
			desc: "single failed node",
			primary: node{
				name:  "primary",
				state: transactions.VoteFailed,
			},
			subtransactions: 1,
		},
		{
			desc: "single node with standard error",
			primary: node{
				name: "primary",
				err:  anyErr,
			},
		},
		{
			desc: "single node with gRPC error",
			primary: node{
				name: "primary",
				err:  grpcErr,
			},
		},
		{
			desc: "single node without subtransactions",
			primary: node{
				name: "primary",
			},
			subtransactions:        0,
			expectedPrimaryDirtied: true,
		},
		{
			desc: "single successful node with replica",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
			},
			replicas: []string{"replica"},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"replica"},
			expectedMetrics: map[string]int{
				"outdated": 1,
			},
		},
		{
			desc: "single failing node with replica is not considered modified",
			primary: node{
				name:  "primary",
				state: transactions.VoteFailed,
			},
			subtransactions: 1,
		},
		{
			desc: "single node with standard error with replica",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   anyErr,
			},
			replicas: []string{"replica"},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"replica"},
			expectedMetrics: map[string]int{
				"outdated": 1,
			},
		},
		{
			desc: "single node with gRPC error with replica",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   grpcErr,
			},
			replicas: []string{"replica"},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"replica"},
			expectedMetrics: map[string]int{
				"outdated": 1,
			},
		},
		{
			desc: "single node with standard error without commit with replica",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   anyErr,
			},
			replicas:               []string{"replica"},
			subtransactions:        1,
			expectedPrimaryDirtied: false,
		},
		{
			desc: "single node with gRPC error without commit with replica",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   grpcErr,
			},
			replicas:               []string{"replica"},
			subtransactions:        1,
			expectedPrimaryDirtied: false,
		},
		{
			desc: "single node without transaction with replica",
			primary: node{
				name: "primary",
			},
			replicas:               []string{"replica"},
			subtransactions:        0,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"replica"},
			expectedMetrics: map[string]int{
				"outdated": 1,
			},
		},
		{
			desc: "multiple committed nodes",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedUpdated:        []string{"s1", "s2"},
			expectedMetrics: map[string]int{
				"updated": 2,
			},
		},
		{
			desc: "multiple committed nodes with primary err",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   anyErr,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1", "s2"},
			expectedMetrics: map[string]int{
				"node-error-status": 2,
			},
		},
		{
			desc: "multiple committed nodes with primary gRPC err",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   grpcErr,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1", "s2"},
			expectedMetrics: map[string]int{
				"node-error-status": 2,
			},
		},
		{
			desc: "multiple committed nodes with same standard error as primary",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   anyErr,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted, err: anyErr},
				{name: "s2", state: transactions.VoteCommitted, err: anyErr},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1", "s2"},
			expectedMetrics: map[string]int{
				"node-error-status": 2,
			},
		},
		{
			desc: "multiple committed nodes with same gRPC error as primary",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   grpcErr,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted, err: grpcErr},
				{name: "s2", state: transactions.VoteCommitted, err: grpcErr},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedUpdated:        []string{"s1", "s2"},
			expectedMetrics: map[string]int{
				"updated": 2,
			},
		},
		{
			desc: "committed node with same generated error code and message",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   status.Error(codes.Internal, "some error"),
			},
			secondaries: []node{
				{
					name:  "s1",
					state: transactions.VoteCommitted,
					err:   status.Error(codes.Internal, "some error"),
				},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedUpdated:        []string{"s1"},
			expectedMetrics: map[string]int{
				"updated": 1,
			},
		},
		{
			desc: "committed node with different generated error code",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   status.Error(codes.Internal, "some error"),
			},
			secondaries: []node{
				{
					name:  "s1",
					state: transactions.VoteCommitted,
					err:   status.Error(codes.FailedPrecondition, "some error"),
				},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1"},
			expectedMetrics: map[string]int{
				"node-error-status": 1,
			},
		},
		{
			desc: "committed node with different generated error message",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   status.Error(codes.Internal, "some error"),
			},
			secondaries: []node{
				{
					name:  "s1",
					state: transactions.VoteCommitted,
					err:   status.Error(codes.Internal, "different error"),
				},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1"},
			expectedMetrics: map[string]int{
				"node-error-status": 1,
			},
		},
		{
			desc: "multiple committed nodes with different error as primary",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   grpcErr,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted, err: status.Error(codes.Internal, "somethingsomething")},
				{name: "s2", state: transactions.VoteCommitted, err: grpcErr},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedUpdated:        []string{"s2"},
			expectedOutdated:       []string{"s1"},
			expectedMetrics: map[string]int{
				"node-error-status": 1,
				"updated":           1,
			},
		},
		{
			desc: "multiple committed nodes with secondary err",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted, err: grpcErr},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedUpdated:        []string{"s2"},
			expectedOutdated:       []string{"s1"},
			expectedMetrics: map[string]int{
				"node-error-status": 1,
				"updated":           1,
			},
		},
		{
			desc: "multiple committed nodes with primary and missing secondary err",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
				err:   grpcErr,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteCommitted, err: grpcErr},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedUpdated:        []string{"s1"},
			expectedOutdated:       []string{"s2"},
			expectedMetrics: map[string]int{
				"node-error-status": 1,
				"updated":           1,
			},
		},
		{
			desc: "partial success",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteFailed},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedUpdated:        []string{"s2"},
			expectedOutdated:       []string{"s1"},
			expectedMetrics: map[string]int{
				"node-not-committed": 1,
				"updated":            1,
			},
		},
		{
			desc: "failure with (impossible) secondary success",
			primary: node{
				name:  "primary",
				state: transactions.VoteFailed,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteFailed},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1", "s2"},
			expectedMetrics: map[string]int{
				"primary-not-committed": 2,
			},
		},
		{
			desc: "failure with no primary votes",
			primary: node{
				name:  "primary",
				state: transactions.VoteFailed,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteFailed},
				{name: "s2", state: transactions.VoteCommitted},
			},
			didVote: map[string]bool{
				"s1": true,
				"s2": true,
			},
			subtransactions: 1,
		},
		{
			desc: "multiple nodes without subtransactions",
			primary: node{
				name:  "primary",
				state: transactions.VoteFailed,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteFailed},
				{name: "s2", state: transactions.VoteCommitted},
			},
			subtransactions:        0,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1", "s2"},
			expectedMetrics: map[string]int{
				"no-votes": 2,
			},
		},
		{
			desc: "multiple nodes with replica and partial failures",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteFailed},
				{name: "s2", state: transactions.VoteCommitted},
			},
			replicas: []string{"r1", "r2"},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1", "r1", "r2"},
			expectedUpdated:        []string{"s2"},
			expectedMetrics: map[string]int{
				"node-not-committed": 1,
				"outdated":           2,
				"updated":            1,
			},
		},
		{
			desc: "multiple nodes with replica and partial err",
			primary: node{
				name:  "primary",
				state: transactions.VoteCommitted,
			},
			secondaries: []node{
				{name: "s1", state: transactions.VoteFailed},
				{name: "s2", state: transactions.VoteCommitted, err: grpcErr},
			},
			replicas: []string{"r1", "r2"},
			didVote: map[string]bool{
				"primary": true,
			},
			subtransactions:        1,
			expectedPrimaryDirtied: true,
			expectedOutdated:       []string{"s1", "s2", "r1", "r2"},
			expectedMetrics: map[string]int{
				"node-error-status":  1,
				"node-not-committed": 1,
				"outdated":           2,
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			nodes := append(tc.secondaries, tc.primary)
			voters := make([]transactions.Voter, len(nodes))

			states := make(map[string]transactions.VoteResult)
			nodeErrors := &nodeErrors{
				errByNode: make(map[string]error),
			}

			for i, node := range nodes {
				voters[i] = transactions.Voter{
					Name:  node.name,
					Votes: 1,
				}
				states[node.name] = node.state
				nodeErrors.errByNode[node.name] = node.err
			}

			transaction := mockTransaction{
				nodeStates:      states,
				subtransactions: tc.subtransactions,
				didVote:         tc.didVote,
			}

			route := RepositoryMutatorRoute{
				Primary: RouterNode{
					Storage: tc.primary.name,
				},
			}
			for _, secondary := range tc.secondaries {
				route.Secondaries = append(route.Secondaries, RouterNode{
					Storage: secondary.name,
				})
			}
			route.ReplicationTargets = append(route.ReplicationTargets, tc.replicas...)

			metric := prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "stub", Help: "help",
			}, []string{"reason"})

			primaryDirtied, updated, outdated := getUpdatedAndOutdatedSecondaries(ctx, route, transaction, nodeErrors, metric)
			require.Equal(t, tc.expectedPrimaryDirtied, primaryDirtied)
			require.ElementsMatch(t, tc.expectedUpdated, updated)
			require.ElementsMatch(t, tc.expectedOutdated, outdated)

			expectedMetrics := "# HELP stub help\n# TYPE stub counter\n"
			for metric, value := range tc.expectedMetrics {
				expectedMetrics += fmt.Sprintf("stub{reason=\"%s\"} %d\n", metric, value)
			}

			require.NoError(t, testutil.CollectAndCompare(metric, strings.NewReader(expectedMetrics)))
		})
	}
}

func TestNewRequestFinalizer_contextIsDisjointedFromTheRPC(t *testing.T) {
	type ctxKey struct{}

	parentDeadline := time.Now()
	ctx := testhelper.Context(t)

	//nolint:forbidigo // We explicitly want to test that the deadline does not propagate into
	// the request's context.
	ctx, cancel := context.WithDeadline(context.WithValue(ctx, ctxKey{}, "value"), parentDeadline)
	defer cancel()

	requireSuppressedCancellation := func(tb testing.TB, ctx context.Context) {
		deadline, ok := ctx.Deadline()
		require.True(tb, ok)
		require.NotEqual(tb, parentDeadline, deadline)
		require.Equal(tb, ctx.Value(ctxKey{}), "value")
		require.Nil(tb, ctx.Err())
		select {
		case <-ctx.Done():
			tb.Fatal("context should not be canceled if the parent is canceled")
		default:
			require.NotNil(tb, ctx.Done())
		}
	}

	err := errors.New("error")

	for _, tc := range []struct {
		desc         string
		change       datastore.ChangeType
		errMsg       string
		enqueueError error
	}{
		{
			desc:   "increment generation receives suppressed cancellation",
			change: datastore.UpdateRepo,
			errMsg: "increment generation: error",
		},
		{
			desc:   "rename repository receives suppressed cancellation",
			change: datastore.RenameRepo,
			errMsg: "rename repository: error",
		},
		{
			desc:         "enqueue receives suppressed cancellation",
			errMsg:       "enqueue replication event: error",
			enqueueError: err,
		},
		{
			desc:   "original context error is returned",
			errMsg: context.DeadlineExceeded.Error(),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.EqualError(t,
				NewCoordinator(
					&datastore.MockReplicationEventQueue{
						EnqueueFunc: func(ctx context.Context, _ datastore.ReplicationEvent) (datastore.ReplicationEvent, error) {
							requireSuppressedCancellation(t, ctx)
							return datastore.ReplicationEvent{}, tc.enqueueError
						},
					},
					datastore.MockRepositoryStore{
						IncrementGenerationFunc: func(ctx context.Context, _ int64, _ string, _ []string) error {
							requireSuppressedCancellation(t, ctx)
							return err
						},
						RenameRepositoryFunc: func(ctx context.Context, _, _, _, _ string) error {
							requireSuppressedCancellation(t, ctx)
							return err
						},
						CreateRepositoryFunc: func(ctx context.Context, _ int64, _, _, _, _ string, _, _ []string, _, _ bool) error {
							requireSuppressedCancellation(t, ctx)
							return err
						},
					},
					nil,
					nil,
					config.Config{},
					nil,
				).newRequestFinalizer(
					ctx,
					0,
					"virtual storage",
					&gitalypb.Repository{},
					"replica-path",
					"primary",
					[]string{},
					[]string{"secondary"},
					tc.change,
					datastore.Params{"RelativePath": "relative-path"},
					"rpc-name",
				)(),
				tc.errMsg,
			)
		})
	}
}

func TestStreamParametersContext(t *testing.T) {
	// Because we're using NewFeatureFlag, they'll end up in the All array.
	enabledFF := featureflag.NewFeatureFlag("default_enabled", "", "", true)
	disabledFF := featureflag.NewFeatureFlag("default_disabled", "", "", false)

	type expectedFlag struct {
		flag    featureflag.FeatureFlag
		enabled bool
	}

	expectedFlags := func(overrides ...expectedFlag) []expectedFlag {
		flagValues := map[featureflag.FeatureFlag]bool{}
		for _, flag := range featureflag.DefinedFlags() {
			flagValues[flag] = flag.OnByDefault
		}
		for _, override := range overrides {
			flagValues[override.flag] = override.enabled
		}

		expectedFlags := make([]expectedFlag, 0, len(flagValues))
		for flag, value := range flagValues {
			expectedFlags = append(expectedFlags, expectedFlag{
				flag: flag, enabled: value,
			})
		}

		return expectedFlags
	}

	metadataForFlags := func(flags []expectedFlag) metadata.MD {
		pairs := []string{}
		for _, flag := range flags {
			pairs = append(pairs, flag.flag.MetadataKey(), strconv.FormatBool(flag.enabled))
		}
		return metadata.Pairs(pairs...)
	}

	//nolint:forbidigo // We explicitly test context values, so we cannot use the testhelper
	// context here given that it would contain unrelated data and thus change the system under
	// test.
	for _, tc := range []struct {
		desc               string
		setupContext       func() context.Context
		expectedIncomingMD metadata.MD
		expectedOutgoingMD metadata.MD
		expectedFlags      []expectedFlag
	}{
		{
			desc: "no metadata",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedFlags:      expectedFlags(),
			expectedOutgoingMD: metadataForFlags(expectedFlags()),
		},
		{
			desc: "with incoming metadata",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("key", "value"))
				return ctx
			},
			expectedIncomingMD: metadata.Pairs("key", "value"),
			expectedOutgoingMD: metadata.Join(
				metadata.Pairs("key", "value"),
				metadataForFlags(expectedFlags()),
			),
			expectedFlags: expectedFlags(),
		},
		{
			desc: "with outgoing metadata",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("key", "value"))
				return ctx
			},
			expectedOutgoingMD: metadata.Join(
				metadata.Pairs("key", "value"),
				metadataForFlags(expectedFlags()),
			),
			expectedFlags: expectedFlags(),
		},
		{
			desc: "with incoming and outgoing metadata",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("incoming", "value"))
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("outgoing", "value"))
				return ctx
			},
			// This behaviour is quite subtle: in the previous test case where we only
			// have outgoing metadata, we retain it. But in case we have both incoming
			// and outgoing we'd discard the outgoing metadata altogether. It is
			// debatable whether this is a bug or feature, so I'll just document this
			// weird edge case here for now.
			expectedIncomingMD: metadata.Pairs("incoming", "value"),
			expectedOutgoingMD: metadata.Join(
				metadata.Pairs("incoming", "value"),
				metadataForFlags(expectedFlags()),
			),
			expectedFlags: expectedFlags(),
		},
		{
			desc: "with flags set to their default values",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = featureflag.IncomingCtxWithFeatureFlag(ctx, enabledFF, true)
				ctx = featureflag.IncomingCtxWithFeatureFlag(ctx, disabledFF, false)
				return ctx
			},
			expectedIncomingMD: metadata.Pairs(
				enabledFF.MetadataKey(), "true",
				disabledFF.MetadataKey(), "false",
			),
			expectedOutgoingMD: metadata.Join(
				metadataForFlags(expectedFlags()),
			),
			expectedFlags: expectedFlags(),
		},
		{
			desc: "with flags set to their reverse default values",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = featureflag.IncomingCtxWithFeatureFlag(ctx, enabledFF, false)
				ctx = featureflag.IncomingCtxWithFeatureFlag(ctx, disabledFF, true)
				return ctx
			},
			expectedIncomingMD: metadata.Pairs(
				enabledFF.MetadataKey(), "false",
				disabledFF.MetadataKey(), "true",
			),
			expectedOutgoingMD: metadata.Join(
				metadataForFlags(expectedFlags(
					expectedFlag{flag: enabledFF, enabled: false},
					expectedFlag{flag: disabledFF, enabled: true},
				)),
			),
			expectedFlags: expectedFlags(
				expectedFlag{flag: enabledFF, enabled: false},
				expectedFlag{flag: disabledFF, enabled: true},
			),
		},
		{
			desc: "mixed flags and metadata",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
					disabledFF.MetadataKey(), "true",
					"incoming", "value"),
				)
				return ctx
			},
			expectedIncomingMD: metadata.Pairs(
				disabledFF.MetadataKey(), "true",
				"incoming", "value",
			),
			expectedOutgoingMD: metadata.Join(
				metadataForFlags(expectedFlags(
					expectedFlag{flag: disabledFF, enabled: true},
				)),
				metadata.Pairs(
					"incoming", "value",
				),
			),
			expectedFlags: expectedFlags(
				expectedFlag{flag: disabledFF, enabled: true},
			),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := streamParametersContext(tc.setupContext())

			incomingMD, ok := metadata.FromIncomingContext(ctx)
			if tc.expectedIncomingMD == nil {
				require.False(t, ok)
			} else {
				require.True(t, ok)
			}
			require.Equal(t, tc.expectedIncomingMD, incomingMD)

			outgoingMD, ok := metadata.FromOutgoingContext(ctx)
			if tc.expectedOutgoingMD == nil {
				require.False(t, ok)
			} else {
				require.True(t, ok)
			}
			require.Equal(t, tc.expectedOutgoingMD, outgoingMD)

			incomingCtx := gitaly_metadata.OutgoingToIncoming(ctx)
			for _, expectedFlag := range tc.expectedFlags {
				require.Equal(t, expectedFlag.enabled, expectedFlag.flag.IsEnabled(incomingCtx))
			}
		})
	}
}
