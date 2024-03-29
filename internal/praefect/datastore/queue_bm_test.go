package datastore

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testdb"
)

func BenchmarkPostgresReplicationEventQueue_Acknowledge(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		benchmarkPostgresReplicationEventQueueAcknowledge(b, map[JobState]int{JobStateReady: 10, JobStateInProgress: 10, JobStateFailed: 10})
	})

	b.Run("medium", func(b *testing.B) {
		benchmarkPostgresReplicationEventQueueAcknowledge(b, map[JobState]int{JobStateReady: 1_000, JobStateInProgress: 100, JobStateFailed: 100})
	})

	b.Run("big", func(b *testing.B) {
		benchmarkPostgresReplicationEventQueueAcknowledge(b, map[JobState]int{JobStateReady: 100_000, JobStateInProgress: 100, JobStateFailed: 100})
	})

	b.Run("huge", func(b *testing.B) {
		benchmarkPostgresReplicationEventQueueAcknowledge(b, map[JobState]int{JobStateReady: 1_000_000, JobStateInProgress: 100, JobStateFailed: 100})
	})
}

func benchmarkPostgresReplicationEventQueueAcknowledge(b *testing.B, setup map[JobState]int) {
	db := testdb.New(b)
	ctx := testhelper.Context(b)

	queue := PostgresReplicationEventQueue{db.DB}
	eventTmpl := ReplicationEvent{
		Job: ReplicationJob{
			Change:            UpdateRepo,
			RelativePath:      "/project/path-",
			TargetNodeStorage: "gitaly-1",
			SourceNodeStorage: "gitaly-0",
			VirtualStorage:    "praefect",
		},
	}

	getEventIDs := func(events []ReplicationEvent) []uint64 {
		ids := make([]uint64, len(events))
		for i, event := range events {
			ids[i] = event.ID
		}
		return ids
	}

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		b.ResetTimer()

		db.TruncateAll(b)

		total := 0
		for _, count := range setup {
			// at first we need to enqueue all events and then move them to proper states
			total += count
		}

		_, err := db.DB.ExecContext(
			ctx,
			`INSERT INTO replication_queue (state, lock_id, job)
			SELECT 'ready', 'praefect|gitaly-1|/project/path-'|| T.I, ('{"change":"update","relative_path":"/project/path-'|| T.I || '","virtual_storage":"praefect","source_node_storage":"gitaly-0","target_node_storage":"gitaly-1"}')::JSONB
			FROM GENERATE_SERIES(1, $1) T(I)`,
			total,
		)
		require.NoError(b, err)

		_, err = db.DB.ExecContext(
			ctx,
			`INSERT INTO replication_queue_lock
			SELECT DISTINCT lock_id FROM replication_queue`,
		)
		require.NoError(b, err)

		events, err := queue.Dequeue(ctx, eventTmpl.Job.VirtualStorage, eventTmpl.Job.TargetNodeStorage, setup[JobStateFailed]+setup[JobStateInProgress])
		require.NoError(b, err)

		_, err = queue.Acknowledge(ctx, JobStateFailed, getEventIDs(events[:setup[JobStateFailed]]))
		require.NoError(b, err)

		events, err = queue.Dequeue(ctx, eventTmpl.Job.VirtualStorage, eventTmpl.Job.TargetNodeStorage, 10)
		require.NoError(b, err)

		b.StartTimer()
		_, err = queue.Acknowledge(ctx, JobStateCompleted, getEventIDs(events))
		b.StopTimer()
		require.NoError(b, err)
	}
}
