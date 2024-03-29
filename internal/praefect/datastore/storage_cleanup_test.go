package datastore

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testdb"
)

func TestStorageCleanup_Populate(t *testing.T) {
	t.Parallel()
	ctx := testhelper.Context(t)
	db := testdb.New(t)
	storageCleanup := NewStorageCleanup(db.DB)

	require.NoError(t, storageCleanup.Populate(ctx, "praefect", "gitaly-1"))
	actual := getAllStoragesCleanup(t, db)
	single := []storageCleanupRow{{ClusterPath: ClusterPath{VirtualStorage: "praefect", Storage: "gitaly-1"}}}
	require.Equal(t, single, actual)

	err := storageCleanup.Populate(ctx, "praefect", "gitaly-1")
	require.NoError(t, err, "population of the same data should not generate an error")
	actual = getAllStoragesCleanup(t, db)
	require.Equal(t, single, actual, "same data should not create additional rows or change existing")

	require.NoError(t, storageCleanup.Populate(ctx, "default", "gitaly-2"))
	multiple := append(single, storageCleanupRow{ClusterPath: ClusterPath{VirtualStorage: "default", Storage: "gitaly-2"}})
	actual = getAllStoragesCleanup(t, db)
	require.ElementsMatch(t, multiple, actual, "new data should create additional row")
}

func TestStorageCleanup_AcquireNextStorage(t *testing.T) {
	t.Parallel()
	ctx := testhelper.Context(t)
	db := testdb.New(t)
	storageCleanup := NewStorageCleanup(db.DB)

	const noExpiration = 24 * time.Hour

	t.Run("ok", func(t *testing.T) {
		db.TruncateAll(t)
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g1"))

		clusterPath, release, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.Equal(t, &ClusterPath{VirtualStorage: "vs", Storage: "g1"}, clusterPath)
	})

	t.Run("last_run condition", func(t *testing.T) {
		db.TruncateAll(t)
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g1"))
		// Acquire it to initialize last_run column.
		_, release, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())

		clusterPath, release, err := storageCleanup.AcquireNextStorage(ctx, time.Hour, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.Nil(t, clusterPath, "no result expected as there can't be such entries")
	})

	t.Run("sorting based on storage name as no executions done yet", func(t *testing.T) {
		db.TruncateAll(t)
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g1"))
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g2"))
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g3"))

		clusterPath, release, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.Equal(t, &ClusterPath{VirtualStorage: "vs", Storage: "g1"}, clusterPath)
	})

	t.Run("sorting based on storage name and last_run", func(t *testing.T) {
		db.TruncateAll(t)
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g1"))
		_, release, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g2"))

		clusterPath, release, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.Equal(t, &ClusterPath{VirtualStorage: "vs", Storage: "g2"}, clusterPath)
	})

	t.Run("sorting based on last_run", func(t *testing.T) {
		db.TruncateAll(t)
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g1"))
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g2"))
		clusterPath, release, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.Equal(t, &ClusterPath{VirtualStorage: "vs", Storage: "g1"}, clusterPath)
		clusterPath, release, err = storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.Equal(t, &ClusterPath{VirtualStorage: "vs", Storage: "g2"}, clusterPath)

		clusterPath, release, err = storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NoError(t, release())
		require.Equal(t, &ClusterPath{VirtualStorage: "vs", Storage: "g1"}, clusterPath)
	})

	t.Run("already acquired won't be acquired until released", func(t *testing.T) {
		db.TruncateAll(t)
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g1"))
		_, release1, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)

		clusterPath, release2, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.Nil(t, clusterPath, clusterPath)
		require.NoError(t, release1())
		require.NoError(t, release2())

		clusterPath, release3, err := storageCleanup.AcquireNextStorage(ctx, 0, noExpiration)
		require.NoError(t, err)
		require.NotNil(t, clusterPath)
		require.NoError(t, release3())
	})

	t.Run("acquired for long time triggers update loop", func(t *testing.T) {
		db.TruncateAll(t)
		require.NoError(t, storageCleanup.Populate(ctx, "vs", "g1"))
		start := time.Now().UTC()
		_, release, err := storageCleanup.AcquireNextStorage(ctx, 0, 200*time.Millisecond)
		require.NoError(t, err)

		// Make sure the triggered_at column has a non NULL value after the record is acquired.
		check1 := getAllStoragesCleanup(t, db)
		require.Len(t, check1, 1)
		require.True(t, check1[0].TriggeredAt.Valid)
		require.True(t, check1[0].TriggeredAt.Time.After(start), check1[0].TriggeredAt.Time.String(), start.String())

		require.Eventuallyf(t, func() bool {
			check2 := getAllStoragesCleanup(t, db)
			require.Len(t, check2, 1)
			require.True(t, check2[0].TriggeredAt.Valid)
			return check2[0].TriggeredAt.Time.After(check1[0].TriggeredAt.Time)
		}, time.Minute, 200*time.Millisecond, "goroutine failed to update triggered_at column")

		require.NoError(t, release())

		// Make sure the triggered_at column has a NULL value after the record is released.
		check3 := getAllStoragesCleanup(t, db)
		require.Len(t, check3, 1)
		require.False(t, check3[0].TriggeredAt.Valid)
	})
}

