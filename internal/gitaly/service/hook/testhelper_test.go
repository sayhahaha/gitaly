package hook

import (
	"testing"

	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/config"
	gitalyhook "gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/hook"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service/repository"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/transaction"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v16/internal/transaction/voting"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestMain(m *testing.M) {
	testhelper.Run(m)
}

func setupHookService(tb testing.TB) (config.Cfg, gitalypb.HookServiceClient) {
	tb.Helper()

	cfg := testcfg.Build(tb)
	cfg.SocketPath = runHooksServer(tb, cfg, nil)
	client, conn := newHooksClient(tb, cfg.SocketPath)
	tb.Cleanup(func() { conn.Close() })

	return cfg, client
}

func newHooksClient(tb testing.TB, serverSocketPath string) (gitalypb.HookServiceClient, *grpc.ClientConn) {
	tb.Helper()

	connOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		tb.Fatal(err)
	}

	return gitalypb.NewHookServiceClient(conn), conn
}

type serverOption func(*server)

func runHooksServer(tb testing.TB, cfg config.Cfg, opts []serverOption, serverOpts ...testserver.GitalyServerOpt) string {
	tb.Helper()

	serverOpts = append(serverOpts, testserver.WithDisablePraefect())

	return testserver.RunGitalyServer(tb, cfg, func(srv *grpc.Server, deps *service.Dependencies) {
		hookServer := NewServer(
			gitalyhook.NewManager(deps.GetCfg(), deps.GetLocator(), deps.GetGitCmdFactory(), deps.GetTxManager(), deps.GetGitlabClient()),
			deps.GetLocator(),
			deps.GetGitCmdFactory(),
			deps.GetPackObjectsCache(),
			deps.GetPackObjectsLimiter(),
		)
		for _, opt := range opts {
			opt(hookServer.(*server))
		}

		gitalypb.RegisterHookServiceServer(srv, hookServer)
		gitalypb.RegisterRepositoryServiceServer(srv, repository.NewServer(
			cfg,
			deps.GetLocator(),
			deps.GetTxManager(),
			deps.GetGitCmdFactory(),
			deps.GetCatfileCache(),
			deps.GetConnsPool(),
			deps.GetHousekeepingManager(),
			deps.GetBackupSink(),
			deps.GetBackupLocator(),
			deps.GetRepositoryCounter(),
		))
	}, serverOpts...)
}

func synchronizedVote(hook string) transaction.PhasedVote {
	return transaction.PhasedVote{
		Vote:  voting.VoteFromData([]byte("synchronize " + hook + " hook")),
		Phase: voting.Synchronized,
	}
}
