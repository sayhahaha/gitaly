package cgroups

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/config/cgroups"
	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/perm"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
)

func defaultCgroupsConfig() cgroups.Config {
	return cgroups.Config{
		HierarchyRoot: "gitaly",
		Repositories: cgroups.Repositories{
			Count:       3,
			MemoryBytes: 1024000,
			CPUShares:   256,
			CPUQuotaUs:  200,
		},
	}
}

func TestSetup_ParentCgroups(t *testing.T) {
	tests := []struct {
		name            string
		cfg             cgroups.Config
		wantMemoryBytes int
		wantCPUShares   int
		wantCPUQuotaUs  int
		wantCFSPeriod   int
	}{
		{
			name: "all config specified",
			cfg: cgroups.Config{
				MemoryBytes: 102400,
				CPUShares:   256,
				CPUQuotaUs:  200,
			},
			wantMemoryBytes: 102400,
			wantCPUShares:   256,
			wantCPUQuotaUs:  200,
			wantCFSPeriod:   int(cfsPeriodUs),
		},
		{
			name: "only memory limit set",
			cfg: cgroups.Config{
				MemoryBytes: 102400,
			},
			wantMemoryBytes: 102400,
		},
		{
			name: "only cpu shares set",
			cfg: cgroups.Config{
				CPUShares: 512,
			},
			wantCPUShares: 512,
		},
		{
			name: "only cpu quota set",
			cfg: cgroups.Config{
				CPUQuotaUs: 200,
			},
			wantCPUQuotaUs: 200,
			wantCFSPeriod:  int(cfsPeriodUs),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := newMock(t)
			pid := 1
			tt.cfg.HierarchyRoot = "gitaly"

			v1Manager := &CGroupV1Manager{
				cfg:       tt.cfg,
				hierarchy: mock.hierarchy,
				pid:       pid,
			}
			require.NoError(t, v1Manager.Setup())

			memoryLimitPath := filepath.Join(
				mock.root, "memory", "gitaly", fmt.Sprintf("gitaly-%d", pid), "memory.limit_in_bytes",
			)
			requireCgroup(t, memoryLimitPath, tt.wantMemoryBytes)

			cpuSharesPath := filepath.Join(
				mock.root, "cpu", "gitaly", fmt.Sprintf("gitaly-%d", pid), "cpu.shares",
			)
			requireCgroup(t, cpuSharesPath, tt.wantCPUShares)

			cpuCFSQuotaPath := filepath.Join(
				mock.root, "cpu", "gitaly", fmt.Sprintf("gitaly-%d", pid), "cpu.cfs_quota_us",
			)
			requireCgroup(t, cpuCFSQuotaPath, tt.wantCPUQuotaUs)

			cpuCFSPeriodPath := filepath.Join(
				mock.root, "cpu", "gitaly", fmt.Sprintf("gitaly-%d", pid), "cpu.cfs_period_us",
			)
			requireCgroup(t, cpuCFSPeriodPath, tt.wantCFSPeriod)
		})
	}
}

func TestSetup_RepoCgroups(t *testing.T) {
	tests := []struct {
		name            string
		cfg             cgroups.Repositories
		wantMemoryBytes int
		wantCPUShares   int
		wantCPUQuotaUs  int
		wantCFSPeriod   int
	}{
		{
			name:            "all config specified",
			cfg:             defaultCgroupsConfig().Repositories,
			wantMemoryBytes: 1024000,
			wantCPUShares:   256,
			wantCPUQuotaUs:  200,
			wantCFSPeriod:   int(cfsPeriodUs),
		},
		{
			name: "only memory limit set",
			cfg: cgroups.Repositories{
				MemoryBytes: 1024000,
			},
			wantMemoryBytes: 1024000,
		},
		{
			name: "only cpu shares set",
			cfg: cgroups.Repositories{
				CPUShares: 512,
			},
			wantCPUShares: 512,
		},
		{
			name: "only cpu quota set",
			cfg: cgroups.Repositories{
				CPUQuotaUs: 100,
			},
			wantCPUQuotaUs: 100,
			wantCFSPeriod:  int(cfsPeriodUs),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := newMock(t)

			pid := 1
			cfg := defaultCgroupsConfig()
			cfg.Repositories = tt.cfg
			cfg.Repositories.Count = 3

			v1Manager := &CGroupV1Manager{
				cfg:       cfg,
				hierarchy: mock.hierarchy,
				pid:       pid,
			}

			require.NoError(t, v1Manager.Setup())

			for i := 0; i < 3; i++ {
				memoryLimitPath := filepath.Join(
					mock.root, "memory", "gitaly", fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", i), "memory.limit_in_bytes",
				)
				requireCgroup(t, memoryLimitPath, tt.wantMemoryBytes)

				cpuSharesPath := filepath.Join(
					mock.root, "cpu", "gitaly", fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", i), "cpu.shares",
				)
				requireCgroup(t, cpuSharesPath, tt.wantCPUShares)

				cpuCFSQuotaPath := filepath.Join(
					mock.root, "cpu", "gitaly", fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", i), "cpu.cfs_quota_us",
				)
				requireCgroup(t, cpuCFSQuotaPath, tt.wantCPUQuotaUs)

				cpuCFSPeriodPath := filepath.Join(
					mock.root, "cpu", "gitaly", fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", i), "cpu.cfs_period_us",
				)
				requireCgroup(t, cpuCFSPeriodPath, tt.wantCFSPeriod)
			}
		})
	}
}

