//go:build !gitaly_test_sha256

package repository

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/housekeeping"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/localrepo"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/stats"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
)

func TestOptimizeRepository(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg, client := setupRepositoryServiceWithoutRepo(t)

	t.Run("gitconfig credentials get pruned", func(t *testing.T) {
		t.Parallel()

		repo, repoPath := gittest.CreateRepository(t, ctx, cfg)
		gitconfigPath := filepath.Join(repoPath, "config")

		readConfig := func() []string {
			return strings.Split(text.ChompBytes(gittest.Exec(t, cfg, "config", "--file", gitconfigPath, "--list")), "\n")
		}

		configWithSecrets := readConfig()
		configWithStrippedSecrets := readConfig()

		// Set up a gitconfig with all sorts of credentials.
		for key, shouldBeStripped := range map[string]bool{
			"http.http://localhost:51744/60631c8695bf041a808759a05de53e36a73316aacb502824fabbb0c6055637c1.git.extraHeader": true,
			"http.http://localhost:51744/60631c8695bf041a808759a05de53e36a73316aacb502824fabbb0c6055637c2.git.extraHeader": true,
			"hTTp.http://localhost:51744/60631c8695bf041a808759a05de53e36a73316aacb502824fabbb0c6055637c5.git.ExtrAheaDeR": true,
			"http.http://extraheader/extraheader/extraheader.git.extraHeader":                                              true,
			// This line should not get stripped as Git wouldn't even know how to
			// interpret it due to the `https` prefix. Git only knows about `http`.
			"https.https://localhost:51744/60631c8695bf041a808759a05de53e36a73316aacb502824fabbb0c6055637c5.git.extraHeader": false,
			// This one should not get stripped as its prefix is wrong.
			"randomStart-http.http://localhost:51744/60631c8695bf041a808759a05de53e36a73316aacb502824fabbb0c6055637c3.git.extraHeader": false,
			// Same here, this one should not get stripped because its suffix is wrong.
			"http.http://localhost:51744/60631c8695bf041a808759a05de53e36a73316aacb502824fabbb0c6055637c4.git.extraHeader-randomEnd": false,
		} {
			value := "Authorization: Basic secret-password"
			line := fmt.Sprintf("%s=%s", strings.ToLower(key), value)

			gittest.Exec(t, cfg, "config", "--file", gitconfigPath, key, value)

			configWithSecrets = append(configWithSecrets, line)
			if !shouldBeStripped {
				configWithStrippedSecrets = append(configWithStrippedSecrets, line)
			}
		}
		require.Equal(t, configWithSecrets, readConfig())

		// Calling OptimizeRepository should cause us to strip any of the added creds from
		// the gitconfig.
		_, err := client.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
			Repository: repo,
		})
		require.NoError(t, err)
		// The gitconfig should not contain any of the stripped gitconfig values anymore.
		require.Equal(t, configWithStrippedSecrets, readConfig())
	})

	t.Run("up-to-date packfile does not get repacked", func(t *testing.T) {
		t.Parallel()

		repo, repoPath := gittest.CreateRepository(t, ctx, cfg)

		// Write a commit and force-repack the whole repository. This is to ensure that the
		// repository is in a state where it shouldn't need to be repacked.
		gittest.WriteCommit(t, cfg, repoPath, gittest.WithBranch("master"))
		_, err := client.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
			Repository: repo,
			Strategy:   gitalypb.OptimizeRepositoryRequest_STRATEGY_EAGER,
		})
		require.NoError(t, err)
		// We should have a single packfile now.
		packfiles, err := filepath.Glob(filepath.Join(repoPath, "objects", "pack", "pack-*.pack"))
		require.NoError(t, err)
		require.Len(t, packfiles, 1)

		// Now we do a second, lazy optimization of the repository. This time around we
		// should see that the repository was in a well-defined state already, so we should
		// not perform any optimization.
		_, err = client.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
			Repository: repo,
		})
		require.NoError(t, err)

		// The packfile should not have changed.
		updatedPackfiles, err := filepath.Glob(filepath.Join(repoPath, "objects", "pack", "pack-*.pack"))
		require.NoError(t, err)
		require.Equal(t, packfiles, updatedPackfiles)
	})

	t.Run("missing bitmap causes full repack", func(t *testing.T) {
		t.Parallel()

		repo, repoPath := gittest.CreateRepository(t, ctx, cfg)
		gittest.WriteCommit(t, cfg, repoPath, gittest.WithBranch("master"))

		bitmaps, err := filepath.Glob(filepath.Join(repoPath, "objects", "pack", "*.bitmap"))
		require.NoError(t, err)
		require.Empty(t, bitmaps)

		// Even though the repository doesn't have a lot of objects and we're not performing
		// an eager optimization, we should still see that the optimization decides to write
		// out a new bitmap via a full repack. This is so that all repositories will have a
		// bitmap available.
		_, err = client.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
			Repository: repo,
		})
		require.NoError(t, err)

		bitmaps, err = filepath.Glob(filepath.Join(repoPath, "objects", "pack", "*.bitmap"))
		require.NoError(t, err)
		require.NotEmpty(t, bitmaps)
	})

	t.Run("optimizing repository without commit-graph bloom filters", func(t *testing.T) {
		t.Parallel()

		repo, repoPath := gittest.CreateRepository(t, ctx, cfg)
		gittest.WriteCommit(t, cfg, repoPath, gittest.WithBranch("master"))

		// Prepare the repository so that it has a commit-graph, but that commit-graph is
		// missing bloom filters.
		gittest.Exec(t, cfg, "-C", repoPath, "commit-graph", "write", "--split", "--reachable")
		commitGraphInfo, err := stats.CommitGraphInfoForRepository(repoPath)
		require.NoError(t, err)
		require.Equal(t, stats.CommitGraphInfo{
			Exists:                 true,
			HasBloomFilters:        false,
			CommitGraphChainLength: 1,
		}, commitGraphInfo)
		// TODO: Due to a bug we also need to repack the repository, or otherwise we would
		// not rewrite the commit-graph.
		gittest.Exec(t, cfg, "-C", repoPath, "repack", "-Adb")

		// As a result, OptimizeRepository should rewrite the commit-graph.
		_, err = client.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
			Repository: repo,
		})
		require.NoError(t, err)

		// Which means that we now should see that bloom filters exist.
		commitGraphInfo, err = stats.CommitGraphInfoForRepository(repoPath)
		require.NoError(t, err)
		require.Equal(t, stats.CommitGraphInfo{
			Exists:                 true,
			HasBloomFilters:        true,
			CommitGraphChainLength: 1,
		}, commitGraphInfo)
	})

	t.Run("empty ref directories get pruned after grace period", func(t *testing.T) {
		t.Parallel()

		repo, repoPath := gittest.CreateRepository(t, ctx, cfg)

		// Git will leave behind empty refs directories at times. In order to not slow down
		// enumerating refs we want to make sure that they get cleaned up properly.
		emptyRefsDir := filepath.Join(repoPath, "refs", "merge-requests", "1")
		require.NoError(t, os.MkdirAll(emptyRefsDir, 0o755))

		// But we don't expect the first call to OptimizeRepository to do anything. This is
		// because we have a grace period so that we don't delete empty ref directories that
		// have just been created by a concurrently running Git process.
		_, err := client.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
			Repository: repo,
		})
		require.NoError(t, err)
		require.DirExists(t, emptyRefsDir)

		// Change the modification time of the complete repository to be older than a day.
		require.NoError(t, filepath.WalkDir(repoPath, func(path string, _ fs.DirEntry, err error) error {
			require.NoError(t, err)
			oneDayAgo := time.Now().Add(-24 * time.Hour)
			require.NoError(t, os.Chtimes(path, oneDayAgo, oneDayAgo))
			return nil
		}))

		// Now the second call to OptimizeRepository should indeed clean up the empty refs
		// directories.
		_, err = client.OptimizeRepository(ctx, &gitalypb.OptimizeRepositoryRequest{
			Repository: repo,
		})
		require.NoError(t, err)
		// We shouldn't have removed the top-level "refs" directory.
		require.DirExists(t, filepath.Join(repoPath, "refs"))
		// But the other two directories should be gone.
		require.NoDirExists(t, filepath.Join(repoPath, "refs", "merge-requests"))
		require.NoDirExists(t, filepath.Join(repoPath, "refs", "merge-requests", "1"))
	})
}

