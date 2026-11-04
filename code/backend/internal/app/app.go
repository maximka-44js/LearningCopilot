// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/maximka-44js/lc/config"
	accessPersistent "github.com/maximka-44js/lc/internal/access/repo/persistent"
	accessUseCase "github.com/maximka-44js/lc/internal/access/usecase/api_keys"
	"github.com/maximka-44js/lc/internal/app/controller/restapi"
	projectsPersistent "github.com/maximka-44js/lc/internal/projects/repo/persistent"
	projectPolicyUseCase "github.com/maximka-44js/lc/internal/projects/usecase/policy"
	projectUseCase "github.com/maximka-44js/lc/internal/projects/usecase/project"
	projectReadModelUseCase "github.com/maximka-44js/lc/internal/projects/usecase/read_model"
	stageUseCase "github.com/maximka-44js/lc/internal/projects/usecase/stage"
	projectTaskUseCase "github.com/maximka-44js/lc/internal/projects/usecase/task"
	storageRepo "github.com/maximka-44js/lc/internal/storage/repo"
	storagePersistent "github.com/maximka-44js/lc/internal/storage/repo/persistent"
	storagepublisher "github.com/maximka-44js/lc/internal/storage/repo/publisher"
	storageS3 "github.com/maximka-44js/lc/internal/storage/repo/s3"
	storageFileUC "github.com/maximka-44js/lc/internal/storage/usecase/file"
	"github.com/maximka-44js/lc/internal/storage/usecase/folder"
	"github.com/maximka-44js/lc/internal/storage/usecase/library"
	"github.com/maximka-44js/lc/pkg/httpserver"
	"github.com/maximka-44js/lc/pkg/jwt"
	"github.com/maximka-44js/lc/pkg/logger"
	"github.com/maximka-44js/lc/pkg/postgres"
	queueriver "github.com/maximka-44js/lc/pkg/queue/river"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type useCases struct {
	library     *library.UseCase
	folder      *folder.UseCase
	file        *storageFileUC.UseCase
	apiKeys     *accessUseCase.UseCase
	project     *projectUseCase.UseCase
	stage       *stageUseCase.UseCase
	projectTask *projectTaskUseCase.UseCase
	policy      *projectPolicyUseCase.UseCase
	readModel   *projectReadModelUseCase.UseCase
}

type servers struct {
	http *httpserver.Server
}

func initUseCases(cfg *config.Config, pg *postgres.Postgres, jwtManager *jwt.Manager, objectStorage storageRepo.ObjectStorage, l logger.Interface) (useCases, error) {
	riverClient, err := river.NewClient[pgx.Tx](riverpgxv5.New(pg.Pool), &river.Config{})
	if err != nil {
		return useCases{}, fmt.Errorf("app - initUseCases - river.NewClient: %w", err)
	}
	queuePublisher := queueriver.NewPublisher(riverClient)

	libraryJobs := storagepublisher.NewLibraryJobs(queuePublisher)
	folderJobs := storagepublisher.NewFolderJobs(queuePublisher)
	fileJobs := storagepublisher.NewFileJobs(queuePublisher)
	libRepo := storagePersistent.NewLibRepo(pg)
	folderRepo := storagePersistent.NewFolderRepo(pg)
	fileRepo := storagePersistent.NewFileRepo(pg)
	projectsProjectRepo := projectsPersistent.NewProjectRepo(pg)
	projectsStageRepo := projectsPersistent.NewStageRepo(pg)
	projectsTaskRepo := projectsPersistent.NewTaskRepo(pg)
	projectsReadModelRepo := projectsPersistent.NewReadModelRepo(pg)
	projectsUserPolicyRepo := projectsPersistent.NewUserPolicyRepo(pg)
	projectsProjectPolicyRepo := projectsPersistent.NewProjectPolicyRepo(pg)
	projectsAssessmentSnapshotRepo := projectsPersistent.NewAssessmentSnapshotRepo(pg)
	apiKeyRepo := accessPersistent.NewAPIKeyRepo(pg)
	fileUC := storageFileUC.NewFileUseCase(fileRepo, objectStorage, folderRepo, libRepo, fileJobs, pg)

	projectUC := projectUseCase.NewProjectUseCase(projectsProjectRepo, projectsStageRepo, projectsProjectRepo)
	stageUC := stageUseCase.NewStageUseCase(projectsProjectRepo, projectsStageRepo, projectsTaskRepo, projectsProjectRepo)
	projectTaskUC := projectTaskUseCase.NewTaskUseCase(
		projectsProjectRepo,
		projectsStageRepo,
		projectsTaskRepo,
		projectsAssessmentSnapshotRepo,
		projectsProjectRepo,
	)
	policyUC := projectPolicyUseCase.NewPolicyUseCase(
		projectsProjectRepo,
		projectsStageRepo,
		projectsTaskRepo,
		projectsUserPolicyRepo,
		projectsProjectPolicyRepo,
		pg,
	)

	apiKeyUC := accessUseCase.NewUseCase(apiKeyRepo, accessUseCase.Config{
		Enabled:           cfg.APIKeys.Enabled,
		HashPepper:        cfg.APIKeys.HashPepper,
		MaxKeysPerUser:    cfg.APIKeys.MaxPerUser,
		RotateGraceWindow: cfg.APIKeys.RotateGraceWindow,
	})

	return useCases{
		library:     library.NewLibraryUseCase(libRepo, libraryJobs, pg),
		folder:      folder.NewFolderUseCase(folderRepo, fileRepo, objectStorage, libRepo, folderJobs, pg),
		file:        fileUC,
		apiKeys:     apiKeyUC,
		project:     projectUC,
		stage:       stageUC,
		projectTask: projectTaskUC,
		policy:      policyUC,
		readModel:   projectReadModelUseCase.NewReadModelUseCase(projectsReadModelRepo),
	}, nil
}

