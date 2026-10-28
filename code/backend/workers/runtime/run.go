// Package runtime configures and runs River workers.
package runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/maximka-44js/lc/config"
	storagepersistent "github.com/maximka-44js/lc/internal/storage/repo/persistent"
	storageS3 "github.com/maximka-44js/lc/internal/storage/repo/s3"
	"github.com/maximka-44js/lc/pkg/logger"
	"github.com/maximka-44js/lc/pkg/postgres"
	folderworker "github.com/maximka-44js/lc/workers/storage/folder"
	libraryworker "github.com/maximka-44js/lc/workers/storage/library"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

const defaultQueueName = "default"

// Run starts River worker process.
func Run(cfg *config.WorkerConfig) {
	l := logger.New(cfg.LogLevel)

	pg, err := postgres.New(cfg.PGURL, postgres.MaxPoolSize(cfg.PGPoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("worker - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	queueName := cfg.RiverQueue
	if queueName == "" {
		queueName = defaultQueueName
	}

	objectStorage, err := storageS3.NewS3Storage(context.Background(), storageS3.S3Config{
		Region:        cfg.S3.Region,
		Endpoint:      cfg.S3.Endpoint(),
		Bucket:        cfg.S3.Bucket,
		AccessKey:     cfg.S3.AccessKey,
		SecretKey:     cfg.S3.SecretKey,
		UsePathStyle:  cfg.S3.UsePathStyle,
		PresignExpire: cfg.S3.PresignExpire,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("worker - Run - s3.NewS3Storage: %w", err))
	}

	workers := river.NewWorkers()

	riverClient, err := river.NewClient[pgx.Tx](riverpgxv5.New(pg.Pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			queueName: {MaxWorkers: cfg.RiverMaxWorkers},
		},
		Workers: workers,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("worker - Run - river.NewClient: %w", err))
	}
	fileRepo := storagepersistent.NewFileRepo(pg)
	folderRepo := storagepersistent.NewFolderRepo(pg)
	libRepo := storagepersistent.NewLibRepo(pg)
	libraryCompress := libraryworker.NewCompressUseCase(
		libRepo,
		folderRepo,
		fileRepo,
		fileRepo,
		objectStorage,
		libraryworker.NewObjectReaderPort(objectStorage),
		pg,
	)
	folderCompress := folderworker.NewCompressUseCase(
		libRepo,
		folderRepo,
		fileRepo,
		objectStorage,
		folderworker.NewObjectReaderPort(objectStorage),
		pg,
	)
	folderExtract := folderworker.NewExtractUseCase(
		libRepo,
		folderRepo,
		fileRepo,
		objectStorage,
		folderworker.NewObjectReaderPort(objectStorage),
		pg,
	)
	RegisterWorkers(workers, Deps{
		PG:              pg,
		Log:             l,
		ObjectStorage:   objectStorage,
		LibraryCompress: libraryCompress,
		FolderCompress:  folderCompress,
		FolderExtract:   folderExtract,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := riverClient.Start(ctx); err != nil {
		l.Fatal(fmt.Errorf("worker - Run - riverClient.Start: %w", err))
	}

	<-riverClient.Stopped()
}
