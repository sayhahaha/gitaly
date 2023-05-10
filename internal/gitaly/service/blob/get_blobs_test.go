package blob

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

func TestGetBlobs(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg, client := setup(t, ctx)

	repo, repoPath := gittest.CreateRepository(t, ctx, cfg)

	blobASize := int64(13)
	blobAData := strings.Repeat("a", int(blobASize))
	blobAOID := gittest.WriteBlob(t, cfg, repoPath, []byte(blobAData))

	blobBSize := int64(50)
	blobBData := strings.Repeat("b", int(blobBSize))
	blobBOID := gittest.WriteBlob(t, cfg, repoPath, []byte(blobBData))

	commitID := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
		gittest.TreeEntry{OID: blobAOID, Path: "a", Mode: "100644"},
		gittest.TreeEntry{OID: blobBOID, Path: "b", Mode: "100644"},
	))

	missingBlob := func(path string, revision git.ObjectID) *gitalypb.GetBlobsResponse {
		return &gitalypb.GetBlobsResponse{
			Path:     []byte(path),
			Revision: revision.String(),
		}
	}

	existingBlob := func(path string, revision, blobID git.ObjectID, data string, size int64) *gitalypb.GetBlobsResponse {
		return &gitalypb.GetBlobsResponse{
			Path:     []byte(path),
			Revision: revision.String(),
			Oid:      blobID.String(),
			Data:     []byte(data),
			Size:     size,
			Mode:     0o100644,
			Type:     gitalypb.ObjectType_BLOB,
		}
	}

	type setupData struct {
		request           *gitalypb.GetBlobsRequest
		expectedErr       error
		expectedResponses []*gitalypb.GetBlobsResponse
	}

	for _, tc := range []struct {
		desc  string
		setup func(t *testing.T) setupData
	}{
		{
			desc: "empty Repository",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: "does-not-exist", Path: []byte("file")},
						},
					},
					expectedErr: structerr.NewInvalidArgument(testhelper.GitalyOrPraefect(
						"empty Repository",
						"repo scoped: empty Repository",
					)),
				}
			},
		},
		{
			desc: "empty RevisionPaths",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
					},
					expectedErr: structerr.NewInvalidArgument("empty RevisionPaths"),
				}
			},
		},
		{
			desc: "invalid Revision",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: "--output=/meow", Path: []byte("CHANGELOG")},
						},
					},
					expectedErr: structerr.NewInvalidArgument("revision can't start with '-'"),
				}
			},
		},
		{
			desc: "zero-limit skips blob data",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
						Limit:      0,
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: commitID.String(), Path: []byte("a")},
							{Revision: commitID.String(), Path: []byte("b")},
						},
					},
					expectedResponses: []*gitalypb.GetBlobsResponse{
						existingBlob("a", commitID, blobAOID, "", blobASize),
						existingBlob("b", commitID, blobBOID, "", blobBSize),
					},
				}
			},
		},
		{
			desc: "negative limit sends all blob data",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
						Limit:      -1,
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: commitID.String(), Path: []byte("a")},
							{Revision: commitID.String(), Path: []byte("b")},
						},
					},
					expectedResponses: []*gitalypb.GetBlobsResponse{
						existingBlob("a", commitID, blobAOID, blobAData, blobASize),
						existingBlob("b", commitID, blobBOID, blobBData, blobBSize),
					},
				}
			},
		},
		{
			desc: "positive limit truncates blob data",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
						Limit:      5,
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: commitID.String(), Path: []byte("a")},
							{Revision: commitID.String(), Path: []byte("b")},
						},
					},
					expectedResponses: []*gitalypb.GetBlobsResponse{
						existingBlob("a", commitID, blobAOID, blobAData[:5], blobASize),
						existingBlob("b", commitID, blobBOID, blobBData[:5], blobBSize),
					},
				}
			},
		},
		{
			desc: "with limit exceeding blob sizes",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
						Limit:      9000,
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: commitID.String(), Path: []byte("a")},
							{Revision: commitID.String(), Path: []byte("b")},
						},
					},
					expectedResponses: []*gitalypb.GetBlobsResponse{
						existingBlob("a", commitID, blobAOID, blobAData, blobASize),
						existingBlob("b", commitID, blobBOID, blobBData, blobBSize),
					},
				}
			},
		},
		{
			desc: "nonexisting path",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: commitID.String(), Path: []byte("does-not-exist")},
						},
					},
					expectedResponses: []*gitalypb.GetBlobsResponse{
						missingBlob("does-not-exist", commitID),
					},
				}
			},
		},
		{
			desc: "mixed existing and nonexisting path",
			setup: func(t *testing.T) setupData {
				return setupData{
					request: &gitalypb.GetBlobsRequest{
						Repository: repo,
						RevisionPaths: []*gitalypb.GetBlobsRequest_RevisionPath{
							{Revision: commitID.String(), Path: []byte("a")},
							{Revision: commitID.String(), Path: []byte("does-not-exist")},
						},
					},
					expectedResponses: []*gitalypb.GetBlobsResponse{
						existingBlob("a", commitID, blobAOID, "", blobASize),
						missingBlob("does-not-exist", commitID),
					},
				}
			},
		},
	} {
		tc := tc

		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			setup := tc.setup(t)

			stream, err := client.GetBlobs(ctx, setup.request)
			require.NoError(t, err)

			var responses []*gitalypb.GetBlobsResponse
			for {
				response, err := stream.Recv()
				if err != nil {
					if setup.expectedErr == nil {
						setup.expectedErr = io.EOF
					}

					testhelper.RequireGrpcError(t, setup.expectedErr, err)
					break
				}
				require.NoError(t, err)

				responses = append(responses, response)
			}

			testhelper.ProtoEqual(t, setup.expectedResponses, responses)
		})
	}
}
