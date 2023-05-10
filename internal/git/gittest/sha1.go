//go:build !gitaly_test_sha256

package gittest

import "gitlab.com/gitlab-org/gitaly/v16/internal/git"

// DefaultObjectHash is the default hash used for running tests.
var DefaultObjectHash = git.ObjectHashSHA1
