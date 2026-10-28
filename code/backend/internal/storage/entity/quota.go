package entity

import (
	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

const (
	maxStorageSizeRegular uint64 = 2 * 1024 * 1024 * 1024  // 2 GB
	maxStorageSizePremium uint64 = 10 * 1024 * 1024 * 1024 // 10 GB
)

type Quota struct {
	MaxStorageSize  uint64 `json:"max_storage_size" example:"3221225472"`  // 2 GB
	UsedStorageSize uint64 `json:"used_storage_size" example:"2147483648"` // 2 GB
	UserID          string `json:"user_id" example:"112eb8c-9f1a-4d2b-8c3e-1234567890ab"`
	IsPremium       bool   `json:"is_premium" example:"false"`
}

func ValidateQuota(q Quota) error {
	if err := validateQuotaBase(q); err != nil {
		return err
	}
	if q.UsedStorageSize > q.MaxStorageSize {
		return d.NewValidationError(CodeQuotaInvalid, "Used storage size cannot exceed max storage size")
	}
	return nil
}

func ValidateQuotaForUpload(q Quota, uploadSize int64) error {
	if err := validateQuotaBase(q); err != nil {
		return err
	}
	if uploadSize < 0 {
		return d.NewValidationError(CodeQuotaInvalid, "Upload size cannot be negative")
	}
	if q.UsedStorageSize+uint64(uploadSize) > q.MaxStorageSize {
		return d.NewValidationError(CodeQuotaInvalid, "Upload exceeds max storage size")
	}
	return nil
}

func validateQuotaBase(q Quota) error {
	if _, err := uuid.Parse(q.UserID); err != nil {
		return d.NewValidationError(CodeInvalidID, "Quota user_id must be a valid UUID")
	}
	if q.MaxStorageSize == 0 {
		return d.NewValidationError(CodeQuotaInvalid, "Max storage size must be greater than 0")
	}
	if q.MaxStorageSize > quotaMaxLimit(q.IsPremium) {
		return d.NewValidationError(CodeQuotaInvalid, "Max storage size exceeds the allowed limit")
	}
	return nil
}

func quotaMaxLimit(isPremium bool) uint64 {
	if isPremium {
		return maxStorageSizePremium
	}
	return maxStorageSizeRegular
}
