package catfile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/featureflag"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testcfg"
)

func TestObjectContentReader_reader(t *testing.T) {
	t.Parallel()

	testhelper.NewFeatureSets(featureflag.MailmapOptions).Run(t, testObjectContentReader)
}

func testObjectContentReader(t *testing.T, ctx context.Context) {
	cfg := testcfg.Build(t)
	repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
		SkipCreationViaService: true,
	})

	commitID := gittest.WriteCommit(t, cfg, repoPath,
		gittest.WithBranch("main"),
		gittest.WithMessage("commit message"),
		gittest.WithTreeEntries(gittest.TreeEntry{Path: "README", Mode: "100644", Content: "something"}),
	)
	gittest.WriteTag(t, cfg, repoPath, "v1.1.1", commitID.Revision(), gittest.WriteTagConfig{
		Message: "annotated tag",
	})

	commitContents := gittest.Exec(t, cfg, "-C", repoPath, "cat-file", "-p", commitID.String())

	t.Run("read existing object by ref", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		object, err := reader.Object(ctx, "refs/heads/main")
		require.NoError(t, err)

		data, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, commitContents, data)
	})

	t.Run("read existing object by object ID", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		object, err := reader.Object(ctx, commitID.Revision())
		require.NoError(t, err)

		data, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, data, commitContents)
	})

	t.Run("read missing ref", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		_, err = reader.Object(ctx, "refs/heads/does-not-exist")
		require.EqualError(t, err, "object not found")

		// Verify that we're still able to read a commit after the previous read has failed.
		object, err := reader.Object(ctx, commitID.Revision())
		require.NoError(t, err)

		data, err := io.ReadAll(object)
		require.NoError(t, err)

		require.Equal(t, commitContents, data)
	})

	t.Run("read fails when not consuming previous object", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		_, err = reader.Object(ctx, commitID.Revision())
		require.NoError(t, err)

		// We haven't yet consumed the previous object, so this must now fail.
		_, err = reader.Object(ctx, commitID.Revision())
		require.EqualError(t, err, "current object has not been fully read")
	})

	t.Run("read fails when partially consuming previous object", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		object, err := reader.Object(ctx, commitID.Revision())
		require.NoError(t, err)

		_, err = io.CopyN(io.Discard, object, 100)
		require.NoError(t, err)

		// We haven't yet consumed the previous object, so this must now fail.
		_, err = reader.Object(ctx, commitID.Revision())
		require.EqualError(t, err, "current object has not been fully read")
	})

	t.Run("read increments Prometheus counter", func(t *testing.T) {
		counter := prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"type"})

		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), counter)
		require.NoError(t, err)

		for objectType, revision := range map[string]git.Revision{
			"commit": "refs/heads/main",
			"tree":   "refs/heads/main^{tree}",
			"blob":   "refs/heads/main:README",
			"tag":    "refs/tags/v1.1.1",
		} {
			require.Equal(t, float64(0), testutil.ToFloat64(counter.WithLabelValues(objectType)))

			object, err := reader.Object(ctx, revision)
			require.NoError(t, err)

			require.Equal(t, float64(1), testutil.ToFloat64(counter.WithLabelValues(objectType)))

			_, err = io.Copy(io.Discard, object)
			require.NoError(t, err)
		}
	})

	t.Run("read existing object with mailmap", func(t *testing.T) {
		mailmapContents := "A U Thor <author@example.com> Scrooge McDuck <scrooge@mcduck.com>"

		commitID := gittest.WriteCommit(t, cfg, repoPath,
			gittest.WithTreeEntries(
				gittest.TreeEntry{Path: ".mailmap", Mode: "100644", Content: mailmapContents},
			),
			gittest.WithBranch("main"),
		)

		var commitContents []byte

		if featureflag.MailmapOptions.IsEnabled(ctx) {
			commitContents = gittest.Exec(t, cfg, "-C", repoPath, "cat-file",
				"--use-mailmap", "-p", commitID.String())
		} else {
			commitContents = gittest.Exec(t, cfg, "-C", repoPath, "cat-file",
				"-p", commitID.String())
		}

		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		object, err := reader.Object(ctx, "refs/heads/main")
		require.NoError(t, err)

		data, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, commitContents, data)
	})
}