type mockHousekeepingManager struct {
	housekeeping.Manager
	cfgCh chan housekeeping.OptimizeRepositoryConfig
}

func (m mockHousekeepingManager) OptimizeRepository(_ context.Context, _ *localrepo.Repo, opts ...housekeeping.OptimizeRepositoryOption) error {
	var cfg housekeeping.OptimizeRepositoryConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	m.cfgCh <- cfg
	return nil
}

func TestOptimizeRepository_strategy(t *testing.T) {
	t.Parallel()

	housekeepingManager := mockHousekeepingManager{
		cfgCh: make(chan housekeeping.OptimizeRepositoryConfig, 1),
	}

	ctx := testhelper.Context(t)
	cfg, client := setupRepositoryServiceWithoutRepo(t, testserver.WithHousekeepingManager(housekeepingManager))

	repoProto, _ := gittest.CreateRepository(t, ctx, cfg)

	for _, tc := range []struct {
		desc        string
		request     *gitalypb.OptimizeRepositoryRequest
		expectedCfg housekeeping.OptimizeRepositoryConfig
	}{
		{
			desc: "no strategy",
			request: &gitalypb.OptimizeRepositoryRequest{
				Repository: repoProto,
			},
			expectedCfg: housekeeping.OptimizeRepositoryConfig{
				Strategy: housekeeping.HeuristicalOptimizationStrategy{},
			},
		},
		{
			desc: "heuristical strategy",
			request: &gitalypb.OptimizeRepositoryRequest{
				Repository: repoProto,
				Strategy:   gitalypb.OptimizeRepositoryRequest_STRATEGY_HEURISTICAL,
			},
			expectedCfg: housekeeping.OptimizeRepositoryConfig{
				Strategy: housekeeping.HeuristicalOptimizationStrategy{},
			},
		},
		{
			desc: "eager strategy",
			request: &gitalypb.OptimizeRepositoryRequest{
				Repository: repoProto,
				Strategy:   gitalypb.OptimizeRepositoryRequest_STRATEGY_EAGER,
			},
			expectedCfg: housekeeping.OptimizeRepositoryConfig{
				Strategy: housekeeping.EagerOptimizationStrategy{},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := client.OptimizeRepository(ctx, tc.request)
			require.NoError(t, err)
			testhelper.ProtoEqual(t, &gitalypb.OptimizeRepositoryResponse{}, response)

			require.Equal(t, tc.expectedCfg, <-housekeepingManager.cfgCh)
		})
	}
}

