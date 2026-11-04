package restapi

import (
	"github.com/jackc/pgx/v5/pgxpool"
	access_uc "github.com/maximka-44js/lc/internal/access/usecase"
	projects_uc "github.com/maximka-44js/lc/internal/projects/usecase"
	storage_repo "github.com/maximka-44js/lc/internal/storage/repo"
	storage_uc "github.com/maximka-44js/lc/internal/storage/usecase"
	"github.com/maximka-44js/lc/pkg/jwt"
	"github.com/maximka-44js/lc/pkg/logger"
)

type StorageDeps struct {
	Library       storage_uc.Library
	Folder        storage_uc.Folder
	File          storage_uc.File
	ObjectStorage storage_repo.ObjectStorage
}

type ProjectsDeps struct {
	Project   projects_uc.Project
	Stage     projects_uc.Stage
	Task      projects_uc.Task
	Policy    projects_uc.Policy
	ReadModel projects_uc.ReadModel
}

type AccessDeps struct {
	APIKeys access_uc.APIKeys
}

type Deps struct {
	Storage  StorageDeps
	Projects ProjectsDeps
	Access   AccessDeps
	JWT      *jwt.Manager
	Log      logger.Interface
	DB       *pgxpool.Pool
}
