package repocleaner

import (
	"context"

	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
)

// LogWarnAction is an implementation of the Action interface that allows to log a warning message
// for the repositories that are not known for the praefect.
type LogWarnAction struct {
	logger log.Logger
}

// NewLogWarnAction return new instance of the LogWarnAction.
func NewLogWarnAction(logger log.Logger) *LogWarnAction {
	return &LogWarnAction{
		logger: logger.WithField("component", "repocleaner.log_warn_action"),
	}
}

// Perform logs a warning for each repository that is not known to praefect.
func (al LogWarnAction) Perform(_ context.Context, virtualStorage, storage string, replicaPaths []string) error {
	for _, replicaPath := range replicaPaths {
		al.logger.WithFields(log.Fields{
			"virtual_storage":       virtualStorage,
			"storage":               storage,
			"relative_replica_path": replicaPath,
		}).Warn("repository is not managed by praefect")
	}
	return nil
}
