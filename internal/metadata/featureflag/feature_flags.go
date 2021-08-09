package featureflag

// A set of feature flags used in Gitaly and Praefect.
// In order to support coverage of combined features usage all feature flags should be marked as enabled for the test.
// NOTE: if you add a new feature flag please add it to the `All` list defined below.
var (
	// GoSetConfig enables git2go implementation of SetConfig.
	GoSetConfig = FeatureFlag{Name: "go_set_config", OnByDefault: false}
	// FindAllTagsPipeline enables the alternative pipeline implementation for finding
	// tags via FindAllTags.
	FindAllTagsPipeline = FeatureFlag{Name: "find_all_tags_pipeline", OnByDefault: false}
	// QuarantinedUserCreateTag enables use of a quarantine object directory for UserCreateTag.
	QuarantinedUserCreateTag = FeatureFlag{Name: "quarantined_user_create_tag", OnByDefault: false}
	// UserSquashWithoutWorktree enables the new implementation of UserSquash which does not
	// require a worktree to compute the squash.
	UserSquashWithoutWorktree = FeatureFlag{Name: "user_squash_without_worktree", OnByDefault: false}
	// QuarantinedResolveCOnflicts enables use of a quarantine object directory for ResolveConflicts.
	QuarantinedResolveConflicts = FeatureFlag{Name: "quarantined_resolve_conflicts", OnByDefault: false}
	// GoUserApplyPatch enables the Go implementation of UserApplyPatch
	GoUserApplyPatch = FeatureFlag{Name: "go_user_apply_patch", OnByDefault: false}
	// FetchInternalNoAlternateRefs disables use of alternate refs in internal fetches.
	FetchInternalNoAlternateRefs = FeatureFlag{Name: "fetch_internal_no_alternate_refs", OnByDefault: false}
	// Quarantine enables the use of quarantine directories.
	Quarantine = FeatureFlag{Name: "quarantine", OnByDefault: false}
)

// All includes all feature flags.
var All = []FeatureFlag{
	GoSetConfig,
	FindAllTagsPipeline,
	QuarantinedUserCreateTag,
	UserSquashWithoutWorktree,
	QuarantinedResolveConflicts,
	GoUserApplyPatch,
	FetchInternalNoAlternateRefs,
	Quarantine,
}
