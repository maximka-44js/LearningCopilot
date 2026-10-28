package entity

import (
	d "github.com/maximka-44js/lc/internal/app/domain"
)

type TaskPolicy struct {
	ExamReadyMinMastery         float64 `json:"exam_ready_min_mastery"`
	RequireNonCriticalBlockRisk bool    `json:"require_non_critical_block_risk"`
}

func DefaultTaskPolicy() TaskPolicy {
	return TaskPolicy{
		ExamReadyMinMastery:         70,
		RequireNonCriticalBlockRisk: true,
	}
}

func (p TaskPolicy) Validate() error {
	if p.ExamReadyMinMastery <= 0 || p.ExamReadyMinMastery > 100 {
		return d.NewValidationError(
			CodeTaskPolicyInvalid,
			"Task policy exam_ready_min_mastery must be in range (0, 100]",
		)
	}
	return nil
}

type AggregationPolicy struct {
	PriorityLowWeight    float64 `json:"priority_low_weight"`
	PriorityNormalWeight float64 `json:"priority_normal_weight"`
	PriorityHighWeight   float64 `json:"priority_high_weight"`

	Urgency24hWeight float64 `json:"urgency_24h_weight"`
	Urgency72hWeight float64 `json:"urgency_72h_weight"`
	Urgency7dWeight  float64 `json:"urgency_7d_weight"`
}

func DefaultAggregationPolicy() AggregationPolicy {
	return AggregationPolicy{
		PriorityLowWeight:    0.8,
		PriorityNormalWeight: 1.0,
		PriorityHighWeight:   1.2,
		Urgency24hWeight:     1.3,
		Urgency72hWeight:     1.2,
		Urgency7dWeight:      1.1,
	}
}

func (p AggregationPolicy) Validate() error {
	if p.PriorityLowWeight <= 0 || p.PriorityNormalWeight <= 0 || p.PriorityHighWeight <= 0 {
		return d.NewValidationError(
			CodeAggregationPolicyInvalid,
			"Aggregation policy priority weights must be greater than 0",
		)
	}
	if p.Urgency24hWeight <= 0 || p.Urgency72hWeight <= 0 || p.Urgency7dWeight <= 0 {
		return d.NewValidationError(
			CodeAggregationPolicyInvalid,
			"Aggregation policy urgency weights must be greater than 0",
		)
	}
	return nil
}

type RiskPolicy struct {
	BlockCriticalReadinessBelow float64 `json:"block_critical_readiness_below"`
	BlockHighReadinessBelow     float64 `json:"block_high_readiness_below"`
	BlockMediumReadinessBelow   float64 `json:"block_medium_readiness_below"`

	BlockCriticalMasteryBelow float64 `json:"block_critical_mastery_below"`
	BlockHighMasteryBelow     float64 `json:"block_high_mastery_below"`
	BlockMediumMasteryBelow   float64 `json:"block_medium_mastery_below"`

	TaskCriticalReadinessBelow float64 `json:"task_critical_readiness_below"`
	TaskHighReadinessBelow     float64 `json:"task_high_readiness_below"`
	TaskMediumReadinessBelow   float64 `json:"task_medium_readiness_below"`

	FalseConfidenceGap float64 `json:"false_confidence_gap"`
}

func DefaultRiskPolicy() RiskPolicy {
	return RiskPolicy{
		BlockCriticalReadinessBelow: 35,
		BlockHighReadinessBelow:     55,
		BlockMediumReadinessBelow:   75,
		BlockCriticalMasteryBelow:   40,
		BlockHighMasteryBelow:       55,
		BlockMediumMasteryBelow:     70,
		TaskCriticalReadinessBelow:  35,
		TaskHighReadinessBelow:      55,
		TaskMediumReadinessBelow:    75,
		FalseConfidenceGap:          25,
	}
}

