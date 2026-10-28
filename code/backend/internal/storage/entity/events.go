package entity

import (
	d "github.com/maximka-44js/lc/internal/app/domain"
)

type LibraryCreated struct {
	d.BaseEvent
	LibID  string
	userID string
}

func NewLibraryCreated(LibID, userID string) LibraryCreated {
	return LibraryCreated{
		BaseEvent: d.NewBaseEvent(),
		LibID:     LibID,
		userID:    userID,
	}
}

func (LibraryCreated) EventType() string {
	return "library.created"
}

type LibraryDeleted struct {
	d.BaseEvent
	LibID  string
	userID string
}

func NewLibraryDeleted(LibID, userID string) LibraryDeleted {
	return LibraryDeleted{
		BaseEvent: d.NewBaseEvent(),
		LibID:     LibID,
		userID:    userID,
	}
}

func (LibraryDeleted) EventType() string {
	return "library.deleted"
}

type FileInitialized struct {
	d.BaseEvent
	FileID string
	userID string
	LibID  string
}

func NewFileInitialized(FileID, LibID, userID string) FileInitialized {
	return FileInitialized{
		BaseEvent: d.NewBaseEvent(),
		FileID:    FileID,
		LibID:     LibID,
		userID:    userID,
	}
}

func (FileInitialized) EventType() string {
	return "file.initialized"
}

type FileUploaded struct {
	d.BaseEvent
	FileID           string
	userID           string
	LibID            string
	DeclaredChecksum string
	Size             int64
	MimeType         string
}

func NewFileUploaded(FileID, LibID, userID, declaredChecksum string, size int64, mimeType string) FileUploaded {
	return FileUploaded{
		BaseEvent:        d.NewBaseEvent(),
		FileID:           FileID,
		LibID:            LibID,
		userID:           userID,
		DeclaredChecksum: declaredChecksum,
		Size:             size,
		MimeType:         mimeType,
	}
}

func (FileUploaded) EventType() string {
	return "file.uploaded"
}

type FileDeleted struct {
	d.BaseEvent
	FileID string
	userID string
	LibID  string
}

func NewFileDeleted(FileID, LibID, userID string) FileDeleted {
	return FileDeleted{
		BaseEvent: d.NewBaseEvent(),
		FileID:    FileID,
		LibID:     LibID,
		userID:    userID,
	}
}

func (FileDeleted) EventType() string {
	return "file.deleted"
}
