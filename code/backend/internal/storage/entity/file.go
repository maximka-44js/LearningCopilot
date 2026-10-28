package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

const (
	maxFileNameLength = 100
	maxFileSizeBytes  = int64(500 * 1024 * 1024)
)

type FileStatus string

const (
	FileStatusCreated     FileStatus = "created"
	FileStatusUploaded    FileStatus = "uploaded"
	FileStatusScanPending FileStatus = "scan_pending"
	FileStatusScanOk      FileStatus = "scan_ok"
	FileStatusScanFailed  FileStatus = "scan_failed"
	FileStatusRejected    FileStatus = "rejected"
	FileStatusDeleted     FileStatus = "deleted"
)

var (
	reSHA256 = regexp.MustCompile(`^[A-Fa-f0-9]{64}$`)

	allowedExactMimeTypes = map[string]struct{}{
		"application/pdf":    {},
		"application/msword": {},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
		"application/vnd.ms-excel": {},
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
		"application/vnd.ms-powerpoint":                                             {},
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
		"text/plain":       {},
		"text/markdown":    {},
		"text/csv":         {},
		"application/json": {},
		"application/xml":  {},
		"application/zip":  {},
	}
)

// File -.
// @name entity.File
// @description Represents a file stored in a library or folder.
type File struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	LibID            string     `json:"library_id"`
	FolderID         string     `json:"folder_id,omitempty"`
	Name             string     `json:"name"`
	Size             int64      `json:"size"`
	MimeType         string     `json:"mime_type"`
	ObjectKey        string     `json:"-"`
	ChecksumDeclared string     `json:"checksum_declared"`
	Checksum         string     `json:"checksum"`
	Status           FileStatus `json:"status"`
	ScanAttempts     int        `json:"scan_attempts"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at"`
}

func NewFile(id, userID, libID, folderID, name, mimeType string, size int64, checksumDeclared string) (File, error) {
	f := File{
		ID:               id,
		UserID:           userID,
		LibID:            libID,
		FolderID:         folderID,
		Name:             name,
		Size:             size,
		MimeType:         mimeType,
		ObjectKey:        "",
		ChecksumDeclared: checksumDeclared,
		Status:           FileStatusCreated,
		ScanAttempts:     0,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := f.Validate(); err != nil {
		return File{}, err
	}
	return f, nil
}

func (f *File) Validate() error {
	if err := f.validateName(); err != nil {
		return err
	}
	if err := f.validateIDs(); err != nil {
		return err
	}
	if err := f.validateSize(); err != nil {
		return err
	}
	if err := f.validateMimeType(); err != nil {
		return err
	}
	if err := f.validateChecksums(); err != nil {
		return err
	}
	if err := f.validateStatus(); err != nil {
		return err
	}
	if err := f.validateScanAttempts(); err != nil {
		return err
	}
	if err := f.validateTime(); err != nil {
		return err
	}
	return nil
}

func (f *File) ValidateStatusTransition(next FileStatus) error {
	if next == f.Status {
		return nil
	}
	if !isFileStatusValid(next) {
		return d.NewValidationError(
			CodeFileStatusInvalid,
			"File status is invalid",
		)
	}
	if !canTransitionFileStatus(f.Status, next) {
		return d.NewValidationError(
			CodeFileStatusInvalid,
			"File status transition is not allowed",
		)
	}
	return nil
}

func (f *File) validateName() error {
	name := strings.TrimSpace(f.Name)
	if len(name) == 0 || len(name) > maxFileNameLength {
		return d.NewValidationError(
			CodeFileNameInvalid,
			"File name must contain between 1 and 100 symbols",
		)
	} else if !reSafeTitle.MatchString(name) {
		return d.NewValidationError(
			CodeFileNameInvalid,
			"File name can only contain letters, numbers, spaces, dashes and underscores",
		)
	} else if name == "." || name == ".." {
		return d.NewValidationError(
			CodeFileNameInvalid,
			"File name cannot be '.' or '..'",
		)
	}
	return nil
}

func (f *File) validateIDs() error {
	if _, err := uuid.Parse(f.ID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"File's id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(f.UserID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"File's user_id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(f.LibID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"File's library_id must be a valid UUID",
		)
	}
	if strings.TrimSpace(f.FolderID) == "" {
	} else if _, err := uuid.Parse(f.FolderID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"File's folder_id must be a valid UUID",
		)
	}
	return nil
}

func (f *File) validateSize() error {
	if f.Size < 0 || f.Size > maxFileSizeBytes {
		return d.NewValidationError(
			CodeFileSizeInvalid,
			"File size must be between 0 and 500MB",
		)
	}
	return nil
}

func (f *File) validateMimeType() error {
	mimeType := strings.TrimSpace(f.MimeType)
	if mimeType == "" {
		return d.NewValidationError(
			CodeFileTypeInvalid,
			"File mime_type is required",
		)
	}
	if !isAllowedMimeType(mimeType) {
		return d.NewValidationError(
			CodeFileTypeInvalid,
			"File mime_type is not allowed",
		)
	}
	return nil
}

func (f *File) validateChecksums() error {
	if strings.TrimSpace(f.ChecksumDeclared) != "" && !reSHA256.MatchString(f.ChecksumDeclared) {
		return d.NewValidationError(
			CodeFileChecksumInvalid,
			"File checksum_declared must be a valid sha256",
		)
	}
	if strings.TrimSpace(f.Checksum) != "" && !reSHA256.MatchString(f.Checksum) {
		return d.NewValidationError(
			CodeFileChecksumInvalid,
			"File checksum must be a valid sha256",
		)
	}
	return nil
}

func (f *File) validateStatus() error {
	if !isFileStatusValid(f.Status) {
		return d.NewValidationError(
			CodeFileStatusInvalid,
			"File status is invalid",
		)
	}
	if f.Status == FileStatusDeleted && f.DeletedAt == nil {
		return d.NewValidationError(
			CodeFileStatusInvalid,
			"Deleted file must have deleted_at set",
		)
	}
	return nil
}

func (f *File) validateScanAttempts() error {
	if f.ScanAttempts < 0 {
		return d.NewValidationError(
			CodeFileStatusInvalid,
			"File scan_attempts cannot be negative",
		)
	}
	return nil
}

func (f *File) validateTime() error {
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

func isAllowedMimeType(mimeType string) bool {
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	_, ok := allowedExactMimeTypes[mimeType]
	return ok
}

func isFileStatusValid(status FileStatus) bool {
	switch status {
	case FileStatusCreated,
		FileStatusUploaded,
		FileStatusScanPending,
		FileStatusScanOk,
		FileStatusScanFailed,
		FileStatusRejected,
		FileStatusDeleted:
		return true
	default:
		return false
	}
}
func canTransitionFileStatus(current, next FileStatus) bool {
	if current == FileStatusDeleted {
		return false
	}
	if next == FileStatusDeleted {
		return true
	}

	switch current {
	case FileStatusCreated:
		return next == FileStatusUploaded || next == FileStatusRejected
	case FileStatusUploaded:
		return next == FileStatusScanPending || next == FileStatusRejected
	case FileStatusScanPending:
		return next == FileStatusScanOk || next == FileStatusScanFailed || next == FileStatusRejected
	case FileStatusScanOk, FileStatusScanFailed, FileStatusRejected:
		return false
	default:
		return false
	}
}