func TestObjectContentReader_queue(t *testing.T) {
	ctx := testhelper.Context(t)

	cfg := testcfg.Build(t)
	repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
		SkipCreationViaService: true,
	})

	foobarBlob := gittest.WriteBlob(t, cfg, repoPath, []byte("foobar"))
	barfooBlob := gittest.WriteBlob(t, cfg, repoPath, []byte("barfoo"))
	treeWithNewlines := gittest.WriteTree(t, cfg, repoPath, []gittest.TreeEntry{
		{Path: "path\nwith\nnewline", Mode: "100644", OID: foobarBlob},
	})

	t.Run("reader is dirty with acquired queue", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.False(t, queue.isDirty())
		require.True(t, reader.isDirty())
	})

	t.Run("read single object", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))

		object, err := queue.ReadObject(ctx)
		require.NoError(t, err)

		contents, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, "foobar", string(contents))
	})

	t.Run("read multiple objects", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		for blobID, blobContents := range map[git.ObjectID]string{
			foobarBlob: "foobar",
			barfooBlob: "barfoo",
		} {
			require.NoError(t, queue.RequestObject(ctx, blobID.Revision()))
			require.NoError(t, queue.Flush(ctx))

			object, err := queue.ReadObject(ctx)
			require.NoError(t, err)

			contents, err := io.ReadAll(object)
			require.NoError(t, err)
			require.Equal(t, blobContents, string(contents))
		}
	})

	t.Run("request multiple objects", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.RequestObject(ctx, barfooBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))

		for _, expectedContents := range []string{"foobar", "barfoo"} {
			object, err := queue.ReadObject(ctx)
			require.NoError(t, err)

			contents, err := io.ReadAll(object)
			require.NoError(t, err)
			require.Equal(t, expectedContents, string(contents))
		}
	})

	t.Run("read without request", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		_, err = queue.ReadObject(ctx)
		require.Equal(t, errors.New("no outstanding request"), err)
	})

	t.Run("flush with single request", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		// We flush once before and once after requesting the object such that we can be
		// sure that it doesn't impact which objects we can read.
		require.NoError(t, queue.Flush(ctx))
		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))

		object, err := queue.ReadObject(ctx)
		require.NoError(t, err)

		contents, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, "foobar", string(contents))
	})

	t.Run("flush with multiple requests", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		for i := 0; i < 10; i++ {
			require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		}
		require.NoError(t, queue.Flush(ctx))

		for i := 0; i < 10; i++ {
			object, err := queue.ReadObject(ctx)
			require.NoError(t, err)

			contents, err := io.ReadAll(object)
			require.NoError(t, err)
			require.Equal(t, "foobar", string(contents))
		}
	})

	t.Run("flush without request", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.NoError(t, queue.Flush(ctx))

		_, err = queue.ReadObject(ctx)
		require.Equal(t, errors.New("no outstanding request"), err)
	})

	t.Run("request invalid object", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.NoError(t, queue.RequestObject(ctx, "does-not-exist"))
		require.NoError(t, queue.Flush(ctx))

		_, err = queue.ReadObject(ctx)
		require.Equal(t, NotFoundError{"does-not-exist"}, err)
	})

	t.Run("reading object with newline", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		err = queue.RequestObject(ctx, treeWithNewlines.Revision()+":path\nwith\nnewline")
		if !catfileSupportsNul(t, ctx, cfg) {
			require.Equal(t, structerr.NewInvalidArgument("Git too old to support requests with newlines"), err)
			return
		}
		require.NoError(t, err)
		require.NoError(t, queue.Flush(ctx))

		object, err := queue.ReadObject(ctx)
		require.NoError(t, err)

		contents, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, "foobar", string(contents))
	})

	t.Run("can continue reading after NotFoundError", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.NoError(t, queue.RequestObject(ctx, "does-not-exist"))
		require.NoError(t, queue.Flush(ctx))

		_, err = queue.ReadObject(ctx)
		require.Equal(t, NotFoundError{"does-not-exist"}, err)

		// Requesting another object after the previous one has failed should continue to
		// work alright.
		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))
		object, err := queue.ReadObject(ctx)
		require.NoError(t, err)

		contents, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, "foobar", string(contents))
	})

	t.Run("missing object with newline", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		err = queue.RequestObject(ctx, "does\nnot\nexist")
		if !catfileSupportsNul(t, ctx, cfg) {
			require.Equal(t, structerr.NewInvalidArgument("Git too old to support requests with newlines"), err)
			return
		}
		require.NoError(t, err)
		require.NoError(t, queue.Flush(ctx))

		_, err = queue.ReadObject(ctx)
		require.Equal(t, NotFoundError{"does\nnot\nexist"}, err)

		// Requesting another object after the previous one has failed should continue to
		// work alright.
		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))
		object, err := queue.ReadObject(ctx)
		require.NoError(t, err)

		contents, err := io.ReadAll(object)
		require.NoError(t, err)
		require.Equal(t, "foobar", string(contents))
	})

	t.Run("requesting multiple queues fails", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		_, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		_, _, err = reader.objectQueue(ctx, "trace")
		require.Equal(t, errors.New("object queue already in use"), err)

		// After calling cleanup we should be able to create an object queue again.
		cleanup()

		_, cleanup, err = reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()
	})

	t.Run("requesting object dirties reader", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		require.False(t, reader.isDirty())

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.True(t, reader.isDirty())
		require.False(t, queue.isDirty())

		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))

		require.True(t, reader.isDirty())
		require.True(t, queue.isDirty())

		object, err := queue.ReadObject(ctx)
		require.NoError(t, err)

		// The object has not been consumed yet, so the reader must still be dirty.
		require.True(t, reader.isDirty())
		require.True(t, queue.isDirty())

		_, err = io.ReadAll(object)
		require.NoError(t, err)

		require.True(t, reader.isDirty())
		require.False(t, queue.isDirty())

		cleanup()

		require.False(t, reader.isDirty())
		require.False(t, queue.isDirty())
	})

	t.Run("closing queue blocks request", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		queue.close()

		require.True(t, reader.isClosed())
		require.True(t, queue.isClosed())

		require.Equal(t, fmt.Errorf("cannot request revision: %w", os.ErrClosed), queue.RequestObject(ctx, foobarBlob.Revision()))
	})

	t.Run("closing queue blocks read", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		// Request the object before we close the queue.
		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))

		queue.close()

		require.True(t, reader.isClosed())
		require.True(t, queue.isClosed())

		_, err = queue.ReadObject(ctx)
		require.Equal(t, fmt.Errorf("cannot read object info: %w", os.ErrClosed), err)
	})

	t.Run("closing queue blocks consuming", func(t *testing.T) {
		reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
		require.NoError(t, err)

		queue, cleanup, err := reader.objectQueue(ctx, "trace")
		require.NoError(t, err)
		defer cleanup()

		require.NoError(t, queue.RequestObject(ctx, foobarBlob.Revision()))
		require.NoError(t, queue.Flush(ctx))

		// Read the object header before closing.
		object, err := queue.ReadObject(ctx)
		require.NoError(t, err)

		queue.close()

		require.True(t, reader.isClosed())
		require.True(t, queue.isClosed())

		_, err = io.ReadAll(object)
		require.Equal(t, os.ErrClosed, err)
	})
}

func TestObjectContentReader_replaceRefs(t *testing.T) {
	ctx := testhelper.Context(t)

	cfg := testcfg.Build(t)
	repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
		SkipCreationViaService: true,
	})

	originalOID := gittest.WriteBlob(t, cfg, repoPath, []byte("original"))
	replacedOID := gittest.WriteBlob(t, cfg, repoPath, []byte("replaced"))

	gittest.WriteRef(t, cfg, repoPath, git.ReferenceName("refs/replace/"+originalOID.String()), replacedOID)

	// Reading the object via our testhelper should result in the object having been replaced.
	require.Equal(t, "replaced", text.ChompBytes(
		gittest.Exec(t, cfg, "-c", "core.useReplaceRefs=true", "-C", repoPath, "cat-file", "-p", originalOID.String()),
	))

	reader, err := newObjectContentReader(ctx, newRepoExecutor(t, cfg, repoProto), nil)
	require.NoError(t, err)

	object, err := reader.Object(ctx, originalOID.Revision())
	require.NoError(t, err)

	contents, err := io.ReadAll(object)
	require.NoError(t, err)

	// But using our "normal" Git command execution code path, we still want to see the original
	// content of the blob.
	require.Equal(t, "original", string(contents))
}
