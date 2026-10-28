package entity

import (
	"maps"
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

type ProjectStatus string // @name entity.ProjectStatus

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusPaused   ProjectStatus = "paused"
	ProjectStatusArchived ProjectStatus = "archived"
)

type ProjectPriority string // @name entity.ProjectPriority

const (
	ProjectPriorityLow    ProjectPriority = "low"
	ProjectPriorityNormal ProjectPriority = "normal"
	ProjectPriorityHigh   ProjectPriority = "high"
)

type RiskLevel string // @name entity.RiskLevel

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// Project represents aggregate root for user learning plan.
type Project struct {
	ID          string            `json:"id"`
	OwnerID     string            `json:"owner_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	DeadlineAt  *time.Time        `json:"deadline_at,omitempty"`
	Priority    ProjectPriority   `json:"priority"`
	Status      ProjectStatus     `json:"status"`
	Coverage    float64           `json:"coverage"`
	Readiness   float64           `json:"readiness"`
	Risk        RiskLevel         `json:"risk"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ArchivedAt  *time.Time        `json:"archived_at,omitempty"`
}

func NewProject(
	id string,
	ownerID string,
	name string,
	description string,
	deadlineAt *time.Time,
	priority ProjectPriority,
	metadata map[string]string,
) (Project, error) {
	now := time.Now().UTC()
	p := Project{
		ID:          id,
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		DeadlineAt:  copyTimePointer(deadlineAt),
		Priority:    priority,
		Status:      ProjectStatusActive,
		Coverage:    0,
		Readiness:   0,
		Risk:        RiskLevelMedium,
		Metadata:    cloneMetadata(metadata),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := p.Validate(); err != nil {
		return Project{}, err
	}
	return p, nil
}

func (p *Project) Validate() error {
	if err := p.validateIDs(); err != nil {
		return err
	}
	if err := p.validateName(); err != nil {
		return err
	}
	if err := p.validateDescription(); err != nil {
		return err
	}
	if err := p.validatePriority(); err != nil {
		return err
	}
	if err := p.validateStatus(); err != nil {
		return err
	}
	if err := p.validateMetrics(); err != nil {
		return err
	}
	if err := p.validateDeadline(); err != nil {
		return err
	}
	if err := p.validateMetadata(); err != nil {
		return err
	}
	if err := p.validateTime(); err != nil {
		return err
	}
	return nil
}

func (p *Project) UpdateDetails(
	name string,
	description string,
	deadlineAt *time.Time,
	priority ProjectPriority,
	metadata map[string]string,
) error {
	p.Name = name
	p.Description = description
	p.DeadlineAt = copyTimePointer(deadlineAt)
	p.Priority = priority
	p.Metadata = cloneMetadata(metadata)
	p.UpdatedAt = time.Now().UTC()
	return p.Validate()
}

func (p *Project) UpdateMetrics(coverage, readiness float64, risk RiskLevel) error {
	p.Coverage = coverage
	p.Readiness = readiness
	p.Risk = risk
	p.UpdatedAt = time.Now().UTC()
	return p.Validate()
}

func (p *Project) Pause() error {
	return p.setStatus(ProjectStatusPaused)
}

func (p *Project) Activate() error {
	return p.setStatus(ProjectStatusActive)
}

func (p *Project) Archive() error {
	if err := p.setStatus(ProjectStatusArchived); err != nil {
		return err
	}
	now := time.Now().UTC()
	p.ArchivedAt = &now
	p.UpdatedAt = now
	return nil
}

func (p *Project) setStatus(next ProjectStatus) error {
	if next == p.Status {
		return nil
	}
	if err := p.ValidateStatusTransition(next); err != nil {
		return err
	}
	p.Status = next
	if next != ProjectStatusArchived {
		p.ArchivedAt = nil
	}
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Project) ValidateStatusTransition(next ProjectStatus) error {
	if next == p.Status {
		return nil
	}
	if !isProjectStatusValid(next) {
		return d.NewValidationError(
			CodeProjectStatusInvalid,
			"Project status is invalid",
		)
	}
	if !canTransitionProjectStatus(p.Status, next) {
		return d.NewValidationError(
			CodeProjectStatusInvalid,
			"Project status transition is not allowed",
		)
	}
	return nil
}

func (p *Project) validateIDs() error {
	if _, err := uuid.Parse(p.ID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Project id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(p.OwnerID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Project owner_id must be a valid UUID",
		)
	}
	return nil
}

func (p *Project) validateName() error {
	name := strings.TrimSpace(p.Name)
	if len(name) == 0 || len(name) > 200 {
		return d.NewValidationError(
			CodeProjectNameInvalid,
			"Project name must contain between 1 and 200 symbols",
		)
	}
	return nil
}

func (p *Project) validateDescription() error {
	description := strings.TrimSpace(p.Description)
	if len(description) > 4000 {
		return d.NewValidationError(
			CodeProjectDescriptionInvalid,
			"Project description cannot be longer than 4000 symbols",
		)
	}
	return nil
}

func (p *Project) validatePriority() error {
	if !isProjectPriorityValid(p.Priority) {
		return d.NewValidationError(
			CodeProjectPriorityInvalid,
			"Project priority is invalid",
		)
	}
	return nil
}

func (p *Project) validateStatus() error {
	if !isProjectStatusValid(p.Status) {
		return d.NewValidationError(
			CodeProjectStatusInvalid,
			"Project status is invalid",
		)
	}
	if p.Status == ProjectStatusArchived && p.ArchivedAt == nil {
		return d.NewValidationError(
			CodeProjectStatusInvalid,
			"Archived project must have archived_at set",
		)
	}
	if p.Status != ProjectStatusArchived && p.ArchivedAt != nil {
		return d.NewValidationError(
			CodeProjectStatusInvalid,
			"Archived_at must be empty unless status is archived",
		)
	}
	return nil
}

func (p *Project) validateMetrics() error {
	if p.Coverage < 0 || p.Coverage > 100 {
		return d.NewValidationError(
			CodeProjectMetricsInvalid,
			"Project coverage must be in range [0, 100]",
		)
	}
	if p.Readiness < 0 || p.Readiness > 100 {
		return d.NewValidationError(
			CodeProjectMetricsInvalid,
			"Project readiness must be in range [0, 100]",
		)
	}
	if !isRiskLevelValid(p.Risk) {
		return d.NewValidationError(
			CodeProjectRiskInvalid,
			"Project risk level is invalid",
		)
	}
	return nil
}

func (p *Project) validateDeadline() error {
	if p.DeadlineAt == nil {
		return nil
	}
	if p.DeadlineAt.IsZero() {
		return d.NewValidationError(
			CodeProjectDeadlineInvalid,
			"Project deadline cannot be zero",
		)
	}
	if p.DeadlineAt.Before(p.CreatedAt) {
		return d.NewValidationError(
			CodeProjectDeadlineInvalid,
			"Project deadline cannot be before created_at",
		)
	}
	return nil
}

func (p *Project) validateMetadata() error {
	for key, value := range p.Metadata {
		if strings.TrimSpace(key) == "" {
			return d.NewValidationError(
				CodeProjectInvalid,
				"Project metadata key cannot be empty",
			)
		}
		if len(key) > 100 {
			return d.NewValidationError(
				CodeProjectInvalid,
				"Project metadata key cannot be longer than 100 symbols",
			)
		}
		if len(value) > 1000 {
			return d.NewValidationError(
				CodeProjectInvalid,
				"Project metadata value cannot be longer than 1000 symbols",
			)
		}
	}
	return nil
}

func (p *Project) validateTime() error {
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"CreatedAt and UpdatedAt fields must be set",
		)
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before CreatedAt",
		)
	}
	if p.ArchivedAt != nil && p.ArchivedAt.Before(p.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"ArchivedAt cannot be before CreatedAt",
		)
	}
	if p.ArchivedAt != nil && p.UpdatedAt.Before(*p.ArchivedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before ArchivedAt",
		)
	}
	return nil
}

func isProjectPriorityValid(priority ProjectPriority) bool {
	switch priority {
	case ProjectPriorityLow, ProjectPriorityNormal, ProjectPriorityHigh:
		return true
	default:
		return false
	}
}

func isProjectStatusValid(status ProjectStatus) bool {
	switch status {
	case ProjectStatusActive, ProjectStatusPaused, ProjectStatusArchived:
		return true
	default:
		return false
	}
}

func canTransitionProjectStatus(current, next ProjectStatus) bool {
	switch current {
	case ProjectStatusActive:
		return next == ProjectStatusPaused || next == ProjectStatusArchived
	case ProjectStatusPaused:
		return next == ProjectStatusActive || next == ProjectStatusArchived
	case ProjectStatusArchived:
		return next == ProjectStatusActive
	default:
		return false
	}
}

func isRiskLevelValid(risk RiskLevel) bool {
	switch risk {
	case RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical:
		return true
	default:
		return false
	}
}

func cloneMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	return maps.Clone(metadata)
}

func copyTimePointer(src *time.Time) *time.Time {
	if src == nil {
		return nil
	}
	t := src.UTC()
	return &t
}