func TestStorageCleanup_Exists(t *testing.T) {
	t.Parallel()
	ctx := testhelper.Context(t)

	db := testdb.New(t)

	repoStore := NewPostgresRepositoryStore(db.DB, nil)
	require.NoError(t, repoStore.CreateRepository(ctx, 0, "vs", "p/1", "replica-path-1", "g1", []string{"g2", "g3"}, nil, false, false))
	require.NoError(t, repoStore.CreateRepository(ctx, 1, "vs", "p/2", "replica-path-2", "g1", []string{"g2", "g3"}, nil, false, false))
	storageCleanup := NewStorageCleanup(db.DB)

	for _, tc := range []struct {
		desc                 string
		virtualStorage       string
		storage              string
		relativeReplicaPaths []string
		out                  []string
	}{
		{
			desc:                 "multiple doesn't exist",
			virtualStorage:       "vs",
			storage:              "g1",
			relativeReplicaPaths: []string{"replica-path-1", "replica-path-2", "path/x", "path/y"},
			out:                  []string{"path/x", "path/y"},
		},
		{
			desc:                 "duplicates",
			virtualStorage:       "vs",
			storage:              "g1",
			relativeReplicaPaths: []string{"replica-path-1", "path/x", "path/x"},
			out:                  []string{"path/x"},
		},
		{
			desc:                 "all exist",
			virtualStorage:       "vs",
			storage:              "g1",
			relativeReplicaPaths: []string{"replica-path-1", "replica-path-2"},
			out:                  nil,
		},
		{
			desc:                 "all doesn't exist",
			virtualStorage:       "vs",
			storage:              "g1",
			relativeReplicaPaths: []string{"path/x", "path/y", "path/z"},
			out:                  []string{"path/x", "path/y", "path/z"},
		},
		{
			desc:                 "doesn't exist because of storage",
			virtualStorage:       "vs",
			storage:              "stub",
			relativeReplicaPaths: []string{"path/x"},
			out:                  []string{"path/x"},
		},
		{
			desc:                 "doesn't exist because of virtual storage",
			virtualStorage:       "stub",
			storage:              "g1",
			relativeReplicaPaths: []string{"path/x"},
			out:                  []string{"path/x"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := storageCleanup.DoesntExist(ctx, tc.virtualStorage, tc.storage, tc.relativeReplicaPaths)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.out, res)
		})
	}
}

type storageCleanupRow struct {
	ClusterPath
	LastRun     sql.NullTime
	TriggeredAt sql.NullTime
}

func getAllStoragesCleanup(tb testing.TB, db testdb.DB) []storageCleanupRow {
	rows, err := db.Query(`SELECT * FROM storage_cleanups`)
	require.NoError(tb, err)
	defer func() {
		require.NoError(tb, rows.Close())
	}()

	var res []storageCleanupRow
	for rows.Next() {
		var dst storageCleanupRow
		err := rows.Scan(&dst.VirtualStorage, &dst.Storage, &dst.LastRun, &dst.TriggeredAt)
		require.NoError(tb, err)
		res = append(res, dst)
	}
	require.NoError(tb, rows.Err())
	return res
}
