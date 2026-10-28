package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

// AccessRole -.
type AccessRole string

const (
	AccessReader AccessRole = "reader"
	AccessWriter AccessRole = "writer"
)

// LibraryAccess -.
// @name entity.LibraryAccess
// @description Represents a user's access to a library.
type LibraryAccess struct {
	UserID string
	Role   AccessRole
}

// Library -.
// @name entity.Library
// @description Represents a library of documents.
type Library struct {
	ID          string `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID      string `json:"user_id"     example:"550e8400-e29b-41d4-a716-446655440000"`
	Accesses    []LibraryAccess
	Title       string     `json:"title"       example:"My Library"`
	Description string     `json:"description" example:"A library of my documents"`
	Public      bool       `json:"public"      example:"false"`
	CreatedAt   time.Time  `json:"created_at"  example:"2026-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at"  example:"2026-01-01T00:00:00Z"`
	DeletedAt   *time.Time `json:"deleted_at"  example:"2026-01-01T00:00:00Z"`
	events      []d.DomainEvent
}

func NewLibrary(id, userID, title, description string, public bool, accesses []LibraryAccess) (Library, error) {
	l := Library{
		ID:          id,
		UserID:      userID,
		Accesses:    accesses,
		Title:       title,
		Description: description,
		Public:      public,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := l.Validate(); err != nil {
		return Library{}, err
	}
	event := NewLibraryCreated(id, userID)
	l.record(event)
	return l, nil
}

func (l *Library) Validate() error {
	if err := l.validateName(); err != nil {
		return err
	}
	if err := l.validateTime(); err != nil {
		return err
	}
	if err := l.validateDescription(); err != nil {
		return err
	}
	if err := l.validateAccesses(); err != nil {
		return err
	}
	return nil
}

func (l *Library) validateName() error {
	title := strings.TrimSpace(l.Title)
	if len(title) == 0 || len(title) > 100 {
		return d.NewValidationError(
			CodeLibraryNameInvalid,
			"Library's title must contain beetwen 1 and 100 symbols",
		)
	} else if !reSafeTitle.MatchString(title) {
		return d.NewValidationError(
			CodeLibraryNameInvalid,
			"Library's name can only contain letters, numbers, spaces, dashes and underscores",
		)
	}
	return nil
}

func (l *Library) validateTime() error {
	if l.CreatedAt.IsZero() || l.UpdatedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"CreatedAt and UpdatedAt fields must be set",
		)
	} else if l.UpdatedAt.Before(l.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before CreatedAt",
		)
	} else if l.DeletedAt != nil && l.DeletedAt.Before(l.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"DeletedAt cannot be before CreatedAt",
		)
	} else if l.DeletedAt != nil && l.UpdatedAt.Before(*l.DeletedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before DeletedAt",
		)
	}
	return nil
}

func (l *Library) validateDescription() error {
	if len(l.Description) > 1000 {
		return d.NewValidationError(
			CodeLibraryDescriptionInvalid,
			"Library's description cannot be longer than 1000 symbols",
		)
	} else if !reSafeDescription.MatchString(l.Description) {
		return d.NewValidationError(
			CodeLibraryDescriptionInvalid,
			"Library's description can only contain letters, numbers, spaces, dashes and underscores",
		)
	}
	return nil
}

func (l *Library) validateAccesses() error {
	// Проверка на дубликаты пользователей
	seen := make(map[string]bool)

	for _, access := range l.Accesses {
		// Проверка валидности UUID
		if _, err := uuid.Parse(access.UserID); err != nil {
			return d.NewValidationError(
				CodeLibraryAccessInvalid,
				"Invalid user ID format in access list",
			)
		}

		// В публичной библиотеке только writer
		if l.Public && access.Role != AccessWriter {
			return d.NewValidationError(
				CodeLibraryAccessInvalid,
				"Public library can only have accesses with 'writer' role",
			)
		}

		// Владелец не может быть в списке доступов
		if access.UserID == l.UserID {
			return d.NewValidationError(
				CodeLibraryAccessInvalid,
				"Library owner cannot be in access list of their library",
			)
		}

		// Проверка валидности роли
		if access.Role != AccessReader && access.Role != AccessWriter {
			return d.NewValidationError(
				CodeLibraryAccessInvalid,
				"Library access role must be either 'reader' or 'writer'",
			)
		}

		// Проверка на дубликаты
		if seen[access.UserID] {
			return d.NewValidationError(
				CodeLibraryAccessInvalid,
				"Duplicate user in access list",
			)
		}
		seen[access.UserID] = true
	}

	return nil
}

func (l *Library) CanRead(userID string) bool {
	if l.UserID == userID {
		return true
	}

	if l.Public {
		return true
	}

	for _, access := range l.Accesses {
		if access.UserID == userID {
			return true
		}
	}

	return false
}

func (l *Library) CanWrite(userID string) bool {
	if l.UserID == userID {
		return true
	}

	for _, access := range l.Accesses {
		if access.UserID == userID && access.Role == AccessWriter {
			return true
		}
	}

	return false
}

func (l *Library) record(event d.DomainEvent) {
	l.events = append(l.events, event)
}

func (l *Library) PullEvents() []d.DomainEvent {
	events := l.events
	l.events = nil
	return events
}
