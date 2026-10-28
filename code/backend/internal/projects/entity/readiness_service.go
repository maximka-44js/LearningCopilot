package entity

import (
	"math"
	"time"

	d "github.com/maximka-44js/lc/internal/app/domain"
)

type ReadinessService struct{}

func NewReadinessService() ReadinessService {
	return ReadinessService{}
}

func (s ReadinessService) RecalculateTaskReadiness(
	task Task,
	profiles map[CompetencyBlockType]DecayProfile,
	at time.Time,
) (ReadinessScore, error) {
	if err := task.Validate(); err != nil {
		return 0, err
	}
	if at.IsZero() {
		return 0, d.NewValidationError(
			CodeReadinessInvalid,
			"Readiness calculation time must be set",
		)
	}
	at = at.UTC()
	score := 0.0
	for _, block := range task.Blocks {
		profile, ok := profiles[block.Type]
		if !ok {
			return 0, d.NewValidationError(
				CodeDecayProfileInvalid,
				"Decay profile is missing for competency block type",
			)
		}
		if err := profile.Validate(); err != nil {
			return 0, err
		}
		freshness := 0.0
		if block.LastAssessedAt != nil {
			freshness = profile.Freshness(*block.LastAssessedAt, at)
		}
		blockScore := block.Mastery * freshness
		score += (float64(block.Weight) / 100.0) * blockScore
	}
	return NewReadinessScore(round2(score))
}

func (s ReadinessService) RecalculateTaskCoverage(task Task) (float64, error) {
	if err := task.Validate(); err != nil {
		return 0, err
	}

	score := 0.0
	for _, block := range task.Blocks {
		score += (float64(block.Weight) / 100.0) * block.Mastery
	}

	return round2(score), nil
}

func (s ReadinessService) RecalculateStageReadiness(stage Stage, tasks []Task) (ReadinessScore, error) {
	return s.RecalculateStageReadinessAt(stage, tasks, DefaultAggregationPolicy(), time.Now().UTC())
}

func (s ReadinessService) RecalculateStageReadinessAt(
	stage Stage,
	tasks []Task,
	policy AggregationPolicy,
	at time.Time,
) (ReadinessScore, error) {
	if err := stage.Validate(); err != nil {
		return 0, err
	}
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	if at.IsZero() {
		return 0, d.NewValidationError(
			CodeReadinessInvalid,
			"Readiness aggregation time must be set",
		)
	}
	at = at.UTC()
	if len(tasks) == 0 {
		return NewReadinessScore(0)
	}
	totalWeight := 0.0
	weightedReadiness := 0.0
	for _, task := range tasks {
		if err := task.Validate(); err != nil {
			return 0, err
		}
		weight := taskAggregationWeight(task, policy, at)
		totalWeight += weight
		weightedReadiness += task.Readiness * weight
	}
	if totalWeight == 0 {
		return NewReadinessScore(0)
	}
	return NewReadinessScore(round2(weightedReadiness / totalWeight))
}

func (s ReadinessService) RecalculateProjectReadiness(project Project, stages []Stage) (ReadinessScore, error) {
	return s.RecalculateProjectReadinessAt(project, stages, DefaultAggregationPolicy(), time.Now().UTC())
}

func (s ReadinessService) RecalculateProjectReadinessAt(
	project Project,
	stages []Stage,
	policy AggregationPolicy,
	at time.Time,
) (ReadinessScore, error) {
	if err := project.Validate(); err != nil {
		return 0, err
	}
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	if at.IsZero() {
		return 0, d.NewValidationError(
			CodeReadinessInvalid,
			"Readiness aggregation time must be set",
		)
	}
	at = at.UTC()
	if len(stages) == 0 {
		return NewReadinessScore(0)
	}
	totalWeight := 0.0
	weightedReadiness := 0.0
	for _, stage := range stages {
		if err := stage.Validate(); err != nil {
			return 0, err
		}
		weight := stageAggregationWeight(stage, policy, at)
		totalWeight += weight
		weightedReadiness += stage.Readiness * weight
	}
	if totalWeight == 0 {
		return NewReadinessScore(0)
	}
	return NewReadinessScore(round2(weightedReadiness / totalWeight))
}

func (s ReadinessService) RecalculateCoverageFromTasks(tasks []Task) (float64, error) {
	return s.RecalculateCoverageFromTasksAt(tasks, DefaultAggregationPolicy(), time.Now().UTC())
}

func (s ReadinessService) RecalculateCoverageFromTasksAt(
	tasks []Task,
	policy AggregationPolicy,
	at time.Time,
) (float64, error) {
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	if at.IsZero() {
		return 0, d.NewValidationError(
			CodeReadinessInvalid,
			"Coverage aggregation time must be set",
		)
	}
	at = at.UTC()
	if len(tasks) == 0 {
		return 0, nil
	}
	totalWeight := 0.0
	weightedCoverage := 0.0
	for _, task := range tasks {
		if err := task.Validate(); err != nil {
			return 0, err
		}
		taskCoverage, err := s.RecalculateTaskCoverage(task)
		if err != nil {
			return 0, err
		}
		weight := taskAggregationWeight(task, policy, at)
		totalWeight += weight
		weightedCoverage += taskCoverage * weight
	}
	if totalWeight == 0 {
		return 0, nil
	}
	return round2(weightedCoverage / totalWeight), nil
}

func taskAggregationWeight(task Task, policy AggregationPolicy, at time.Time) float64 {
	priorityWeight := policy.PriorityNormalWeight
	switch task.Priority {
	case ProjectPriorityLow:
		priorityWeight = policy.PriorityLowWeight
	case ProjectPriorityNormal:
		priorityWeight = policy.PriorityNormalWeight
	case ProjectPriorityHigh:
		priorityWeight = policy.PriorityHighWeight
	}

	urgencyWeight := 1.0
	if task.DeadlineAt != nil {
		hours := task.DeadlineAt.Sub(at).Hours()
		switch {
		case hours <= 24:
			urgencyWeight = policy.Urgency24hWeight
		case hours <= 72:
			urgencyWeight = policy.Urgency72hWeight
		case hours <= 168:
			urgencyWeight = policy.Urgency7dWeight
		}
	}

	return priorityWeight * urgencyWeight
}

func stageAggregationWeight(stage Stage, policy AggregationPolicy, at time.Time) float64 {
	weight := 1.0
	if stage.DeadlineAt != nil {
		hours := stage.DeadlineAt.Sub(at).Hours()
		switch {
		case hours <= 24:
			weight = policy.Urgency24hWeight
		case hours <= 72:
			weight = policy.Urgency72hWeight
		case hours <= 168:
			weight = policy.Urgency7dWeight
		}
	}
	return weight
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
