package entity

import (
	d "github.com/maximka-44js/lc/internal/app/domain"
)

const (
	CodeLibraryNameInvalid        d.ErrorCode = "library.name_invalid"
	CodeTimeInvalid               d.ErrorCode = "time.invalid"
	CodeLibraryDescriptionInvalid d.ErrorCode = "library.description_invalid"
	CodeLibraryAccessInvalid      d.ErrorCode = "library.access_invalid"
	CodeLibNotFound               d.ErrorCode = "library.not_found"
	CodeFolderNameInvalid         d.ErrorCode = "folder.name_invalid"
	CodeFolderAccessInvalid       d.ErrorCode = "folder.access_invalid"
	CodeFolderNotFound            d.ErrorCode = "folder.not_found"
	CodeFileNameInvalid           d.ErrorCode = "file.name_invalid"
	CodeFileTypeInvalid           d.ErrorCode = "file.type_invalid"
	CodeFileChecksumInvalid       d.ErrorCode = "file.checksum_invalid"
	CodeFileStatusInvalid         d.ErrorCode = "file.status_invalid"
	CodeFileSizeInvalid           d.ErrorCode = "file.size_invalid"
	CodeFileNotFound              d.ErrorCode = "file.not_found"
	CodeInvalidID                 d.ErrorCode = "invalid_uuid"
	CodeQuotaInvalid              d.ErrorCode = "quota_invalid"
	CodeUndefined                 d.ErrorCode = "undefined"
)
