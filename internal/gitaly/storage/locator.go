package storage

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Repository represents a storage-scoped repository.
type Repository interface {
	GetStorageName() string
	GetRelativePath() string
	GetGitObjectDirectory() string
	GetGitAlternateObjectDirectories() []string
}

// RepoPathEqual compares if two repositories are in the same location
func RepoPathEqual(a, b Repository) bool {
	return a.GetStorageName() == b.GetStorageName() &&
		a.GetRelativePath() == b.GetRelativePath()
}

// Locator allows to get info about location of the repository or storage at the local file system.
type Locator interface {
	// GetRepoPath returns the full path of the repository referenced by an RPC Repository message.
	// By default, it verifies that the path is an existing git directory. However, if invoked with
	// the `GetRepoPathOption` produced by `WithRepositoryVerificationSkipped()`, this validation
	// will be skipped. The errors returned are gRPC errors with relevant error codes and should be
	// passed back to gRPC without further decoration.
	GetRepoPath(repo Repository, opts ...GetRepoPathOption) (string, error)
	// GetPath returns the path of the repo passed as first argument. An error is
	// returned when either the storage can't be found or the path includes
	// constructs trying to perform directory traversal.
	GetPath(repo Repository) (string, error)
	// GetStorageByName will return the path for the storage, which is fetched by
	// its key. An error is return if it cannot be found.
	GetStorageByName(storageName string) (string, error)

	// CacheDir returns the path to the cache dir for a storage.
	CacheDir(storageName string) (string, error)
	// TempDir returns the path to the temp dir for a storage.
	TempDir(storageName string) (string, error)
	// StateDir returns the path to the state dir for a storage.
	StateDir(storageName string) (string, error)
}

// GetRepoPathConfig is used to configure GetRepoPath.
type GetRepoPathConfig struct {
	// SkipRepositoryVerification will cause GetRepoPath to skip verification the verification whether the
	// computed path is an actual Git repository or not.
	SkipRepositoryVerification bool
}

// GetRepoPathOption can be passed to GetRepoPath to change its default behavior.
type GetRepoPathOption func(*GetRepoPathConfig)

// WithRepositoryVerificationSkipped skips the repository path validation that ensures the Git
// directory is valid.
func WithRepositoryVerificationSkipped() GetRepoPathOption {
	return func(cfg *GetRepoPathConfig) {
		cfg.SkipRepositoryVerification = true
	}
}

//nolint:revive // This is unintentionally missing documentation.
var ErrRelativePathEscapesRoot = errors.New("relative path escapes root directory")

// ValidateRelativePath validates a relative path by joining it with rootDir and verifying the result
// is either rootDir or a path within rootDir. Returns clean relative path from rootDir to relativePath
// or an ErrRelativePathEscapesRoot if the resulting path is not contained within rootDir.
func ValidateRelativePath(rootDir, relativePath string) (string, error) {
	absPath := filepath.Join(rootDir, relativePath)
	if rootDir != absPath && !strings.HasPrefix(absPath, rootDir+string(os.PathSeparator)) {
		return "", ErrRelativePathEscapesRoot
	}

	return filepath.Rel(rootDir, absPath)
}

// IsGitDirectory checks if the directory passed as first argument looks like
// a valid git directory.
func IsGitDirectory(dir string) bool {
	if dir == "" {
		return false
	}

	for _, element := range []string{"objects", "refs", "HEAD"} {
		if _, err := os.Stat(filepath.Join(dir, element)); err != nil {
			return false
		}
	}

	// See: https://gitlab.com/gitlab-org/gitaly/issues/1339
	//
	// This is a workaround for Gitaly running on top of an NFS mount. There
	// is a Linux NFS v4.0 client bug where opening the packed-refs file can
	// either result in a stale file handle or stale data. This can happen if
	// git gc runs for a long time while keeping open the packed-refs file.
	// Running stat() on the file causes the kernel to revalidate the cached
	// directory entry. We don't actually care if this file exists.
	_, _ = os.Stat(filepath.Join(dir, "packed-refs"))

	return true
}

// QuarantineDirectoryPrefix returns a prefix for use in the temporary directory. The prefix is
// based on the relative repository path and will stay stable for any given repository. This allows
// us to verify that a given quarantine object directory indeed belongs to the repository at hand.
// Ideally, this function would directly be located in the quarantine module, but this is not
// possible due to cyclic dependencies.
func QuarantineDirectoryPrefix(repo Repository) string {
	hash := [20]byte{}
	if repo != nil {
		hash = sha1.Sum([]byte(repo.GetRelativePath()))
	}
	return fmt.Sprintf("quarantine-%x-", hash[:8])
}
