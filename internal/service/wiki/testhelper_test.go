package wiki

import (
	"bytes"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	gitlog "gitlab.com/gitlab-org/gitaly/internal/git/log"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	mockPageContent = bytes.Repeat([]byte("Mock wiki page content"), 10000)
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

var rubyServer *rubyserver.Server

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	var err error
	testhelper.ConfigureRuby()
	rubyServer, err = rubyserver.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer rubyServer.Stop()

	return m.Run()
}

func runWikiServiceServer(t *testing.T) (*grpc.Server, string) {
	grpcServer := testhelper.NewTestGrpcServer(t, nil, nil)
	serverSocketPath := testhelper.GetTemporaryGitalySocketFileName()

	listener, err := net.Listen("unix", serverSocketPath)
	if err != nil {
		t.Fatal(err)
	}

	pb.RegisterWikiServiceServer(grpcServer, &server{rubyServer})
	reflection.Register(grpcServer)

	go grpcServer.Serve(listener)

	return grpcServer, serverSocketPath
}

func newWikiClient(t *testing.T, serverSocketPath string) (pb.WikiServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	}
	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return pb.NewWikiServiceClient(conn), conn
}

func writeWikiPage(t *testing.T, client pb.WikiServiceClient, wikiRepo *pb.Repository, name string, content []byte) {
	commitDetails := &pb.WikiCommitDetails{
		Name:    []byte("Ahmad Sherif"),
		Email:   []byte("ahmad@gitlab.com"),
		Message: []byte("Add " + name),
	}

	request := &pb.WikiWritePageRequest{
		Repository:    wikiRepo,
		Name:          []byte(name),
		Format:        "markdown",
		CommitDetails: commitDetails,
		Content:       content,
	}

	ctx, cancel := testhelper.Context()
	defer cancel()

	stream, err := client.WikiWritePage(ctx)
	require.NoError(t, err)

	require.NoError(t, stream.Send(request))

	_, err = stream.CloseAndRecv()
	require.NoError(t, err)
}

func updateWikiPage(t *testing.T, client pb.WikiServiceClient, wikiRepo *pb.Repository, name string, content []byte) {
	commitDetails := &pb.WikiCommitDetails{
		Name:    []byte("Ahmad Sherif"),
		Email:   []byte("ahmad@gitlab.com"),
		Message: []byte("Update " + name),
	}

	request := &pb.WikiUpdatePageRequest{
		Repository:    wikiRepo,
		PagePath:      []byte(name),
		Title:         []byte(name),
		Format:        "markdown",
		CommitDetails: commitDetails,
		Content:       content,
	}

	ctx, cancel := testhelper.Context()
	defer cancel()

	stream, err := client.WikiUpdatePage(ctx)
	require.NoError(t, err)
	require.NoError(t, stream.Send(request))

	_, err = stream.CloseAndRecv()
	require.NoError(t, err)
}

func setupWikiRepo(t *testing.T) (*pb.Repository, string, func()) {
	testhelper.ConfigureTestStorage()
	relPath := strings.Join([]string{t.Name(), "wiki-test.git"}, "-")
	storagePath := testhelper.GitlabTestStoragePath()
	wikiRepoPath := path.Join(storagePath, relPath)

	testhelper.MustRunCommand(nil, nil, "git", "init", "--bare", wikiRepoPath)

	wikiRepo := &pb.Repository{
		StorageName:  "default",
		RelativePath: relPath,
	}

	return wikiRepo, wikiRepoPath, func() { os.RemoveAll(wikiRepoPath) }
}

func sendBytes(data []byte, chunkSize int, sender func([]byte) error) (int, error) {
	i := 0
	for ; len(data) > 0; i++ {
		n := chunkSize
		if n > len(data) {
			n = len(data)
		}

		if err := sender(data[:n]); err != nil {
			return i, err
		}
		data = data[n:]
	}

	return i, nil
}

func createTestWikiPage(t *testing.T, client pb.WikiServiceClient, wikiRepo *pb.Repository, pageName string) *pb.GitCommit {
	ctx, cancel := testhelper.Context()
	defer cancel()

	wikiRepoPath, err := helper.GetRepoPath(wikiRepo)
	require.NoError(t, err)

	writeWikiPage(t, client, wikiRepo, pageName, mockPageContent)
	head1ID := testhelper.MustRunCommand(t, nil, "git", "-C", wikiRepoPath, "show", "--format=format:%H", "--no-patch", "HEAD")
	pageCommit, err := gitlog.GetCommit(ctx, wikiRepo, string(head1ID), "")
	require.NoError(t, err, "look up git commit after writing a wiki page")

	return pageCommit
}

func requireWikiPagesEqual(t *testing.T, expectedPage *pb.WikiPage, actualPage *pb.WikiPage) {
	// require.Equal doesn't display a proper diff when either expected/actual has a field
	// with large data (RawData in our case), so we compare file attributes and content separately.
	expectedContent := expectedPage.GetRawData()
	if expectedPage != nil {
		expectedPage.RawData = nil
		defer func() {
			expectedPage.RawData = expectedContent
		}()
	}
	actualContent := actualPage.GetRawData()
	if actualPage != nil {
		actualPage.RawData = nil
		defer func() {
			actualPage.RawData = actualContent
		}()
	}

	require.Equal(t, expectedPage, actualPage, "mismatched page attributes")
	if expectedPage != nil {
		require.Equal(t, expectedContent, actualContent, "mismatched page content")
	}
}
