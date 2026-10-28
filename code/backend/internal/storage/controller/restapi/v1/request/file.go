// Request schemas for file controller
package request

type InitUpload struct {
	LibraryID        string `json:"library_id" validate:"required,uuid4"`
	FolderID         string `json:"folder_id,omitempty" validate:"omitempty,uuid4"`
	Name             string `json:"name" validate:"required,max=100"`
	MimeType         string `json:"mime_type" validate:"required"`
	Size             int64  `json:"size" validate:"required,min=0"`
	ChecksumDeclared string `json:"checksum_declared,omitempty"`
}

type PresignPart struct {
	UploadID       string `json:"upload_id" validate:"required"`
	PartNumber     int    `json:"part_number" validate:"required,min=1"`
	ChecksumSHA256 string `json:"checksum_sha256,omitempty" validate:"omitempty"`
	Size           int64  `json:"size" validate:"required,min=1"`
}

type CompletedPart struct {
	ETag       string `json:"etag" validate:"required"`
	PartNumber int    `json:"part_number" validate:"required,min=1"`
}

// ConfirmUpload is sent by client after it finished multipart upload in object storage.
// Server will perform integrity checks and publish FileUploaded event.
type ConfirmUpload struct {
	UploadID       string          `json:"upload_id,omitempty"`
	CompletedParts []CompletedPart `json:"completed_parts,omitempty"`
}

// AbortUpload cancels an unfinished upload and triggers cleanup.
type AbortUpload struct {
	UploadID string `json:"upload_id,omitempty"`
}

type UpdateFileParams struct {
	FileID string `params:"id" validate:"required,uuid4"`
}

type UpdateFile struct {
	FolderID string `json:"folder_id,omitempty" validate:"omitempty,uuid4"`
	Name     string `json:"name" validate:"required,max=100"`
}

type CopyFileParams struct {
	FileID string `params:"id" validate:"required,uuid4"`
}

type CopyFile struct {
	TargetFolderID string `json:"target_folder_id,omitempty" validate:"omitempty,uuid4"`
	Name           string `json:"name,omitempty" validate:"omitempty,max=100"`
}
