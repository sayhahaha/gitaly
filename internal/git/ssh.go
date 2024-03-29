package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/perm"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
)

// BuildSSHInvocation builds a command line to invoke SSH with the provided key and known hosts.
// Both are optional.
func BuildSSHInvocation(ctx context.Context, sshKey, knownHosts string) (string, func(), error) {
	const sshCommand = "ssh"
	if sshKey == "" && knownHosts == "" {
		return sshCommand, func() {}, nil
	}

	tmpDir, err := os.MkdirTemp("", "gitaly-ssh-invocation")
	if err != nil {
		return "", func() {}, fmt.Errorf("create temporary directory: %w", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.FromContext(ctx).WithError(err).Error("failed to remove tmp directory with ssh key/config")
		}
	}

	args := []string{sshCommand}
	if sshKey != "" {
		sshKeyFile := filepath.Join(tmpDir, "ssh-key")
		if err := os.WriteFile(sshKeyFile, []byte(sshKey), perm.PrivateWriteOnceFile); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("create ssh key file: %w", err)
		}

		args = append(args, "-oIdentitiesOnly=yes", "-oIdentityFile="+sshKeyFile)
	}

	if knownHosts != "" {
		knownHostsFile := filepath.Join(tmpDir, "known-hosts")
		if err := os.WriteFile(knownHostsFile, []byte(knownHosts), perm.PrivateWriteOnceFile); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("create known hosts file: %w", err)
		}

		args = append(args, "-oStrictHostKeyChecking=yes", "-oCheckHostIP=no", "-oUserKnownHostsFile="+knownHostsFile)
	}

	return strings.Join(args, " "), cleanup, nil
}
