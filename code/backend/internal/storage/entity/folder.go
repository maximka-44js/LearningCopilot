package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

type Folder struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	LibID     string     `json:"library_id"`
	Name      string     `json:"name"`
	ParentID  string     `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func NewFolder(id, userID, libID, name, parentID string) (Folder, error) {
	f := Folder{
		ID:        id,
		UserID:    userID,
		LibID:     libID,
		Name:      name,
		ParentID:  parentID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := f.Validate(); err != nil {
		return Folder{}, err
	}
	return f, nil
}

func (f *Folder) Validate() error {
	if err := f.validateName(); err != nil {
		return err
	}
	if err := f.validateIDs(); err != nil {
		return err
	}
	if err := f.validateParentID(); err != nil {
		return err
	}
	if err := f.validateTime(); err != nil {
		return err
	}
	return nil
}

func (f *Folder) validateName() error {
	name := strings.TrimSpace(f.Name)
	if len(name) == 0 || len(name) > 100 {
		return d.NewValidationError(
			CodeFolderNameInvalid,
			"Folder's name must contain between 1 and 100 symbols",
		)
	} else if !reSafeTitle.MatchString(name) {
		return d.NewValidationError(
			CodeFolderNameInvalid,
			"Folder's name can only contain letters, numbers, spaces, dashes and underscores",
		)
	} else if name == "." || name == ".." {
		return d.NewValidationError(
			CodeFolderNameInvalid,
			"Folder's name cannot be '.' or '..'",
		)
	}
	return nil
}

func (f *Folder) validateIDs() error {
	if _, err := uuid.Parse(f.ID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Folder's id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(f.UserID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Folder's user_id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(f.LibID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Folder's library_id must be a valid UUID",
		)
	}
	return nil
}

func (f *Folder) validateParentID() error {
	if strings.TrimSpace(f.ParentID) == "" {
		return nil
	}
	if _, err := uuid.Parse(f.ParentID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Folder's parent_id must be a valid UUID",
		)
	}
	if f.ParentID == f.ID {
		return d.NewValidationError(
			CodeInvalidID,
			"Folder's parent_id cannot be the same as id",
		)
	}
	return nil
}

func (f *Folder) validateTime() error {
	if f.CreatedAt.IsZero() || f.UpdatedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"CreatedAt and UpdatedAt fields must be set",
		)
	} else if f.UpdatedAt.Before(f.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before CreatedAt",
		)
	} else if f.DeletedAt != nil && f.DeletedAt.Before(f.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"DeletedAt cannot be before CreatedAt",
		)
	} else if f.DeletedAt != nil && f.UpdatedAt.Before(*f.DeletedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before DeletedAt",
		)
	}
	return nil
}
