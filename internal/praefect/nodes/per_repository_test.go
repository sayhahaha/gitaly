package nodes

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testdb"
)

func TestPerRepositoryElector(t *testing.T) {
	t.Parallel()
	ctx := testhelper.Context(t)

	type storageRecord struct {
		generation int
		assigned   bool
	}

	type state map[string]map[string]map[string]storageRecord

	type matcher func(tb testing.TB, primary string)
	any := func(expected ...string) matcher {
		return func(tb testing.TB, primary string) {
			tb.Helper()
			require.Contains(tb, expected, primary)
		}
	}

	noPrimary := func() matcher {
		return func(tb testing.TB, primary string) {
			tb.Helper()
			require.Empty(tb, primary)
		}
	}

	type steps []struct {
		healthyNodes   map[string][]string
		error          error
		primary        matcher
		noBlockedQuery bool
	}

	db := testdb.New(t)

	for _, tc := range []struct {
		desc         string
		state        state
		steps        steps
		existingJobs []datastore.ReplicationEvent
	}{
		{
			desc: "elects the most up to date storage",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 1},
						"gitaly-2": {generation: 0},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-1"),
				},
			},
		},
		{
			desc: "does not elect healthy outdated replicas",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 1},
						"gitaly-2": {generation: 0},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					error:          ErrNoPrimary,
					primary:        noPrimary(),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "no valid primary",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 0},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					error:          ErrNoPrimary,
					primary:        noPrimary(),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "random healthy node on the latest generation",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 0},
						"gitaly-2": {generation: 0},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-1", "gitaly-2"),
				},
			},
		},
		{
			desc: "fails over to up to date healthy note",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 1},
						"gitaly-2": {generation: 1},
						"gitaly-3": {generation: 0},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-3"},
					},
					primary: any("gitaly-1"),
				},
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-2"),
				},
			},
		},
		{
			desc: "does not fail over to healthy outdated nodes",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 1},
						"gitaly-3": {generation: 0},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-1"),
				},
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					primary:        any("gitaly-1"),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "fails over to assigned nodes when assignments are set",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 2, assigned: true},
						"gitaly-2": {generation: 2, assigned: true},
						"gitaly-3": {generation: 2, assigned: false},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-3"},
					},
					primary: any("gitaly-1"),
				},
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-2"),
				},
			},
		},
		{
			desc: "fails over to unassigned replicas if no valid assigned primaries exist",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 2, assigned: true},
						"gitaly-2": {generation: 2, assigned: false},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-1"),
				},
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-2"),
				},
			},
		},
		{
			desc: "fails over to up to date assigned replica from healthy unassigned",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 2, assigned: true},
						"gitaly-2": {generation: 2, assigned: false},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-2"),
				},
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-1"),
				},
			},
		},
		{
			desc: "doesnt fail over to outdated assigned replica from healthy unassigned",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 1, assigned: true},
						"gitaly-2": {generation: 2, assigned: false},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-2"),
				},
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-2", "gitaly-3"},
					},
					primary:        any("gitaly-2"),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "does not demote the primary when there are no valid candidates",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 1, assigned: true},
						"gitaly-2": {generation: 0, assigned: false},
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1", "gitaly-2", "gitaly-3"},
					},
					primary: any("gitaly-1"),
				},
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-2", "gitaly-3"},
					},
					primary:        any("gitaly-1"),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "doesnt elect replicas with delete_replica in ready state",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 0, assigned: true},
					},
				},
			},
			existingJobs: []datastore.ReplicationEvent{
				{
					State: datastore.JobStateReady,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "virtual-storage-1",
						RelativePath:      "relative-path-1",
						TargetNodeStorage: "gitaly-1",
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1"},
					},
					error:          ErrNoPrimary,
					primary:        noPrimary(),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "doesnt elect replicas with delete_replica in in_progress state",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 0, assigned: true},
					},
				},
			},
			existingJobs: []datastore.ReplicationEvent{
				{
					State: datastore.JobStateInProgress,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "virtual-storage-1",
						RelativePath:      "relative-path-1",
						TargetNodeStorage: "gitaly-1",
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1"},
					},
					error:          ErrNoPrimary,
					primary:        noPrimary(),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "doesnt elect replicas with delete_replica in failed state",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 0, assigned: true},
					},
				},
			},
			existingJobs: []datastore.ReplicationEvent{
				{
					State: datastore.JobStateFailed,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "virtual-storage-1",
						RelativePath:      "relative-path-1",
						TargetNodeStorage: "gitaly-1",
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1"},
					},
					error:          ErrNoPrimary,
					primary:        noPrimary(),
					noBlockedQuery: true,
				},
			},
		},
		{
			desc: "irrelevant delete_replica jobs are ignored",
			state: state{
				"virtual-storage-1": {
					"relative-path-1": {
						"gitaly-1": {generation: 0, assigned: true},
					},
				},
			},
			existingJobs: []datastore.ReplicationEvent{
				{
					State: datastore.JobStateReady,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "wrong-virtual-storage",
						RelativePath:      "relative-path-1",
						TargetNodeStorage: "gitaly-1",
					},
				},
				{
					State: datastore.JobStateReady,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "virtual-storage-1",
						RelativePath:      "wrong-relative-path",
						TargetNodeStorage: "gitaly-1",
					},
				},
				{
					State: datastore.JobStateReady,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "virtual-storage-1",
						RelativePath:      "relative-path-1",
						TargetNodeStorage: "wrong-storage",
					},
				},
				{
					State: datastore.JobStateDead,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "virtual-storage-1",
						RelativePath:      "relative-path-1",
						TargetNodeStorage: "gitaly-1",
					},
				},
				{
					State: datastore.JobStateCompleted,
					Job: datastore.ReplicationJob{
						Change:            datastore.DeleteReplica,
						VirtualStorage:    "virtual-storage-1",
						RelativePath:      "relative-path-1",
						TargetNodeStorage: "gitaly-1",
					},
				},
			},
			steps: steps{
				{
					healthyNodes: map[string][]string{
						"virtual-storage-1": {"gitaly-1"},
					},
					primary: any("gitaly-1"),
				},
			},
		},
		{
			desc: "repository does not exist",
			steps: steps{
				{
					error:          datastore.ErrRepositoryNotFound,
					primary:        noPrimary(),
					noBlockedQuery: true,
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			db.TruncateAll(t)

			rs := datastore.NewPostgresRepositoryStore(db, nil)
			for virtualStorage, relativePaths := range tc.state {
				for relativePath, storages := range relativePaths {
					repositoryID, err := rs.ReserveRepositoryID(ctx, virtualStorage, relativePath)
					require.NoError(t, err)

					repoCreated := false
					for storage, record := range storages {
						if !repoCreated {
							repoCreated = true
							require.NoError(t, rs.CreateRepository(ctx, repositoryID, virtualStorage, relativePath, relativePath, storage, nil, nil, false, false))
						}

						require.NoError(t, rs.SetGeneration(ctx, repositoryID, storage, relativePath, record.generation))

						if record.assigned {
							_, err := db.ExecContext(ctx, `
								INSERT INTO repository_assignments VALUES ($1, $2, $3, $4)
							`, virtualStorage, relativePath, storage, repositoryID)
							require.NoError(t, err)
						}
					}
				}
			}

			for _, event := range tc.existingJobs {
				repositoryID, err := rs.GetRepositoryID(ctx, event.Job.VirtualStorage, event.Job.RelativePath)
				if err != nil {
					require.Equal(t, datastore.ErrRepositoryNotFound, err)
				}

				event.Job.RepositoryID = repositoryID

				_, err = db.ExecContext(ctx,
					"INSERT INTO replication_queue (state, job) VALUES ($1, $2)",
					event.State, event.Job,
				)
				require.NoError(t, err)
			}

			previousPrimary := ""
			const repositoryID int64 = 1

			for _, step := range tc.steps {
				runElection := func(tx *testdb.TxWrapper) (string, *logrus.Entry) {
					logger := testhelper.NewLogger(t)
					hook := testhelper.AddLoggerHook(logger)
					elector := NewPerRepositoryElector(tx)

					primary, err := elector.GetPrimary(logger.ToContext(ctx), "", repositoryID)
					assert.Equal(t, step.error, err)
					assert.Less(t, len(hook.AllEntries()), 2)

					var entry *logrus.Entry
					if len(hook.AllEntries()) == 1 {
						entry = hook.AllEntries()[0]
					}

					return primary, entry
				}

				// There is a 10s race from setting the healthy nodes to the transaction beginning. If creating
				// the transaction takes longer, none of the nodes will be considered healthy.
				testdb.SetHealthyNodes(t, ctx, db, map[string]map[string][]string{"praefect-0": step.healthyNodes})

				// Run every step with two concurrent transactions to ensure two Praefect's running
				// election at the same time do not elect the primary multiple times. We begin both
				// transactions at the same time to ensure they have the same snapshot of the
				// database. The second transaction would be blocked until the first transaction commits.
				// To verify concurrent election runs do not elect the primary multiple times, we assert
				// the second transaction performed no changes and the primary is what the first run elected
				// it to be.
				txFirst := db.Begin(t)
				defer txFirst.Rollback(t)

				txSecond := db.Begin(t)
				defer txSecond.Rollback(t)

				primary, logEntry := runElection(txFirst)
				step.primary(t, primary)

				if previousPrimary != primary {
					require.NotNil(t, logEntry)
					require.Equal(t, "primary node changed", logEntry.Message)
					require.Equal(t, log.Fields{
						"repository_id":    repositoryID,
						"current_primary":  primary,
						"previous_primary": previousPrimary,
					}, logEntry.Data)
				} else {
					require.Nil(t, logEntry)
				}

				// Run the second election on the same database snapshot. This should result in no changes.
				// Running this prior to the first transaction committing would block.

				var (
					secondPrimary  string
					secondLogEntry *logrus.Entry
				)

				secondStmtDone := make(chan struct{})
				go func() {
					defer close(secondStmtDone)
					secondPrimary, secondLogEntry = runElection(txSecond)
				}()

				// With read-committed isolation mode, it's not sufficient to start the transaction to ensure concurrent
				// execution different statements within a single transacation can see commits done during the transaction.
				// Below we wait for the statement to actually begin executing to ensure the test actually exercises concurrent
				// execution.
				if !step.noBlockedQuery {
					testdb.WaitForBlockedQuery(t, ctx, db, "WITH reread AS (")
				}

				txFirst.Commit(t)

				<-secondStmtDone
				require.Equal(t, primary, secondPrimary)
				require.Nil(t, secondLogEntry)
				txSecond.Commit(t)

				previousPrimary = primary
			}
		})
	}
}
