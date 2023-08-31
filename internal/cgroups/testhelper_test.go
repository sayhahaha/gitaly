//go:build linux

package cgroups

import (
	"testing"

	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
)

func TestMain(m *testing.M) {
	testhelper.Run(m)
}
