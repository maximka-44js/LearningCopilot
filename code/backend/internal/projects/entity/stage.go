package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

type StageStatus string // @name entity.StageStatus

const (
	StageStatusActive   StageStatus = "active"
	StageStatusPaused   StageStatus = "paused"
	StageStatusArchived StageStatus = "archived"
)

// Stage represents a project phase within Project aggregate.
type Stage struct {
	ID         string      `json:"id"`
	ProjectID  string      `json:"project_id"`
	Name       string      `json:"name"`
	Goal       string      `json:"goal"`
	Order      int         `json:"order"`
	Status     StageStatus `json:"status"`
	DeadlineAt *time.Time  `json:"deadline_at,omitempty"`
	Coverage   float64     `json:"coverage"`
	Readiness  float64     `json:"readiness"`
	Risk       RiskLevel   `json:"risk"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	ArchivedAt *time.Time  `json:"archived_at,omitempty"`
}

func NewStage(
	id string,
	projectID string,
	name string,
	goal string,
	order int,
	deadlineAt *time.Time,
) (Stage, error) {
	now := time.Now().UTC()
	s := Stage{
		ID:         id,
		ProjectID:  projectID,
		Name:       name,
		Goal:       goal,
		Order:      order,
		Status:     StageStatusActive,
		DeadlineAt: copyTimePointer(deadlineAt),
		Coverage:   0,
		Readiness:  0,
		Risk:       RiskLevelMedium,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.Validate(); err != nil {
		return Stage{}, err
	}
	return s, nil
}

func (s *Stage) Validate() error {
	if err := s.validateIDs(); err != nil {
		return err
	}
	if err := s.validateName(); err != nil {
		return err
	}
	if err := s.validateGoal(); err != nil {
		return err
	}
	if err := s.validateOrder(); err != nil {
		return err
	}
	if err := s.validateStatus(); err != nil {
		return err
	}
	if err := s.validateDeadline(); err != nil {
		return err
	}
	if err := s.validateMetrics(); err != nil {
		return err
	}
	if err := s.validateTime(); err != nil {
		return err
	}
	return nil
}

func (s *Stage) UpdateDetails(name, goal string, order int, deadlineAt *time.Time) error {
	s.Name = name
	s.Goal = goal
	s.Order = order
	s.DeadlineAt = copyTimePointer(deadlineAt)
	s.UpdatedAt = time.Now().UTC()
	return s.Validate()
}

func (s *Stage) UpdateMetrics(coverage, readiness float64, risk RiskLevel) error {
	s.Coverage = coverage
	s.Readiness = readiness
	s.Risk = risk
	s.UpdatedAt = time.Now().UTC()
	return s.Validate()
}

func (s *Stage) Pause() error {
	return s.setStatus(StageStatusPaused)
}

func (s *Stage) Activate() error {
	return s.setStatus(StageStatusActive)
}

func (s *Stage) Archive() error {
	if err := s.setStatus(StageStatusArchived); err != nil {
		return err
	}
	now := time.Now().UTC()
	s.ArchivedAt = &now
	s.UpdatedAt = now
	return nil
}

func (s *Stage) setStatus(next StageStatus) error {
	if next == s.Status {
		return nil
	}
	if err := s.ValidateStatusTransition(next); err != nil {
		return err
	}
	s.Status = next
	if next != StageStatusArchived {
		s.ArchivedAt = nil
	}
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Stage) ValidateStatusTransition(next StageStatus) error {
	if next == s.Status {
		return nil
	}
	if !isStageStatusValid(next) {
		return d.NewValidationError(
			CodeStageStatusInvalid,
			"Stage status is invalid",
		)
	}
	if !canTransitionStageStatus(s.Status, next) {
		return d.NewValidationError(
			CodeStageStatusInvalid,
			"Stage status transition is not allowed",
		)
	}
	return nil
}

func (s *Stage) validateIDs() error {
	if _, err := uuid.Parse(s.ID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Stage id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(s.ProjectID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Stage project_id must be a valid UUID",
		)
	}
	return nil
}

func (s *Stage) validateName() error {
	name := strings.TrimSpace(s.Name)
	if len(name) == 0 || len(name) > 160 {
		return d.NewValidationError(
			CodeStageNameInvalid,
			"Stage name must contain between 1 and 160 symbols",
		)
	}
	return nil
}

func (s *Stage) validateGoal() error {
	goal := strings.TrimSpace(s.Goal)
	if len(goal) > 2000 {
		return d.NewValidationError(
			CodeStageGoalInvalid,
			"Stage goal cannot be longer than 2000 symbols",
		)
	}
	return nil
}

func (s *Stage) validateOrder() error {
	if s.Order < 1 {
		return d.NewValidationError(
			CodeStageOrderInvalid,
			"Stage order must be greater than or equal to 1",
		)
	}
	return nil
}

func (s *Stage) validateStatus() error {
	if !isStageStatusValid(s.Status) {
		return d.NewValidationError(
			CodeStageStatusInvalid,
			"Stage status is invalid",
		)
	}
	if s.Status == StageStatusArchived && s.ArchivedAt == nil {
		return d.NewValidationError(
			CodeStageStatusInvalid,
			"Archived stage must have archived_at set",
		)
	}
	if s.Status != StageStatusArchived && s.ArchivedAt != nil {
		return d.NewValidationError(
			CodeStageStatusInvalid,
			"Archived_at must be empty unless status is archived",
		)
	}
	return nil
}

func (s *Stage) validateDeadline() error {
	if s.DeadlineAt == nil {
		return nil
	}
	if s.DeadlineAt.IsZero() {
		return d.NewValidationError(
			CodeStageDeadlineInvalid,
			"Stage deadline cannot be zero",
		)
	}
	if s.DeadlineAt.Before(s.CreatedAt) {
		return d.NewValidationError(
			CodeStageDeadlineInvalid,
			"Stage deadline cannot be before created_at",
		)
	}
	return nil
}

func (s *Stage) validateMetrics() error {
	if s.Coverage < 0 || s.Coverage > 100 {
		return d.NewValidationError(
			CodeStageMetricsInvalid,
			"Stage coverage must be in range [0, 100]",
		)
	}
	if s.Readiness < 0 || s.Readiness > 100 {
		return d.NewValidationError(
			CodeStageMetricsInvalid,
			"Stage readiness must be in range [0, 100]",
		)
	}
	if !isRiskLevelValid(s.Risk) {
		return d.NewValidationError(
			CodeStageRiskInvalid,
			"Stage risk level is invalid",
		)
	}
	return nil
}

func (s *Stage) validateTime() error {
	if s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"CreatedAt and UpdatedAt fields must be set",
		)
	}
	if s.UpdatedAt.Before(s.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before CreatedAt",
		)
	}
	if s.ArchivedAt != nil && s.ArchivedAt.Before(s.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"ArchivedAt cannot be before CreatedAt",
		)
	}
	if s.ArchivedAt != nil && s.UpdatedAt.Before(*s.ArchivedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before ArchivedAt",
		)
	}
	return nil
}

func isStageStatusValid(status StageStatus) bool {
	switch status {
	case StageStatusActive, StageStatusPaused, StageStatusArchived:
		return true
	default:
		return false
	}
}

func canTransitionStageStatus(current, next StageStatus) bool {
	switch current {
	case StageStatusActive:
		return next == StageStatusPaused || next == StageStatusArchived
	case StageStatusPaused:
		return next == StageStatusActive || next == StageStatusArchived
	case StageStatusArchived:
		return next == StageStatusActive
	default:
		return false
	}
}
