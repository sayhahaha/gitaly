package featureflag

type FeatureFlag struct {
	Name        string `json:"name"`
	OnByDefault bool   `json:"on_by_default"`
}

// A set of feature flags used in Gitaly and Praefect.
// In order to support coverage of combined features usage all feature flags should be marked as enabled for the test.
// NOTE: if you add a new feature flag please add it to the `All` list defined below.
var (
	// GoFetchSourceBranch enables a go implementation of FetchSourceBranch
	GoFetchSourceBranch = FeatureFlag{Name: "go_fetch_source_branch", OnByDefault: false}
	// DistributedReads allows praefect to redirect accessor operations to up-to-date secondaries
	DistributedReads = FeatureFlag{Name: "distributed_reads", OnByDefault: false}
	// LogCommandStats will log additional rusage stats for commands
	LogCommandStats = FeatureFlag{Name: "log_command_stats", OnByDefault: false}
	// GoUserMergeBranch enables the Go implementation of UserMergeBranch
	GoUserMergeBranch = FeatureFlag{Name: "go_user_merge_branch", OnByDefault: false}
	// GoUserMergeToRef enable the Go implementation of UserMergeToRef
	GoUserMergeToRef = FeatureFlag{Name: "go_user_merge_to_ref", OnByDefault: true}
	// GoUserFFBranch enables the Go implementation of UserFFBranch
	GoUserFFBranch = FeatureFlag{Name: "go_user_ff_branch", OnByDefault: false}
	// GoUserCreateBranch enables the Go implementation of UserCreateBranch
	GoUserCreateBranch = FeatureFlag{Name: "go_user_create_branch", OnByDefault: false}
	// GoUserDeleteBranch enables the Go implementation of UserDeleteBranch
	GoUserDeleteBranch = FeatureFlag{Name: "go_user_delete_branch", OnByDefault: false}
	// GoUserSquash enables the Go implementation of UserSquash
	GoUserSquash = FeatureFlag{Name: "go_user_squash", OnByDefault: true}
	// GoListConflictFiles enables the Go implementation of ListConflictFiles
	GoListConflictFiles = FeatureFlag{Name: "go_list_conflict_files", OnByDefault: true}
	// GoUserCommitFiles enables the Go implementation of UserCommitFiles
	GoUserCommitFiles = FeatureFlag{Name: "go_user_commit_files", OnByDefault: false}
	// GoResolveConflicts enables the Go implementation of ResolveConflicts
	GoResolveConflicts = FeatureFlag{Name: "go_resolve_conflicts", OnByDefault: false}
	// GoUserUpdateSubmodule enables the Go implementation of
	// UserUpdateSubmodules
	GoUserUpdateSubmodule = FeatureFlag{Name: "go_user_update_submodule", OnByDefault: false}
	// GoFetchRemote enables the Go implementation of FetchRemote
	GoFetchRemote = FeatureFlag{Name: "go_fetch_remote", OnByDefault: true}
	// GoUserDeleteTag enables the Go implementation of UserDeleteTag
	GoUserDeleteTag = FeatureFlag{Name: "go_user_delete_tag", OnByDefault: false}
	// GoUserRevert enables the Go implementation of UserRevert
	GoUserRevert = FeatureFlag{Name: "go_user_revert", OnByDefault: false}
)

// All includes all feature flags.
var All = []FeatureFlag{
	GoFetchSourceBranch,
	DistributedReads,
	LogCommandStats,
	GoUserMergeBranch,
	GoUserMergeToRef,
	GoUserFFBranch,
	GoUserCreateBranch,
	GoUserDeleteBranch,
	GoUserSquash,
	GoListConflictFiles,
	GoUserCommitFiles,
	GoResolveConflicts,
	GoUserUpdateSubmodule,
	GoFetchRemote,
	GoUserDeleteTag,
	GoUserRevert,
}
