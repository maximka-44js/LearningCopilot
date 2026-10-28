package request

import (
	"github.com/maximka-44js/lc/internal/storage/entity"
)

type CreateLibrary struct {
	Title       string                 `json:"title"       validate:"required,max=100" example:"My library"`
	Description string                 `json:"description" validate:"max=1000"         example:"Library description"`
	Public      bool                   `json:"public"      example:"false"`
	Accesses    []entity.LibraryAccess `json:"accesses"`
}

type GetLibraryByID struct {
	LibID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type UpdateLibraryParams struct {
	LibID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type UpdateLibrary struct {
	Title       string                 `json:"title"       validate:"required,max=100" example:"My library"`
	Description string                 `json:"description" validate:"max=1000"         example:"Library description"`
	Public      bool                   `json:"public"      example:"false"`
	Accesses    []entity.LibraryAccess `json:"accesses"`
}