func TestOptimizeRepository_validation(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg, repo, _, client := setupRepositoryService(t, ctx)

	for _, tc := range []struct {
		desc        string
		request     *gitalypb.OptimizeRepositoryRequest
		expectedErr error
	}{
		{
			desc:    "empty repository",
			request: &gitalypb.OptimizeRepositoryRequest{},
			expectedErr: helper.ErrInvalidArgumentf(testhelper.GitalyOrPraefect(
				"empty Repository",
				"repo scoped: empty Repository",
			)),
		},
		{
			desc: "invalid repository storage",
			request: &gitalypb.OptimizeRepositoryRequest{
				Repository: &gitalypb.Repository{
					StorageName:  "non-existent",
					RelativePath: repo.GetRelativePath(),
				},
			},
			expectedErr: helper.ErrInvalidArgumentf(testhelper.GitalyOrPraefect(
				`GetStorageByName: no such storage: "non-existent"`,
				"repo scoped: invalid Repository"),
			),
		},
		{
			desc: "invalid repository path",
			request: &gitalypb.OptimizeRepositoryRequest{
				Repository: &gitalypb.Repository{
					StorageName:  repo.GetStorageName(),
					RelativePath: "path/not/exist",
				},
			},
			expectedErr: helper.ErrNotFoundf(testhelper.GitalyOrPraefect(
				fmt.Sprintf(`GetRepoPath: not a git repository: "%s/path/not/exist"`, cfg.Storages[0].Path),
				`routing repository maintenance: getting repository metadata: repository not found`,
			)),
		},
		{
			desc: "invalid optimization strategy",
			request: &gitalypb.OptimizeRepositoryRequest{
				Repository: repo,
				Strategy:   12,
			},
			expectedErr: helper.ErrInvalidArgumentf("unsupported optimization strategy 12"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := client.OptimizeRepository(ctx, tc.request)
			testhelper.RequireGrpcError(t, tc.expectedErr, err)
		})
	}
}
