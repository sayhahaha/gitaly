package datastore

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testdb"
)

// The test setup takes a lot of time, so it is better to run each sub-benchmark separately with limit on number of repeats.
func BenchmarkPostgresRepositoryStore_GetConsistentStorages(b *testing.B) {
	b.Run("extra-small", func(b *testing.B) {
		benchmarkGetConsistentStorages(b, 3, 1000)
	})

	b.Run("small", func(b *testing.B) {
		benchmarkGetConsistentStorages(b, 3, 10_000)
	})

	b.Run("medium", func(b *testing.B) {
		benchmarkGetConsistentStorages(b, 3, 100_000)
	})

	b.Run("large", func(b *testing.B) {
		benchmarkGetConsistentStorages(b, 3, 1_000_000)
	})

	b.Run("huge", func(b *testing.B) {
		benchmarkGetConsistentStorages(b, 6, 1_000_000)
	})
}

func benchmarkGetConsistentStorages(b *testing.B, nstorages, nrepositories int) {
	db := testdb.New(b)
	ctx := testhelper.Context(b)

	for n := 0; n < b.N; n++ {
		b.StopTimer()

		db.Truncate(b, "storage_repositories")

		var storages []string
		for i := 0; i < nstorages; i++ {
			storages = append(storages, "gitaly-"+strconv.Itoa(i))
		}

		repoStore := NewPostgresRepositoryStore(db, map[string][]string{"vs": storages})

		_, err := db.DB.ExecContext(
			ctx,
			`INSERT INTO storage_repositories(virtual_storage, relative_path, storage, generation)
			SELECT 'vs', '/path/repo/' || R.I, 'gitaly-' || S.I, 1
			FROM GENERATE_SERIES(1, $1) R(I)
			CROSS JOIN GENERATE_SERIES(1, $2) S(I)`,
			nrepositories, nstorages,
		)
		require.NoError(b, err)

		b.StartTimer()
		_, _, err = repoStore.GetConsistentStorages(ctx, "vs", "/path/repo/"+strconv.Itoa(nrepositories/2))
		b.StopTimer()

		require.NoError(b, err)
	}
}
