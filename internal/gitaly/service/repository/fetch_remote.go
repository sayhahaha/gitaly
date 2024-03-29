package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/localrepo"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/transaction"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/internal/transaction/txinfo"
	"gitlab.com/gitlab-org/gitaly/v16/internal/transaction/voting"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

func (s *server) FetchRemote(ctx context.Context, req *gitalypb.FetchRemoteRequest) (*gitalypb.FetchRemoteResponse, error) {
	if err := s.validateFetchRemoteRequest(req); err != nil {
		return nil, err
	}

	var stderr bytes.Buffer
	opts := localrepo.FetchOpts{
		Stderr:              &stderr,
		Force:               req.Force,
		Prune:               !req.NoPrune,
		Tags:                localrepo.FetchOptsTagsAll,
		Verbose:             req.GetCheckTagsChanged(),
		DisableTransactions: true,
	}

	if req.GetNoTags() {
		opts.Tags = localrepo.FetchOptsTagsNone
	}

	repo := s.localrepo(req.GetRepository())
	remoteName := "inmemory"
	remoteURL := req.GetRemoteParams().GetUrl()
	var config []git.ConfigPair

	for _, refspec := range s.getRefspecs(req.GetRemoteParams().GetMirrorRefmaps()) {
		config = append(config, git.ConfigPair{
			Key: "remote.inmemory.fetch", Value: refspec,
		})
	}

	if resolvedAddress := req.GetRemoteParams().GetResolvedAddress(); resolvedAddress != "" {
		modifiedURL, resolveConfig, err := git.GetURLAndResolveConfig(remoteURL, resolvedAddress)
		if err != nil {
			return nil, fmt.Errorf("couldn't get curloptResolve config: %w", err)
		}

		remoteURL = modifiedURL
		config = append(config, resolveConfig...)
	}

	config = append(config, git.ConfigPair{Key: "remote.inmemory.url", Value: remoteURL})

	if authHeader := req.GetRemoteParams().GetHttpAuthorizationHeader(); authHeader != "" {
		config = append(config, git.ConfigPair{
			Key:   fmt.Sprintf("http.%s.extraHeader", req.GetRemoteParams().GetUrl()),
			Value: "Authorization: " + authHeader,
		})
	}

	opts.CommandOptions = append(opts.CommandOptions, git.WithConfigEnv(config...))

	sshCommand, cleanup, err := git.BuildSSHInvocation(ctx, req.GetSshKey(), req.GetKnownHosts())
	if err != nil {
		return nil, err
	}
	defer cleanup()

	opts.Env = append(opts.Env, "GIT_SSH_COMMAND="+sshCommand)

	if req.GetTimeout() > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.GetTimeout())*time.Second)
		defer cancel()
	}

	if err := repo.FetchRemote(ctx, remoteName, opts); err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return nil, structerr.NewInternal("fetch remote: %q: %w", errMsg, err)
		}

		return nil, structerr.NewInternal("fetch remote: %w", err)
	}

	// Ideally, we'd do the voting process via git-fetch(1) using the reference-transaction
	// hook. But by default this would lead to one hook invocation per updated ref, which is
	// infeasible performance-wise. While this could be fixed via the `--atomic` flag, that's
	// not a solution either: we rely on the fact that refs get updated even if a subset of refs
	// diverged, and with atomic transactions it would instead be an all-or-nothing operation.
	//
	// Instead, we do the second-best thing, which is to vote on the resulting references. This
	// is of course racy and may conflict with other mutators, causing the vote to fail. But it
	// is arguably preferable to accept races in favour always replicating. If loosing the race,
	// we'd fail this RPC and schedule a replication job afterwards.
	if err := transaction.RunOnContext(ctx, func(tx txinfo.Transaction) error {
		hash := voting.NewVoteHash()

		if err := repo.ExecAndWait(ctx, git.Command{
			Name: "for-each-ref",
		}, git.WithStdout(hash)); err != nil {
			return fmt.Errorf("cannot compute references vote: %w", err)
		}

		vote, err := hash.Vote()
		if err != nil {
			return err
		}

		return s.txManager.Vote(ctx, tx, vote, voting.UnknownPhase)
	}); err != nil {
		return nil, structerr.NewAborted("failed vote on refs: %w", err)
	}

	out := &gitalypb.FetchRemoteResponse{TagsChanged: true}
	if req.GetCheckTagsChanged() {
		out.TagsChanged = didTagsChange(&stderr)
	}

	return out, nil
}

func didTagsChange(r io.Reader) bool {
	scanner := git.NewFetchScanner(r)
	for scanner.Scan() {
		status := scanner.StatusLine()

		// We can't detect if tags have been deleted, but we never call fetch
		// with --prune-tags at the moment, so it should never happen.
		if status.IsTagAdded() || status.IsTagUpdated() {
			return true
		}
	}

	// If the scanner fails for some reason, we don't know if tags changed, so
	// assume they did for safety reasons.
	return scanner.Err() != nil
}

func (s *server) validateFetchRemoteRequest(req *gitalypb.FetchRemoteRequest) error {
	if err := s.locator.ValidateRepository(req.GetRepository()); err != nil {
		return structerr.NewInvalidArgument("%w", err)
	}

	if req.GetRemoteParams() == nil {
		return structerr.NewInvalidArgument("missing remote params")
	}

	if req.GetRemoteParams().GetUrl() == "" {
		return structerr.NewInvalidArgument("blank or empty remote URL")
	}

	return nil
}

func (s *server) getRefspecs(refmaps []string) []string {
	if len(refmaps) == 0 {
		return []string{"refs/*:refs/*"}
	}

	refspecs := make([]string, 0, len(refmaps))

	for _, refmap := range refmaps {
		switch refmap {
		case "all_refs":
			// with `all_refs`, the repository is equivalent to the result of `git clone --mirror`
			refspecs = append(refspecs, "refs/*:refs/*")
		case "heads":
			refspecs = append(refspecs, "refs/heads/*:refs/heads/*")
		case "tags":
			refspecs = append(refspecs, "refs/tags/*:refs/tags/*")
		default:
			refspecs = append(refspecs, refmap)
		}
	}
	return refspecs
}
