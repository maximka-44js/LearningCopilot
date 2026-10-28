package request

type GetFolderByID struct {
	FdID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type CreateFolder struct {
	LibID    string `json:"lib_id"    validate:"required,uuid4"  example:"123e4567-e89b-12d3-a456-426614174000"`
	ParentID string `json:"parent_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name     string `json:"name"      validate:"required,max=100" example:"Project docs"`
}

type UpdateFolderParams struct {
	FdID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type UpdateFolder struct {
	ParentID string `json:"parent_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name     string `json:"name"      validate:"required,max=100" example:"Project docs"`
}

type DeleteFolderParams struct {
	FdID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type CopyFolderParams struct {
	FdID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type CopyFolder struct {
	TargetParentID string `json:"target_parent_id,omitempty" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name           string `json:"name,omitempty" validate:"omitempty,max=100" example:"Project docs copy"`
}