func initObjectStorage(cfg *config.Config) (storageRepo.ObjectStorage, error) {
	return storageS3.NewS3Storage(context.Background(), storageS3.S3Config{
		Region:        cfg.S3.Region,
		Endpoint:      cfg.S3.Endpoint(),
		Bucket:        cfg.S3.Bucket,
		AccessKey:     cfg.S3.AccessKey,
		SecretKey:     cfg.S3.SecretKey,
		UsePathStyle:  cfg.S3.UsePathStyle,
		PresignExpire: cfg.S3.PresignExpire,
	})
}

func initServers(cfg *config.Config, uc useCases, objectStorage storageRepo.ObjectStorage, jwtManager *jwt.Manager, pg *postgres.Postgres, l logger.Interface) servers {
	httpServer := httpserver.New(
		l,
		httpserver.Port(cfg.HTTP.Port),
		httpserver.Prefork(cfg.HTTP.UsePreforkMode),
		httpserver.ReadTimeout(cfg.HTTP.ReadTimeout),
		httpserver.WriteTimeout(cfg.HTTP.WriteTimeout),
		httpserver.ShutdownTimeout(cfg.HTTP.ShutdownTimeout),
		httpserver.BodyLimit(cfg.HTTP.BodyLimitBytes),
	)
	restapi.NewRouter(httpServer.App, cfg, restapi.Deps{
		Storage: restapi.StorageDeps{
			Library:       uc.library,
			Folder:        uc.folder,
			File:          uc.file,
			ObjectStorage: objectStorage,
		},
		Projects: restapi.ProjectsDeps{
			Project:   uc.project,
			Stage:     uc.stage,
			Task:      uc.projectTask,
			Policy:    uc.policy,
			ReadModel: uc.readModel,
		},
		Access: restapi.AccessDeps{
			APIKeys: uc.apiKeys,
		},
		JWT: jwtManager,
		Log: l,
		DB:  pg.Pool,
	})

	return servers{
		http: httpServer,
	}
}

func (s *servers) startServers() {
	s.http.Start()
}

func (s *servers) waitForShutdown(l logger.Interface) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error

	select {
	case sig := <-interrupt:
		l.Info("app - Run - signal: %s", sig.String())
	case err = <-s.http.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	s.shutdownServers(l)
}

func (s *servers) shutdownServers(l logger.Interface) {
	if err := s.http.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)
	if cfg.APIKeys.Enabled && strings.TrimSpace(cfg.APIKeys.HashPepper) == "" {
		l.Fatal("app - Run - API_KEYS_HASH_PEPPER is required when API_KEYS_ENABLED=true")
	}

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// JWT
	jwtManager, err := jwt.NewWithDiscovery(cfg.JWT.DiscoveryIssuer, cfg.JWT.Issuer, cfg.JWT.ClientID)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - jwt.New: %w", err))
	}

	objectStorage, err := initObjectStorage(cfg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - s3.NewS3Storage: %w", err))
	}

	uc, err := initUseCases(cfg, pg, jwtManager, objectStorage, l)
	if err != nil {
		l.Fatal(err)
	}
	s := initServers(cfg, uc, objectStorage, jwtManager, pg, l)
	s.startServers()
	s.waitForShutdown(l)
}