func TestAddCommand(t *testing.T) {
	mock := newMock(t)

	config := defaultCgroupsConfig()
	config.Repositories.Count = 10
	config.Repositories.MemoryBytes = 1024
	config.Repositories.CPUShares = 16

	pid := 1
	v1Manager1 := &CGroupV1Manager{
		cfg:       config,
		hierarchy: mock.hierarchy,
		pid:       pid,
	}
	require.NoError(t, v1Manager1.Setup())
	ctx := testhelper.Context(t)

	cmd2 := exec.CommandContext(ctx, "ls", "-hal", ".")
	require.NoError(t, cmd2.Run())

	v1Manager2 := &CGroupV1Manager{
		cfg:       config,
		hierarchy: mock.hierarchy,
		pid:       pid,
	}

	t.Run("without overridden key", func(t *testing.T) {
		_, err := v1Manager2.AddCommand(cmd2)
		require.NoError(t, err)

		checksum := crc32.ChecksumIEEE([]byte(strings.Join(cmd2.Args, "/")))
		groupID := uint(checksum) % config.Repositories.Count

		for _, s := range mock.subsystems {
			path := filepath.Join(mock.root, string(s.Name()), "gitaly",
				fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", groupID), "cgroup.procs")
			content := readCgroupFile(t, path)

			cmdPid, err := strconv.Atoi(string(content))
			require.NoError(t, err)

			require.Equal(t, cmd2.Process.Pid, cmdPid)
		}
	})

	t.Run("with overridden key", func(t *testing.T) {
		_, err := v1Manager2.AddCommand(cmd2, WithCgroupKey("foobar"))
		require.NoError(t, err)

		checksum := crc32.ChecksumIEEE([]byte("foobar"))
		groupID := uint(checksum) % config.Repositories.Count

		for _, s := range mock.subsystems {
			path := filepath.Join(mock.root, string(s.Name()), "gitaly",
				fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", groupID), "cgroup.procs")
			content := readCgroupFile(t, path)

			cmdPid, err := strconv.Atoi(string(content))
			require.NoError(t, err)

			require.Equal(t, cmd2.Process.Pid, cmdPid)
		}
	})
}

func TestCleanup(t *testing.T) {
	mock := newMock(t)

	pid := 1
	v1Manager := &CGroupV1Manager{
		cfg:       defaultCgroupsConfig(),
		hierarchy: mock.hierarchy,
		pid:       pid,
	}
	require.NoError(t, v1Manager.Setup())
	require.NoError(t, v1Manager.Cleanup())

	for i := 0; i < 3; i++ {
		memoryPath := filepath.Join(mock.root, "memory", "gitaly", fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", i))
		cpuPath := filepath.Join(mock.root, "cpu", "gitaly", fmt.Sprintf("gitaly-%d", pid), fmt.Sprintf("repos-%d", i))

		require.NoDirExists(t, memoryPath)
		require.NoDirExists(t, cpuPath)
	}
}

func TestMetrics(t *testing.T) {
	t.Parallel()

	mock := newMock(t)

	config := defaultCgroupsConfig()
	config.Repositories.Count = 1
	config.Repositories.MemoryBytes = 1048576
	config.Repositories.CPUShares = 16

	v1Manager1 := newV1Manager(config, 1)
	v1Manager1.hierarchy = mock.hierarchy

	mock.setupMockCgroupFiles(t, v1Manager1, 2)

	require.NoError(t, v1Manager1.Setup())

	ctx := testhelper.Context(t)

	cmd := exec.CommandContext(ctx, "ls", "-hal", ".")
	require.NoError(t, cmd.Start())
	_, err := v1Manager1.AddCommand(cmd)
	require.NoError(t, err)

	gitCmd1 := exec.CommandContext(ctx, "ls", "-hal", ".")
	require.NoError(t, gitCmd1.Start())
	_, err = v1Manager1.AddCommand(gitCmd1)
	require.NoError(t, err)

	gitCmd2 := exec.CommandContext(ctx, "ls", "-hal", ".")
	require.NoError(t, gitCmd2.Start())
	_, err = v1Manager1.AddCommand(gitCmd2)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, gitCmd2.Wait())
	}()

	require.NoError(t, cmd.Wait())
	require.NoError(t, gitCmd1.Wait())

	repoCgroupPath := filepath.Join(v1Manager1.currentProcessCgroup(), "repos-0")

	expected := strings.NewReader(strings.ReplaceAll(`# HELP gitaly_cgroup_cpu_usage_total CPU Usage of Cgroup
# TYPE gitaly_cgroup_cpu_usage_total gauge
gitaly_cgroup_cpu_usage_total{path="%s",type="kernel"} 0
gitaly_cgroup_cpu_usage_total{path="%s",type="user"} 0
# HELP gitaly_cgroup_memory_reclaim_attempts_total Number of memory usage hits limits
# TYPE gitaly_cgroup_memory_reclaim_attempts_total gauge
gitaly_cgroup_memory_reclaim_attempts_total{path="%s"} 2
# HELP gitaly_cgroup_procs_total Total number of procs
# TYPE gitaly_cgroup_procs_total gauge
gitaly_cgroup_procs_total{path="%s",subsystem="cpu"} 1
gitaly_cgroup_procs_total{path="%s",subsystem="memory"} 1
# HELP gitaly_cgroup_cpu_cfs_periods_total Number of elapsed enforcement period intervals
# TYPE gitaly_cgroup_cpu_cfs_periods_total counter
gitaly_cgroup_cpu_cfs_periods_total{path="%s"} 10
# HELP gitaly_cgroup_cpu_cfs_throttled_periods_total Number of throttled period intervals
# TYPE gitaly_cgroup_cpu_cfs_throttled_periods_total counter
gitaly_cgroup_cpu_cfs_throttled_periods_total{path="%s"} 20
# HELP gitaly_cgroup_cpu_cfs_throttled_seconds_total Total time duration the Cgroup has been throttled
# TYPE gitaly_cgroup_cpu_cfs_throttled_seconds_total counter
gitaly_cgroup_cpu_cfs_throttled_seconds_total{path="%s"} 0.001
`, "%s", repoCgroupPath))

	for _, metricsEnabled := range []bool{true, false} {
		t.Run(fmt.Sprintf("metrics enabled: %v", metricsEnabled), func(t *testing.T) {
			v1Manager1.cfg.MetricsEnabled = metricsEnabled

			if metricsEnabled {
				assert.NoError(t, testutil.CollectAndCompare(
					v1Manager1,
					expected))
			} else {
				assert.NoError(t, testutil.CollectAndCompare(
					v1Manager1,
					bytes.NewBufferString("")))
			}
		})
	}
}

func requireCgroup(t *testing.T, cgroupFile string, want int) {
	t.Helper()

	if want <= 0 {
		// If files doesn't exist kernel will create it with default values
		require.NoFileExistsf(t, cgroupFile, "cgroup file should not exist: %q", cgroupFile)
		return
	}

	require.Equal(t,
		string(readCgroupFile(t, cgroupFile)),
		strconv.Itoa(want),
	)
}

func readCgroupFile(t *testing.T, path string) []byte {
	t.Helper()

	// The cgroups package defaults to permission 0 as it expects the file to be existing (the kernel creates the file)
	// and its testing override the permission private variable to something sensible, hence we have to chmod ourselves
	// so we can read the file.
	require.NoError(t, os.Chmod(path, perm.PublicFile))

	return testhelper.MustReadFile(t, path)
}
