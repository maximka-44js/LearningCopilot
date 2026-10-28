package entity

import (
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

type TaskStatus string // @name entity.TaskStatus

const (
	TaskStatusBacklog        TaskStatus = "backlog"
	TaskStatusLearning       TaskStatus = "learning"
	TaskStatusBaselinePassed TaskStatus = "baseline_passed"
	TaskStatusAtRisk         TaskStatus = "at_risk"
	TaskStatusStable         TaskStatus = "stable"
	TaskStatusExamReady      TaskStatus = "exam_ready"
	TaskStatusArchived       TaskStatus = "archived"
)

type CompetencyBlockType string // @name entity.CompetencyBlockType

const (
	CompetencyBlockComprehension CompetencyBlockType = "comprehension"
	CompetencyBlockMemorization  CompetencyBlockType = "memorization"
	CompetencyBlockPractice      CompetencyBlockType = "practice"
)

// CompetencyBlock captures block weight and current quality signals.
type CompetencyBlock struct {
	Type           CompetencyBlockType `json:"type"`
	Weight         int                 `json:"weight"`
	Mastery        float64             `json:"mastery"`
	LastAssessedAt *time.Time          `json:"last_assessed_at,omitempty"`
	DecayProfile   string              `json:"decay_profile,omitempty"`
	RiskState      RiskLevel           `json:"risk_state"`
}

type Task struct {
	ID         string            `json:"id"`
	StageID    string            `json:"stage_id"`
	Outcome    string            `json:"outcome"`
	Difficulty int               `json:"difficulty"`
	Priority   ProjectPriority   `json:"priority"`
	Policy     TaskPolicy        `json:"policy"`
	DeadlineAt *time.Time        `json:"deadline_at,omitempty"`
	Status     TaskStatus        `json:"status"`
	Blocks     []CompetencyBlock `json:"blocks"`
	Coverage   float64           `json:"coverage"`
	Readiness  float64           `json:"readiness"`
	Risk       RiskLevel         `json:"risk"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	ArchivedAt *time.Time        `json:"archived_at,omitempty"`
}

func NewTask(
	id string,
	stageID string,
	outcome string,
	difficulty int,
	priority ProjectPriority,
	policy TaskPolicy,
	deadlineAt *time.Time,
	blocks []CompetencyBlock,
) (Task, error) {
	now := time.Now().UTC()
	t := Task{
		ID:         id,
		StageID:    stageID,
		Outcome:    outcome,
		Difficulty: difficulty,
		Priority:   priority,
		Policy:     policy,
		DeadlineAt: copyTimePointer(deadlineAt),
		Status:     TaskStatusBacklog,
		Blocks:     cloneBlocks(blocks),
		Coverage:   0,
		Readiness:  0,
		Risk:       RiskLevelMedium,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := t.Validate(); err != nil {
		return Task{}, err
	}
	return t, nil
}

func (t *Task) Validate() error {
	if err := t.validateIDs(); err != nil {
		return err
	}
	if err := t.validateOutcome(); err != nil {
		return err
	}
	if err := t.validateDifficulty(); err != nil {
		return err
	}
	if err := t.validatePriority(); err != nil {
		return err
	}
	if err := t.validatePolicy(); err != nil {
		return err
	}
	if err := t.validateStatus(); err != nil {
		return err
	}
	if err := t.validateDeadline(); err != nil {
		return err
	}
	if err := t.validateBlocks(); err != nil {
		return err
	}
	if err := t.validateMetrics(); err != nil {
		return err
	}
	if err := t.validateTime(); err != nil {
		return err
	}
	return nil
}

func (t *Task) UpdatePolicy(policy TaskPolicy) error {
	t.Policy = policy
	t.UpdatedAt = time.Now().UTC()
	return t.Validate()
}

func (t *Task) UpdateDetails(
	outcome string,
	difficulty int,
	priority ProjectPriority,
	deadlineAt *time.Time,
) error {
	t.Outcome = outcome
	t.Difficulty = difficulty
	t.Priority = priority
	t.DeadlineAt = copyTimePointer(deadlineAt)
	t.UpdatedAt = time.Now().UTC()
	return t.Validate()
}

func (t *Task) ConfigureBlocks(blocks []CompetencyBlock) error {
	t.Blocks = cloneBlocks(blocks)
	t.UpdatedAt = time.Now().UTC()
	return t.Validate()
}

func (t *Task) UpdateMetrics(coverage, readiness float64, risk RiskLevel) error {
	t.Coverage = coverage
	t.Readiness = readiness
	t.Risk = risk
	t.UpdatedAt = time.Now().UTC()
	return t.Validate()
}

func (t *Task) RecordAssessmentSnapshot(
	blockType CompetencyBlockType,
	mastery float64,
	assessedAt time.Time,
	risk RiskLevel,
) error {
	index := slices.IndexFunc(t.Blocks, func(block CompetencyBlock) bool {
		return block.Type == blockType
	})
	if index < 0 {
		return d.NewValidationError(
			CodeTaskBlocksInvalid,
			"Task does not contain requested competency block",
		)
	}
	if mastery < 0 || mastery > 100 {
		return d.NewValidationError(
			CodeCompetencyMasteryInvalid,
			"Competency block mastery must be in range [0, 100]",
		)
	}
	if !isRiskLevelValid(risk) {
		return d.NewValidationError(
			CodeTaskRiskInvalid,
			"Task risk level is invalid",
		)
	}
	if assessedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"Assessment timestamp must be set",
		)
	}
	assessedAt = assessedAt.UTC()
	t.Blocks[index].Mastery = mastery
	t.Blocks[index].LastAssessedAt = &assessedAt
	t.Blocks[index].RiskState = risk
	t.UpdatedAt = time.Now().UTC()
	if err := t.Validate(); err != nil {
		return err
	}
	if t.Status == TaskStatusLearning {
		return t.TransitionStatus(TaskStatusBaselinePassed)
	}
	return nil
}

func (t *Task) TransitionStatus(next TaskStatus) error {
	if next == t.Status {
		return nil
	}
	if err := t.ValidateStatusTransition(next); err != nil {
		return err
	}
	t.Status = next
	if next == TaskStatusArchived {
		now := time.Now().UTC()
		t.ArchivedAt = &now
		t.UpdatedAt = now
		return nil
	}
	t.ArchivedAt = nil
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (t *Task) ValidateStatusTransition(next TaskStatus) error {
	if next == t.Status {
		return nil
	}
	if !isTaskStatusValid(next) {
		return d.NewValidationError(
			CodeTaskStatusInvalid,
			"Task status is invalid",
		)
	}
	if !canTransitionTaskStatus(t.Status, next) {
		return d.NewValidationError(
			CodeTaskStatusInvalid,
			"Task status transition is not allowed",
		)
	}
	if next == TaskStatusExamReady && !t.canBeExamReady() {
		return d.NewValidationError(
			CodeTaskStatusInvalid,
			"Task cannot be exam_ready before baseline is confirmed for each block",
		)
	}
	return nil
}

func (t *Task) validateIDs() error {
	if _, err := uuid.Parse(t.ID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Task id must be a valid UUID",
		)
	}
	if _, err := uuid.Parse(t.StageID); err != nil {
		return d.NewValidationError(
			CodeInvalidID,
			"Task stage_id must be a valid UUID",
		)
	}
	return nil
}

func (t *Task) validateOutcome() error {
	outcome := strings.TrimSpace(t.Outcome)
	if len(outcome) == 0 || len(outcome) > 2000 {
		return d.NewValidationError(
			CodeTaskOutcomeInvalid,
			"Task outcome must contain between 1 and 2000 symbols",
		)
	}
	return nil
}

func (t *Task) validateDifficulty() error {
	if t.Difficulty < 1 || t.Difficulty > 10 {
		return d.NewValidationError(
			CodeTaskDifficultyInvalid,
			"Task difficulty must be in range [1, 10]",
		)
	}
	return nil
}

func (t *Task) validatePriority() error {
	if !isProjectPriorityValid(t.Priority) {
		return d.NewValidationError(
			CodeTaskPriorityInvalid,
			"Task priority is invalid",
		)
	}
	return nil
}

func (t *Task) validatePolicy() error {
	if err := t.Policy.Validate(); err != nil {
		return err
	}
	return nil
}

func (t *Task) validateStatus() error {
	if !isTaskStatusValid(t.Status) {
		return d.NewValidationError(
			CodeTaskStatusInvalid,
			"Task status is invalid",
		)
	}
	if t.Status == TaskStatusArchived && t.ArchivedAt == nil {
		return d.NewValidationError(
			CodeTaskStatusInvalid,
			"Archived task must have archived_at set",
		)
	}
	if t.Status != TaskStatusArchived && t.ArchivedAt != nil {
		return d.NewValidationError(
			CodeTaskStatusInvalid,
			"Archived_at must be empty unless status is archived",
		)
	}
	return nil
}

func (t *Task) validateDeadline() error {
	if t.DeadlineAt == nil {
		return nil
	}
	if t.DeadlineAt.IsZero() {
		return d.NewValidationError(
			CodeTaskDeadlineInvalid,
			"Task deadline cannot be zero",
		)
	}
	if t.DeadlineAt.Before(t.CreatedAt) {
		return d.NewValidationError(
			CodeTaskDeadlineInvalid,
			"Task deadline cannot be before created_at",
		)
	}
	return nil
}

func (t *Task) validateBlocks() error {
	if len(t.Blocks) == 0 {
		return d.NewValidationError(
			CodeTaskBlocksInvalid,
			"Task must contain at least one competency block",
		)
	}
	weightsTotal := 0
	types := make(map[CompetencyBlockType]struct{}, len(t.Blocks))
	for _, block := range t.Blocks {
		if err := validateCompetencyBlock(block); err != nil {
			return err
		}
		if _, exists := types[block.Type]; exists {
			return d.NewValidationError(
				CodeTaskBlocksInvalid,
				"Task cannot contain duplicated competency block type",
			)
		}
		types[block.Type] = struct{}{}
		weightsTotal += block.Weight
	}
	if weightsTotal != 100 {
		return d.NewValidationError(
			CodeTaskBlocksInvalid,
			"Task competency block weights must sum to 100",
		)
	}
	return nil
}

func (t *Task) validateMetrics() error {
	if t.Coverage < 0 || t.Coverage > 100 {
		return d.NewValidationError(
			CodeTaskMetricsInvalid,
			"Task coverage must be in range [0, 100]",
		)
	}
	if t.Readiness < 0 || t.Readiness > 100 {
		return d.NewValidationError(
			CodeTaskMetricsInvalid,
			"Task readiness must be in range [0, 100]",
		)
	}
	if !isRiskLevelValid(t.Risk) {
		return d.NewValidationError(
			CodeTaskRiskInvalid,
			"Task risk level is invalid",
		)
	}
	return nil
}

func (t *Task) validateTime() error {
	if t.CreatedAt.IsZero() || t.UpdatedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"CreatedAt and UpdatedAt fields must be set",
		)
	}
	if t.UpdatedAt.Before(t.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before CreatedAt",
		)
	}
	if t.ArchivedAt != nil && t.ArchivedAt.Before(t.CreatedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"ArchivedAt cannot be before CreatedAt",
		)
	}
	if t.ArchivedAt != nil && t.UpdatedAt.Before(*t.ArchivedAt) {
		return d.NewValidationError(
			CodeTimeInvalid,
			"UpdatedAt cannot be before ArchivedAt",
		)
	}
	return nil
}

func (t *Task) canBeExamReady() bool {
	policy := t.Policy
	for _, block := range t.Blocks {
		if block.LastAssessedAt == nil {
			return false
		}
		if block.Mastery < policy.ExamReadyMinMastery {
			return false
		}
		if policy.RequireNonCriticalBlockRisk && block.RiskState == RiskLevelCritical {
			return false
		}
	}
	return true
}

func validateCompetencyBlock(block CompetencyBlock) error {
	if !isCompetencyBlockTypeValid(block.Type) {
		return d.NewValidationError(
			CodeCompetencyTypeInvalid,
			"Competency block type is invalid",
		)
	}
	if block.Weight < 1 || block.Weight > 100 {
		return d.NewValidationError(
			CodeCompetencyWeightInvalid,
			"Competency block weight must be in range [1, 100]",
		)
	}
	if block.Mastery < 0 || block.Mastery > 100 {
		return d.NewValidationError(
			CodeCompetencyMasteryInvalid,
			"Competency block mastery must be in range [0, 100]",
		)
	}
	if block.LastAssessedAt != nil && block.LastAssessedAt.IsZero() {
		return d.NewValidationError(
			CodeTimeInvalid,
			"Competency block last_assessed_at cannot be zero",
		)
	}
	if !isRiskLevelValid(block.RiskState) {
		return d.NewValidationError(
			CodeTaskRiskInvalid,
			"Competency block risk level is invalid",
		)
	}
	return nil
}

func isCompetencyBlockTypeValid(blockType CompetencyBlockType) bool {
	switch blockType {
	case CompetencyBlockComprehension, CompetencyBlockMemorization, CompetencyBlockPractice:
		return true
	default:
		return false
	}
}

func isTaskStatusValid(status TaskStatus) bool {
	switch status {
	case TaskStatusBacklog,
		TaskStatusLearning,
		TaskStatusBaselinePassed,
		TaskStatusAtRisk,
		TaskStatusStable,
		TaskStatusExamReady,
		TaskStatusArchived:
		return true
	default:
		return false
	}
}

func canTransitionTaskStatus(current, next TaskStatus) bool {
	switch current {
	case TaskStatusBacklog:
		return next == TaskStatusLearning || next == TaskStatusArchived
	case TaskStatusLearning:
		return next == TaskStatusBaselinePassed || next == TaskStatusAtRisk || next == TaskStatusArchived
	case TaskStatusBaselinePassed:
		return next == TaskStatusLearning || next == TaskStatusAtRisk || next == TaskStatusStable || next == TaskStatusArchived
	case TaskStatusAtRisk:
		return next == TaskStatusLearning || next == TaskStatusStable || next == TaskStatusArchived
	case TaskStatusStable:
		return next == TaskStatusLearning || next == TaskStatusAtRisk || next == TaskStatusExamReady || next == TaskStatusArchived
	case TaskStatusExamReady:
		return next == TaskStatusAtRisk || next == TaskStatusArchived
	case TaskStatusArchived:
		return next == TaskStatusBacklog
	default:
		return false
	}
}

func cloneBlocks(blocks []CompetencyBlock) []CompetencyBlock {
	if len(blocks) == 0 {
		return nil
	}
	cloned := make([]CompetencyBlock, 0, len(blocks))
	for _, block := range blocks {
		copied := block
		copied.LastAssessedAt = copyTimePointer(block.LastAssessedAt)
		cloned = append(cloned, copied)
	}
	return cloned
}
