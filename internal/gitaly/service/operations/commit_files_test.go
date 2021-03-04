	"gitlab.com/gitlab-org/gitaly/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/internal/git/localrepo"
	startRepo, _, cleanStartRepo := gittest.InitBareRepo(t)
		treeEntries     []gittest.TreeEntry
		{
			desc: "create file with .git/hooks/pre-commit",
			steps: []step{
				{
					actions: []*gitalypb.UserCommitFilesRequest{
						createFileHeaderRequest(".git/hooks/pre-commit"),
						actionContentRequest("content-1"),
					},
					indexError: "invalid path: '.git/hooks/pre-commit'",
				},
			},
		},
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					treeEntries: []gittest.TreeEntry{
					gittest.InitRepoDir(t, storageRoot, targetRelativePath),
				gittest.RequireTree(t, config.Config.Git.BinPath, repoPath, branch, step.treeEntries)
	repoProto, repoPath, cleanup := gittest.InitBareRepo(t)
	repo := localrepo.New(git.NewExecCommandFactory(config.Config), repoProto, config.Config)
	headerRequest := headerRequest(repoProto, testhelper.TestUser, "master", []byte("commit message"))
	gittest.RequireTree(t, config.Config.Git.BinPath, repoPath, "refs/heads/master", []gittest.TreeEntry{
	commit, err := repo.ReadCommit(ctx, "refs/heads/master")
	testRepo, testRepoPath, cleanup := gittest.CloneRepo(t)
	newRepo, newRepoPath, newRepoCleanupFn := gittest.InitBareRepo(t)
			headCommit, err := localrepo.New(git.NewExecCommandFactory(config.Config), tc.repo, config.Config).ReadCommit(ctx, git.Revision(tc.branchName))
			testRepo, testRepoPath, cleanupFn := gittest.CloneRepo(t)
	repoProto, repoPath, cleanupFn := gittest.CloneRepo(t)
	repo := localrepo.New(git.NewExecCommandFactory(config.Config), repoProto, config.Config)
	startBranchCommit, err := repo.ReadCommit(ctx, git.Revision(startBranchName))
	targetBranchCommit, err := repo.ReadCommit(ctx, git.Revision(targetBranchName))
	mergeBaseOut := testhelper.MustRunCommand(t, nil, "git", "-C", repoPath, "merge-base", targetBranchCommit.Id, startBranchCommit.Id)
	headerRequest := headerRequest(repoProto, testhelper.TestUser, targetBranchName, commitFilesMessage)
	newTargetBranchCommit, err := repo.ReadCommit(ctx, git.Revision(targetBranchName))
	repoProto, _, cleanupFn := gittest.CloneRepo(t)
	repo := localrepo.New(git.NewExecCommandFactory(config.Config), repoProto, config.Config)
	startCommit, err := repo.ReadCommit(ctx, "master")
	headerRequest := headerRequest(repoProto, testhelper.TestUser, targetBranchName, commitFilesMessage)
	newTargetBranchCommit, err := repo.ReadCommit(ctx, git.Revision(targetBranchName))
		gitCmdFactory := git.NewExecCommandFactory(config.Config)

		repoProto, _, cleanupFn := gittest.CloneRepo(t)
		repo := localrepo.New(gitCmdFactory, repoProto, config.Config)
		newRepoProto, _, newRepoCleanupFn := gittest.InitBareRepo(t)
		newRepo := localrepo.New(gitCmdFactory, newRepoProto, config.Config)
		startCommit, err := repo.ReadCommit(ctx, "master")
		headerRequest := headerRequest(newRepoProto, testhelper.TestUser, targetBranchName, commitFilesMessage)
		setStartRepository(headerRequest, repoProto)
		newTargetBranchCommit, err := newRepo.ReadCommit(ctx, git.Revision(targetBranchName))
	repoProto, _, cleanupFn := gittest.InitBareRepo(t)
	repo := localrepo.New(git.NewExecCommandFactory(config.Config), repoProto, config.Config)
			headerRequest := headerRequest(repoProto, tc.user, targetBranchName, commitFilesMessage)
			newCommit, err := repo.ReadCommit(ctx, git.Revision(targetBranchName))
	testRepo, testRepoPath, cleanupFn := gittest.CloneRepo(t)
			remove := gittest.WriteCustomHook(t, testRepoPath, hookName, hookContent)
	testRepo, _, cleanupFn := gittest.CloneRepo(t)
	testRepo, _, cleanupFn := gittest.CloneRepo(t)