func (p RiskPolicy) Validate() error {
	if !(p.BlockCriticalReadinessBelow < p.BlockHighReadinessBelow && p.BlockHighReadinessBelow < p.BlockMediumReadinessBelow) {
		return d.NewValidationError(
			CodeRiskPolicyInvalid,
			"Risk policy block readiness thresholds are inconsistent",
		)
	}
	if !(p.BlockCriticalMasteryBelow < p.BlockHighMasteryBelow && p.BlockHighMasteryBelow < p.BlockMediumMasteryBelow) {
		return d.NewValidationError(
			CodeRiskPolicyInvalid,
			"Risk policy block mastery thresholds are inconsistent",
		)
	}
	if !(p.TaskCriticalReadinessBelow < p.TaskHighReadinessBelow && p.TaskHighReadinessBelow < p.TaskMediumReadinessBelow) {
		return d.NewValidationError(
			CodeRiskPolicyInvalid,
			"Risk policy task readiness thresholds are inconsistent",
		)
	}
	if p.FalseConfidenceGap <= 0 || p.FalseConfidenceGap > 100 {
		return d.NewValidationError(
			CodeRiskPolicyInvalid,
			"Risk policy false_confidence_gap must be in range (0, 100]",
		)
	}
	return nil
}

type RecommendationPolicy struct {
	LowMasteryThreshold        float64 `json:"low_mastery_threshold"`
	AcceptableMasteryThreshold float64 `json:"acceptable_mastery_threshold"`
	CalibrationGapThreshold    float64 `json:"calibration_gap_threshold"`
	MaxRecommendations         int     `json:"max_recommendations"`
}

func DefaultRecommendationPolicy() RecommendationPolicy {
	return RecommendationPolicy{
		LowMasteryThreshold:        60,
		AcceptableMasteryThreshold: 65,
		CalibrationGapThreshold:    25,
		MaxRecommendations:         3,
	}
}

func (p RecommendationPolicy) Validate() error {
	if p.LowMasteryThreshold <= 0 || p.LowMasteryThreshold > 100 {
		return d.NewValidationError(
			CodeRecommendationPolicyInvalid,
			"Recommendation policy low_mastery_threshold must be in range (0, 100]",
		)
	}
	if p.AcceptableMasteryThreshold <= 0 || p.AcceptableMasteryThreshold > 100 {
		return d.NewValidationError(
			CodeRecommendationPolicyInvalid,
			"Recommendation policy acceptable_mastery_threshold must be in range (0, 100]",
		)
	}
	if p.LowMasteryThreshold >= p.AcceptableMasteryThreshold {
		return d.NewValidationError(
			CodeRecommendationPolicyInvalid,
			"Recommendation policy low_mastery_threshold must be below acceptable_mastery_threshold",
		)
	}
	if p.CalibrationGapThreshold <= 0 || p.CalibrationGapThreshold > 100 {
		return d.NewValidationError(
			CodeRecommendationPolicyInvalid,
			"Recommendation policy calibration_gap_threshold must be in range (0, 100]",
		)
	}
	if p.MaxRecommendations < 1 || p.MaxRecommendations > 10 {
		return d.NewValidationError(
			CodeRecommendationPolicyInvalid,
			"Recommendation policy max_recommendations must be in range [1, 10]",
		)
	}
	return nil
}

type CalibrationPolicy struct {
	GapThreshold      float64 `json:"gap_threshold"`
	PenaltyMultiplier float64 `json:"penalty_multiplier"`
}

func DefaultCalibrationPolicy() CalibrationPolicy {
	return CalibrationPolicy{
		GapThreshold:      25,
		PenaltyMultiplier: 0.7,
	}
}

func (p CalibrationPolicy) Validate() error {
	if p.GapThreshold <= 0 || p.GapThreshold > 100 {
		return d.NewValidationError(
			CodeCalibrationPolicyInvalid,
			"Calibration policy gap_threshold must be in range (0, 100]",
		)
	}
	if p.PenaltyMultiplier <= 0 || p.PenaltyMultiplier > 1 {
		return d.NewValidationError(
			CodeCalibrationPolicyInvalid,
			"Calibration policy penalty_multiplier must be in range (0, 1]",
		)
	}
	return nil
}
