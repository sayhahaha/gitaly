package praefect

import (
	"errors"
	"fmt"

	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"google.golang.org/grpc"
)

func validateRenameRepositoryRequest(req *gitalypb.RenameRepositoryRequest, virtualStorages map[string]struct{}) error {
	// These checks are not strictly necessary but they exist to keep retain compatibility with
	// Gitaly's tested behavior.
	repository := req.GetRepository()
	if repository == nil || repository.GetStorageName() == "" && repository.GetRelativePath() == "" {
		return structerr.NewInvalidArgument("%w", storage.ErrRepositoryNotSet)
	} else if repository.GetStorageName() == "" {
		return structerr.NewInvalidArgument("%w", storage.ErrStorageNotSet)
	} else if repository.GetRelativePath() == "" {
		return structerr.NewInvalidArgument("%w", storage.ErrRepositoryPathNotSet)
	} else if _, ok := virtualStorages[repository.GetStorageName()]; !ok {
		return storage.NewStorageNotFoundError(repository.GetStorageName())
	} else if req.GetRelativePath() == "" {
		return structerr.NewInvalidArgument("destination relative path is empty")
	} else if _, err := storage.ValidateRelativePath("/fake-root", req.GetRelativePath()); err != nil {
		// Gitaly uses ValidateRelativePath to verify there are no traversals, so we use the same function
		// here. Praefect is not susceptible to path traversals as it generates its own disk paths but we
		// do this to retain API compatibility with Gitaly. ValidateRelativePath checks for traversals by
		// seeing whether the relative path escapes the root directory. It's not possible to traverse up
		// from the /, so the traversals in the path wouldn't be caught. To allow for the check to work,
		// we use the /fake-root directory simply to notice if there were traversals in the path.
		return structerr.NewInvalidArgument("%w", err).WithMetadata("relative_path", req.GetRelativePath())
	}

	return nil
}

// RenameRepositoryHandler handles /gitaly.RepositoryService/RenameRepository calls by renaming
// the repository in the lookup table stored in the database.
func RenameRepositoryHandler(virtualStoragesNames []string, rs datastore.RepositoryStore) grpc.StreamHandler {
	virtualStorages := make(map[string]struct{}, len(virtualStoragesNames))
	for _, virtualStorage := range virtualStoragesNames {
		virtualStorages[virtualStorage] = struct{}{}
	}

	return func(srv interface{}, stream grpc.ServerStream) error {
		var req gitalypb.RenameRepositoryRequest
		if err := stream.RecvMsg(&req); err != nil {
			return fmt.Errorf("receive request: %w", err)
		}

		if err := validateRenameRepositoryRequest(&req, virtualStorages); err != nil {
			return err
		}

		if err := rs.RenameRepositoryInPlace(stream.Context(),
			req.GetRepository().GetStorageName(),
			req.GetRepository().GetRelativePath(),
			req.GetRelativePath(),
		); err != nil {
			if errors.Is(err, datastore.ErrRepositoryNotFound) {
				return storage.NewRepositoryNotFoundError(
					req.GetRepository().GetStorageName(),
					req.GetRepository().GetRelativePath(),
				)
			} else if errors.Is(err, datastore.ErrRepositoryAlreadyExists) {
				return structerr.NewAlreadyExists("target repo exists already")
			}

			return structerr.NewInternal("%w", err)
		}

		return stream.SendMsg(&gitalypb.RenameRepositoryResponse{})
	}
}
