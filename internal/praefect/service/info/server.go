package info

import (
	"context"
	"errors"

	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/service"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

// AssignmentStore is an interface for getting repository host node assignments.
//
// This duplicates the praefect.AssignmentGetter type as it is not possible to import anything from
// `praefect` to `info` packages due to cyclic dependencies.
type AssignmentStore interface {
	// GetHostAssignments returns the names of the storages assigned to host the repository.
	// The primary node must always be assigned.
	GetHostAssignments(ctx context.Context, virtualStorage string, repositoryID int64) ([]string, error)
	// SetReplicationFactor sets a repository's replication factor and returns the current assignments.
	SetReplicationFactor(ctx context.Context, virtualStorage, relativePath string, replicationFactor int) ([]string, error)
}

// PrimaryGetter is an interface for getting a primary of a repository.
//
// This duplicates the praefect.PrimaryGetter type as it is not possible to import anything from
// `praefect` to `info` packages due to cyclic dependencies.
type PrimaryGetter interface {
	// GetPrimary returns the primary storage for a given repository.
	GetPrimary(ctx context.Context, virtualStorage string, repositoryID int64) (string, error)
}

// Server is a InfoService server
type Server struct {
	gitalypb.UnimplementedPraefectInfoServiceServer
	conf            config.Config
	rs              datastore.RepositoryStore
	assignmentStore AssignmentStore
	conns           service.Connections
	primaryGetter   PrimaryGetter
}

// NewServer creates a new instance of a grpc InfoServiceServer
func NewServer(
	conf config.Config,
	rs datastore.RepositoryStore,
	assignmentStore AssignmentStore,
	conns service.Connections,
	primaryGetter PrimaryGetter,
) gitalypb.PraefectInfoServiceServer {
	return &Server{
		conf:            conf,
		rs:              rs,
		assignmentStore: assignmentStore,
		conns:           conns,
		primaryGetter:   primaryGetter,
	}
}

//nolint:revive // This is unintentionally missing documentation.
func (s *Server) SetAuthoritativeStorage(ctx context.Context, req *gitalypb.SetAuthoritativeStorageRequest) (*gitalypb.SetAuthoritativeStorageResponse, error) {
	storages := s.conf.StorageNames()[req.VirtualStorage]
	if storages == nil {
		return nil, structerr.NewInvalidArgument("unknown virtual storage: %q", req.VirtualStorage)
	}

	foundStorage := false
	for i := range storages {
		if storages[i] == req.AuthoritativeStorage {
			foundStorage = true
			break
		}
	}

	if !foundStorage {
		return nil, structerr.NewInvalidArgument("unknown authoritative storage: %q", req.AuthoritativeStorage)
	}

	if err := s.rs.SetAuthoritativeReplica(ctx, req.VirtualStorage, req.RelativePath, req.AuthoritativeStorage); err != nil {
		if errors.Is(err, datastore.ErrRepositoryNotFound) {
			return nil, structerr.NewInvalidArgument("repository does not exist on virtual storage").WithMetadataItems(
				structerr.MetadataItem{Key: "virtual_storage", Value: req.VirtualStorage},
				structerr.MetadataItem{Key: "relative_path", Value: req.RelativePath},
				structerr.MetadataItem{Key: "authoritative_storage", Value: req.AuthoritativeStorage},
			)
		}

		return nil, structerr.NewInternal("%w", err)
	}

	return &gitalypb.SetAuthoritativeStorageResponse{}, nil
}
