package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/maximka-44js/lc/internal/projects/usecase"
	"github.com/maximka-44js/lc/pkg/logger"
)

type V1 struct {
	p  usecase.Project
	s  usecase.Stage
	t  usecase.Task
	po usecase.Policy
	rm usecase.ReadModel
	l  logger.Interface
	v  *validator.Validate
}
