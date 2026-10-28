package entity

type RiskService struct{}

func NewRiskService() RiskService {
	return RiskService{}
}

func (s RiskService) CalculateBlockRisk(block CompetencyBlock, nowReadiness float64) RiskLevel {
	return s.CalculateBlockRiskWithPolicy(block, nowReadiness, DefaultRiskPolicy())
}

func (s RiskService) CalculateBlockRiskWithPolicy(
	block CompetencyBlock,
	nowReadiness float64,
	policy RiskPolicy,
) RiskLevel {
	if err := policy.Validate(); err != nil {
		return RiskLevelCritical
	}
	if block.LastAssessedAt == nil {
		return RiskLevelCritical
	}
	if nowReadiness < policy.BlockCriticalReadinessBelow || block.Mastery < policy.BlockCriticalMasteryBelow {
		return RiskLevelCritical
	}
	if nowReadiness < policy.BlockHighReadinessBelow || block.Mastery < policy.BlockHighMasteryBelow {
		return RiskLevelHigh
	}
	if nowReadiness < policy.BlockMediumReadinessBelow || block.Mastery < policy.BlockMediumMasteryBelow {
		return RiskLevelMedium
	}
	return RiskLevelLow
}

func (s RiskService) CalculateTaskRisk(task Task) RiskLevel {
	return s.CalculateTaskRiskWithPolicy(task, DefaultRiskPolicy())
}

func (s RiskService) CalculateTaskRiskWithPolicy(task Task, policy RiskPolicy) RiskLevel {
	if err := policy.Validate(); err != nil {
		return RiskLevelCritical
	}
	if task.Readiness < policy.TaskCriticalReadinessBelow {
		return RiskLevelCritical
	}
	if task.Readiness < policy.TaskHighReadinessBelow {
		return RiskLevelHigh
	}
	if task.Readiness < policy.TaskMediumReadinessBelow {
		return RiskLevelMedium
	}
	return RiskLevelLow
}

func (s RiskService) DetectFalseConfidence(snapshot AssessmentSnapshot) bool {
	return s.DetectFalseConfidenceWithPolicy(snapshot, DefaultRiskPolicy())
}

func (s RiskService) DetectFalseConfidenceWithPolicy(
	snapshot AssessmentSnapshot,
	policy RiskPolicy,
) bool {
	if err := policy.Validate(); err != nil {
		return true
	}
	gap := snapshot.ConfidenceAfter - snapshot.NormalizedScore
	return gap >= policy.FalseConfidenceGap
}
