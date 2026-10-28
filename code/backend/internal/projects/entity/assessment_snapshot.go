package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

type AssessmentSource string

const (
	AssessmentSourceDialog        AssessmentSource = "dialog"
	AssessmentSourceMiniCase      AssessmentSource = "mini_case"
	AssessmentSourceExternal      AssessmentSource = "external_system"
	AssessmentSourceManualConfirm AssessmentSource = "manual_confirmation"
)

// AssessmentSnapshot is append-only evidence of block evaluation.
type AssessmentSnapshot struct {
	ID                  string              `json:"id"`
	TaskID              string              `json:"task_id"`
	BlockType           CompetencyBlockType `json:"block_type"`
	Source              AssessmentSource    `json:"source"`
	RawScore            float64             `json:"raw_score"`
	NormalizedScore     float64             `json:"normalized_score"`
	ConfidenceBefore    float64             `json:"confidence_before"`
	ConfidenceAfter     float64             `json:"confidence_after"`
	RecordedAt          time.Time           `json:"recorded_at"`
	TraceID             string              `json:"trace_id,omitempty"`
	SourceReference     string              `json:"source_reference,omitempty"`
	SourceReferenceType string              `json:"source_reference_type,omitempty"`
}

func NewAssessmentSnapshot(
	id string,
	taskID string,
	blockType CompetencyBlockType,
	source AssessmentSource,
	rawScore float64,
	normalizedScore float64,
	confidenceBefore float64,
	confidenceAfter float64,
	recordedAt time.Time,
	traceID string,
	sourceReference string,
	sourceReferenceType string,
) (AssessmentSnapshot, error) {
	s := AssessmentSnapshot{
		ID:                  id,
		TaskID:              taskID,
		BlockType:           blockType,
		Source:              source,
		RawScore:            rawScore,
		NormalizedScore:     normalizedScore,
		ConfidenceBefore:    confidenceBefore,
		ConfidenceAfter:     confidenceAfter,
		RecordedAt:          recordedAt.UTC(),
		TraceID:             traceID,
		SourceReference:     sourceReference,
		SourceReferenceType: sourceReferenceType,
	}
	if err := s.Validate(); err != nil {
		return AssessmentSnapshot{}, err
	}
	return s, nil
}

func (s *AssessmentSnapshot) Validate() error {
	if _, err := uuid.Parse(s.ID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"AssessmentSnapshot id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(s.TaskID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"AssessmentSnapshot task_id must be a valid UUID",
		)
	}
	if !isCompetencyBlockTypeValid(s.BlockType) {
		return d.NewValidationError(
			CodeCompetencyTypeInvalid,
			"AssessmentSnapshot block_type is invalid",
		)
	}
	if !isAssessmentSourceValid(s.Source) {
		return d.NewValidationError(
			CodeAssessmentSourceInvalid,
			"AssessmentSnapshot source is invalid",
		)
	}
	if err := validateScoreRange(s.RawScore, "raw_score"); err != nil {
		return err
	}
	if err := validateScoreRange(s.NormalizedScore, "normalized_score"); err != nil {
		return err
	}
	if err := validateConfidenceRange(s.ConfidenceBefore, "confidence_before"); err != nil {
		return err
	}
	if err := validateConfidenceRange(s.ConfidenceAfter, "confidence_after"); err != nil {
		return err
	}
	if s.RecordedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"AssessmentSnapshot recorded_at must be set",
		)
	}
	if len(strings.TrimSpace(s.TraceID)) > 128 {
		return d.NewValidationError(
			CodeAssessmentSnapshotInvalid,
			"AssessmentSnapshot trace_id cannot be longer than 128 symbols",
		)
	}
	if len(strings.TrimSpace(s.SourceReference)) > 1024 {
		return d.NewValidationError(
			CodeAssessmentSnapshotInvalid,
			"AssessmentSnapshot source_reference cannot be longer than 1024 symbols",
		)
	}
	if len(strings.TrimSpace(s.SourceReferenceType)) > 64 {
		return d.NewValidationError(
			CodeAssessmentSnapshotInvalid,
			"AssessmentSnapshot source_reference_type cannot be longer than 64 symbols",
		)
	}
	return nil
}

func isAssessmentSourceValid(source AssessmentSource) bool {
	switch source {
	case AssessmentSourceDialog,
		AssessmentSourceMiniCase,
		AssessmentSourceExternal,
		AssessmentSourceManualConfirm:
		return true
	default:
		return false
	}
}

func validateScoreRange(value float64, field string) error {
	if value < 0 || value > 100 {
		return d.NewValidationError(
			CodeAssessmentScoreInvalid,
			"AssessmentSnapshot "+field+" must be in range [0, 100]",
		)
	}
	return nil
}

func validateConfidenceRange(value float64, field string) error {
	if value < 0 || value > 100 {
		return d.NewValidationError(
			CodeAssessmentConfidenceInvalid,
			"AssessmentSnapshot "+field+" must be in range [0, 100]",
		)
	}
	return nil
}
