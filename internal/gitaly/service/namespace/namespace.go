package namespace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/perm"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var noNameError = status.Errorf(codes.InvalidArgument, "Name: cannot be empty")

func (s *server) NamespaceExists(ctx context.Context, in *gitalypb.NamespaceExistsRequest) (*gitalypb.NamespaceExistsResponse, error) {
	storagePath, err := s.locator.GetStorageByName(in.GetStorageName())
	if err != nil {
		return nil, err
	}

	// This case should return an error, as else we'd actually say the path exists as the
	// storage exists
	if in.GetName() == "" {
		return nil, noNameError
	}

	if fi, err := os.Stat(namespacePath(storagePath, in.GetName())); os.IsNotExist(err) {
		return &gitalypb.NamespaceExistsResponse{Exists: false}, nil
	} else if err != nil {
		return nil, structerr.NewInternal("could not stat the directory: %w", err)
	} else {
		return &gitalypb.NamespaceExistsResponse{Exists: fi.IsDir()}, nil
	}
}

func (s *server) AddNamespace(ctx context.Context, in *gitalypb.AddNamespaceRequest) (*gitalypb.AddNamespaceResponse, error) {
	storagePath, err := s.locator.GetStorageByName(in.GetStorageName())
	if err != nil {
		return nil, err
	}

	name := in.GetName()
	if len(name) == 0 {
		return nil, noNameError
	}

	if err = os.MkdirAll(namespacePath(storagePath, name), perm.GroupPrivateDir); err != nil {
		return nil, structerr.NewInternal("create directory: %w", err)
	}

	return &gitalypb.AddNamespaceResponse{}, nil
}

func (s *server) validateRenameNamespaceRequest(ctx context.Context, in *gitalypb.RenameNamespaceRequest) error {
	if in.GetFrom() == "" || in.GetTo() == "" {
		return errors.New("from and to cannot be empty")
	}

	// No need to check if the from path exists, if it doesn't, we'd later get an
	// os.LinkError
	toExistsCheck := &gitalypb.NamespaceExistsRequest{StorageName: in.StorageName, Name: in.GetTo()}
	if exists, err := s.NamespaceExists(ctx, toExistsCheck); err != nil {
		return err
	} else if exists.Exists {
		return fmt.Errorf("to directory %s already exists", in.GetTo())
	}

	return nil
}

func (s *server) RenameNamespace(ctx context.Context, in *gitalypb.RenameNamespaceRequest) (*gitalypb.RenameNamespaceResponse, error) {
	if err := s.validateRenameNamespaceRequest(ctx, in); err != nil {
		return nil, structerr.NewInvalidArgument("%w", err)
	}

	storagePath, err := s.locator.GetStorageByName(in.GetStorageName())
	if err != nil {
		return nil, err
	}

	targetPath := namespacePath(storagePath, in.GetTo())

	// Create the parent directory.
	if err = os.MkdirAll(filepath.Dir(targetPath), perm.SharedDir); err != nil {
		return nil, structerr.NewInternal("create directory: %w", err)
	}

	err = os.Rename(namespacePath(storagePath, in.GetFrom()), targetPath)
	if _, ok := err.(*os.LinkError); ok {
		return nil, structerr.NewInvalidArgument("from directory %s not found", in.GetFrom())
	} else if err != nil {
		return nil, structerr.NewInternal("rename: %w", err)
	}

	return &gitalypb.RenameNamespaceResponse{}, nil
}

func (s *server) RemoveNamespace(ctx context.Context, in *gitalypb.RemoveNamespaceRequest) (*gitalypb.RemoveNamespaceResponse, error) {
	storagePath, err := s.locator.GetStorageByName(in.GetStorageName())
	if err != nil {
		return nil, err
	}

	// Needed as else we might destroy the whole storage
	if in.GetName() == "" {
		return nil, noNameError
	}

	// os.RemoveAll is idempotent by itself
	// No need to check if the directory exists, or not
	if err = os.RemoveAll(namespacePath(storagePath, in.GetName())); err != nil {
		return nil, structerr.NewInternal("removal: %w", err)
	}
	return &gitalypb.RemoveNamespaceResponse{}, nil
}

func namespacePath(storage, ns string) string {
	return filepath.Join(storage, ns)
}
