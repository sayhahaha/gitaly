package gittest

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

const (
	// GlRepository is the default repository name for newly created test
	// repos.
	GlRepository = "project-1"
	// GlProjectPath is the default project path for newly created test
	// repos.
	GlProjectPath = "gitlab-org/gitlab-test"
)

// InitRepoDir creates a temporary directory for a repo, without initializing it
func InitRepoDir(t testing.TB, storagePath, relativePath string) *gitalypb.Repository {
	repoPath := filepath.Join(storagePath, relativePath, "..")
	require.NoError(t, os.MkdirAll(repoPath, 0755), "making repo parent dir")
	return &gitalypb.Repository{
		StorageName:   "default",
		RelativePath:  relativePath,
		GlRepository:  GlRepository,
		GlProjectPath: GlProjectPath,
	}
}

// InitBareRepo creates a new bare repository
func InitBareRepo(t testing.TB) (*gitalypb.Repository, string, func()) {
	return initRepoAt(t, true, config.Storage{Name: "default", Path: testhelper.GitlabTestStoragePath()})
}

// InitBareRepoAt creates a new bare repository in the storage
func InitBareRepoAt(t testing.TB, storage config.Storage) (*gitalypb.Repository, string, func()) {
	return initRepoAt(t, true, storage)
}

// InitRepoWithWorktree creates a new repository with a worktree
func InitRepoWithWorktree(t testing.TB) (*gitalypb.Repository, string, func()) {
	return initRepoAt(t, false, config.Storage{Name: "default", Path: testhelper.GitlabTestStoragePath()})
}

// NewObjectPoolName returns a random pool repository name in format
// '@pools/[0-9a-z]{2}/[0-9a-z]{2}/[0-9a-z]{64}.git'.
func NewObjectPoolName(t testing.TB) string {
	return filepath.Join("@pools", newDiskHash(t)+".git")
}

// NewRepositoryName returns a random repository hash
// in format '@hashed/[0-9a-f]{2}/[0-9a-f]{2}/[0-9a-f]{64}(.git)?'.
func NewRepositoryName(t testing.TB, bare bool) string {
	suffix := ""
	if bare {
		suffix = ".git"
	}

	return filepath.Join("@hashed", newDiskHash(t)+suffix)
}

// newDiskHash generates a random directory path following the Rails app's
// approach in the hashed storage module, formatted as '[0-9a-f]{2}/[0-9a-f]{2}/[0-9a-f]{64}'.
// https://gitlab.com/gitlab-org/gitlab/-/blob/f5c7d8eb1dd4eee5106123e04dec26d277ff6a83/app/models/storage/hashed.rb#L38-43
func newDiskHash(t testing.TB) string {
	// rails app calculates a sha256 and uses its hex representation
	// as the directory path
	b, err := text.RandomHex(sha256.Size)
	require.NoError(t, err)
	return filepath.Join(b[0:2], b[2:4], b)
}

func initRepoAt(t testing.TB, bare bool, storage config.Storage) (*gitalypb.Repository, string, func()) {
	relativePath := NewRepositoryName(t, bare)
	repoPath := filepath.Join(storage.Path, relativePath)

	args := []string{"init"}
	if bare {
		args = append(args, "--bare")
	}

	testhelper.MustRunCommand(t, nil, "git", append(args, repoPath)...)

	repo := InitRepoDir(t, storage.Path, relativePath)
	repo.StorageName = storage.Name
	if !bare {
		repo.RelativePath = filepath.Join(repo.RelativePath, ".git")
	}

	return repo, repoPath, func() { require.NoError(t, os.RemoveAll(repoPath)) }
}

// CloneRepoAtStorageRoot clones a new copy of test repository under a subdirectory in the storage root.
func CloneRepoAtStorageRoot(t testing.TB, storageRoot, relativePath string) *gitalypb.Repository {
	repo, _, _ := cloneRepo(t, storageRoot, relativePath, true)
	return repo
}

// CloneRepoAtStorage clones a new copy of test repository under a subdirectory in the storage root.
func CloneRepoAtStorage(t testing.TB, storage config.Storage, relativePath string) (*gitalypb.Repository, string, testhelper.Cleanup) {
	repo, repoPath, cleanup := cloneRepo(t, storage.Path, relativePath, true)
	repo.StorageName = storage.Name
	return repo, repoPath, cleanup
}

// CloneRepo creates a bare copy of the test repository.
func CloneRepo(t testing.TB) (repo *gitalypb.Repository, repoPath string, cleanup func()) {
	return cloneRepo(t, testhelper.GitlabTestStoragePath(), NewRepositoryName(t, true), true)
}

// CloneRepoWithWorktree creates a copy of the test repository with a worktree. This is allows you
// to run normal 'non-bare' Git commands.
func CloneRepoWithWorktree(t testing.TB) (repo *gitalypb.Repository, repoPath string, cleanup func()) {
	return cloneRepo(t, testhelper.GitlabTestStoragePath(), NewRepositoryName(t, false), false)
}

// testRepositoryPath returns the absolute path of local 'gitlab-org/gitlab-test.git' clone.
// It is cloned under the path by the test preparing step of make.
func testRepositoryPath(t testing.TB) string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		require.Fail(t, "could not get caller info")
	}

	path := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "_build", "testrepos", "gitlab-test.git")
	if !isValidRepoPath(path) {
		makePath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..")
		makeTarget := "prepare-test-repos"
		log.Printf("local clone of 'gitlab-org/gitlab-test.git' not found in %q, running `make %v`", path, makeTarget)
		testhelper.MustRunCommand(t, nil, "make", "-C", makePath, makeTarget)
	}

	return path
}

// isValidRepoPath checks whether a valid git repository exists at the given path.
func isValidRepoPath(absolutePath string) bool {
	if _, err := os.Stat(filepath.Join(absolutePath, "objects")); err != nil {
		return false
	}

	return true
}

func cloneRepo(t testing.TB, storageRoot, relativePath string, bare bool) (repo *gitalypb.Repository, repoPath string, cleanup func()) {
	repoPath = filepath.Join(storageRoot, relativePath)

	repo = InitRepoDir(t, storageRoot, relativePath)
	args := []string{"clone", "--no-hardlinks", "--dissociate"}
	if bare {
		args = append(args, "--bare")
	} else {
		// For non-bare repos the relative path is the .git folder inside the path
		repo.RelativePath = filepath.Join(relativePath, ".git")
	}

	testhelper.MustRunCommand(t, nil, "git", append(args, testRepositoryPath(t), repoPath)...)

	return repo, repoPath, func() { require.NoError(t, os.RemoveAll(repoPath)) }
}

// AddWorktreeArgs returns git command arguments for adding a worktree at the
// specified repo
func AddWorktreeArgs(repoPath, worktreeName string) []string {
	return []string{"-C", repoPath, "worktree", "add", "--detach", worktreeName}
}

// AddWorktree creates a worktree in the repository path for tests
func AddWorktree(t testing.TB, repoPath string, worktreeName string) {
	testhelper.MustRunCommand(t, nil, "git", AddWorktreeArgs(repoPath, worktreeName)...)
}