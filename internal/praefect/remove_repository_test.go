package praefect

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service/setup"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/internal/grpc/protoregistry"
	"gitlab.com/gitlab-org/gitaly/v16/internal/grpc/proxy"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testdb"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRemoveRepositoryHandler(t *testing.T) {
	t.Parallel()
	ctx := testhelper.Context(t)

	errServedByGitaly := structerr.NewInternal("request passed to Gitaly")
	const virtualStorage, relativePath = "virtual-storage", "relative-path"

	db := testdb.New(t)
	for _, tc := range []struct {
		desc          string
		routeToGitaly bool
		repository    *gitalypb.Repository
		repoDeleted   bool
		error         error
	}{
		{
			desc:  "missing repository",
			error: structerr.NewInvalidArgument("%w", storage.ErrRepositoryNotSet),
		},
		{
			desc:       "repository not found",
			repository: &gitalypb.Repository{StorageName: "virtual-storage", RelativePath: "doesn't exist"},
			error:      structerr.NewNotFound("repository does not exist"),
		},
		{
			desc:        "repository found",
			repository:  &gitalypb.Repository{StorageName: "virtual-storage", RelativePath: relativePath},
			repoDeleted: true,
		},
		{
			desc:          "routed to gitaly",
			routeToGitaly: true,
			error:         errServedByGitaly,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			db.TruncateAll(t)

			const gitaly1Storage = "gitaly-1"
			gitaly1Cfg := testcfg.Build(t, testcfg.WithStorages(gitaly1Storage))
			gitaly1RepoPath := filepath.Join(gitaly1Cfg.Storages[0].Path, relativePath)
			gitaly1Addr := testserver.RunGitalyServer(t, gitaly1Cfg, setup.RegisterAll, testserver.WithDisablePraefect())

			const gitaly2Storage = "gitaly-2"
			gitaly2Cfg := testcfg.Build(t, testcfg.WithStorages(gitaly2Storage))
			gitaly2RepoPath := filepath.Join(gitaly2Cfg.Storages[0].Path, relativePath)
			gitaly2Addr := testserver.RunGitalyServer(t, gitaly2Cfg, setup.RegisterAll, testserver.WithDisablePraefect())

			cfg := config.Config{VirtualStorages: []*config.VirtualStorage{
				{
					Name: virtualStorage,
					Nodes: []*config.Node{
						{Storage: gitaly1Storage, Address: gitaly1Addr},
						{Storage: gitaly2Storage, Address: gitaly2Addr},
					},
				},
			}}

			for _, repoPath := range []string{gitaly1RepoPath, gitaly2RepoPath} {
				gittest.Exec(t, gitaly1Cfg, "init", "--bare", repoPath)
			}

			rs := datastore.NewPostgresRepositoryStore(db, cfg.StorageNames())

			require.NoError(t, rs.CreateRepository(ctx, 0, virtualStorage, relativePath, relativePath, gitaly1Storage, []string{gitaly2Storage, "non-existent-storage"}, nil, false, false))

			tmp := testhelper.TempDir(t)

			ln, err := net.Listen("unix", filepath.Join(tmp, "praefect"))
			require.NoError(t, err)

			electionStrategy := config.ElectionStrategyPerRepository
			if tc.routeToGitaly {
				electionStrategy = config.ElectionStrategySQL
			}

			nodeSet, err := DialNodes(ctx, cfg.VirtualStorages, nil, nil, nil, nil, testhelper.SharedLogger(t))
			require.NoError(t, err)
			defer nodeSet.Close()

			srv := NewGRPCServer(&Dependencies{
				Config: config.Config{Failover: config.Failover{ElectionStrategy: electionStrategy}},
				Logger: testhelper.SharedLogger(t),
				Director: func(ctx context.Context, fullMethodName string, peeker proxy.StreamPeeker) (*proxy.StreamParameters, error) {
					return nil, errServedByGitaly
				},
				RepositoryStore: rs,
				Registry:        protoregistry.GitalyProtoPreregistered,
				Conns:           nodeSet.Connections(),
			}, nil)
			defer srv.Stop()

			go testhelper.MustServe(t, srv, ln)

			clientConn, err := grpc.DialContext(ctx, "unix:"+ln.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer clientConn.Close()

			client := gitalypb.NewRepositoryServiceClient(clientConn)
			_, err = client.RepositorySize(ctx, &gitalypb.RepositorySizeRequest{Repository: tc.repository})
			testhelper.RequireGrpcError(t, errServedByGitaly, err)

			assertExistence := require.DirExists
			if tc.repoDeleted {
				assertExistence = require.NoDirExists
			}

			resp, err := client.RemoveRepository(ctx, &gitalypb.RemoveRepositoryRequest{Repository: tc.repository})
			if tc.error != nil {
				testhelper.RequireGrpcError(t, tc.error, err)
				assertExistence(t, gitaly1RepoPath)
				assertExistence(t, gitaly2RepoPath)
				return
			}

			require.NoError(t, err)
			testhelper.ProtoEqual(t, &gitalypb.RemoveRepositoryResponse{}, resp)
			assertExistence(t, gitaly1RepoPath)
			assertExistence(t, gitaly2RepoPath)
		})
	}
}
