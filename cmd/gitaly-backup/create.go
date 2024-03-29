package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"runtime"
	"time"

	"gitlab.com/gitlab-org/gitaly/v16/internal/backup"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/internal/grpc/client"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

type serverRepository struct {
	storage.ServerInfo
	StorageName   string `json:"storage_name"`
	RelativePath  string `json:"relative_path"`
	GlProjectPath string `json:"gl_project_path"`
}

type createSubcommand struct {
	backupPath      string
	parallel        int
	parallelStorage int
	layout          string
	incremental     bool
	backupID        string
	serverSide      bool
}

func (cmd *createSubcommand) Flags(fs *flag.FlagSet) {
	fs.StringVar(&cmd.backupPath, "path", "", "repository backup path")
	fs.IntVar(&cmd.parallel, "parallel", runtime.NumCPU(), "maximum number of parallel backups")
	fs.IntVar(&cmd.parallelStorage, "parallel-storage", 2, "maximum number of parallel backups per storage. Note: actual parallelism when combined with `-parallel` depends on the order the repositories are received.")
	fs.StringVar(&cmd.layout, "layout", "pointer", "how backup files are located. Either pointer or legacy.")
	fs.BoolVar(&cmd.incremental, "incremental", false, "creates an incremental backup if possible.")
	fs.StringVar(&cmd.backupID, "id", time.Now().UTC().Format("20060102150405"), "the backup ID used when creating a full backup.")
	fs.BoolVar(&cmd.serverSide, "server-side", false, "use server-side backups. Note: The feature is not ready for production use.")
}

func (cmd *createSubcommand) Run(ctx context.Context, logger log.Logger, stdin io.Reader, stdout io.Writer) error {
	pool := client.NewPool(client.WithDialOptions(client.UnaryInterceptor(), client.StreamInterceptor()))
	defer func() {
		_ = pool.Close()
	}()

	var manager backup.Strategy
	if cmd.serverSide {
		if cmd.backupPath != "" {
			return fmt.Errorf("create: path cannot be used with server-side backups")
		}

		manager = backup.NewServerSideAdapter(pool)
	} else {
		sink, err := backup.ResolveSink(ctx, cmd.backupPath)
		if err != nil {
			return fmt.Errorf("create: resolve sink: %w", err)
		}

		locator, err := backup.ResolveLocator(cmd.layout, sink)
		if err != nil {
			return fmt.Errorf("create: resolve locator: %w", err)
		}

		manager = backup.NewManager(sink, locator, pool)
	}

	var pipeline backup.Pipeline
	pipeline = backup.NewLoggingPipeline(logger)
	if cmd.parallel > 0 || cmd.parallelStorage > 0 {
		pipeline = backup.NewParallelPipeline(pipeline, cmd.parallel, cmd.parallelStorage)
	}

	decoder := json.NewDecoder(stdin)
	for {
		var sr serverRepository
		if err := decoder.Decode(&sr); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		repo := gitalypb.Repository{
			StorageName:   sr.StorageName,
			RelativePath:  sr.RelativePath,
			GlProjectPath: sr.GlProjectPath,
		}
		pipeline.Handle(ctx, backup.NewCreateCommand(manager, backup.CreateRequest{
			Server:           sr.ServerInfo,
			Repository:       &repo,
			VanityRepository: &repo,
			Incremental:      cmd.incremental,
			BackupID:         cmd.backupID,
		}))
	}

	if err := pipeline.Done(); err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}
