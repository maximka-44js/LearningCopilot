package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/maximka-44js/lc/internal/storage/repo"
	"github.com/maximka-44js/lc/internal/storage/usecase"
	"github.com/maximka-44js/lc/pkg/logger"
)

type V1 struct {
	lb usecase.Library
	fd usecase.Folder
	f  usecase.File
	os repo.ObjectStorage
	l  logger.Interface
	v  *validator.Validate
}
