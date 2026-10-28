package entity

import (
	"math"
	"slices"
	"time"

	d "github.com/maximka-44js/lc/internal/app/domain"
)

type Recommendation struct {
	TaskID       string              `json:"task_id"`
	BlockType    CompetencyBlockType `json:"block_type,omitempty"`
	Reason       string              `json:"reason"`
	NextActions  []string            `json:"next_actions"`
	PriorityHint ProjectPriority     `json:"priority_hint"`
}

func (r Recommendation) Validate() error {
	if r.TaskID == "" {
		return d.NewValidationError(
			CodeRecommendationInvalid,
			"Recommendation task_id cannot be empty",
		)
	}
	if len(r.NextActions) < 2 || len(r.NextActions) > 3 {
		return d.NewValidationError(
			CodeRecommendationInvalid,
			"Recommendation must contain 2-3 next actions",
		)
	}
	if !isProjectPriorityValid(r.PriorityHint) {
		return d.NewValidationError(
			CodeRecommendationInvalid,
			"Recommendation priority_hint is invalid",
		)
	}
	return nil
}

type RecommendationService struct{}

func NewRecommendationService() RecommendationService {
	return RecommendationService{}
}

func (s RecommendationService) GenerateForTask(
	task Task,
	latestSnapshot *AssessmentSnapshot,
) ([]Recommendation, error) {
	return s.GenerateForTaskWithPolicy(task, latestSnapshot, DefaultRecommendationPolicy())
}

func (s RecommendationService) GenerateForTaskWithPolicy(
	task Task,
	latestSnapshot *AssessmentSnapshot,
	policy RecommendationPolicy,
) ([]Recommendation, error) {
	if err := task.Validate(); err != nil {
		return nil, err
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	blocks := map[CompetencyBlockType]CompetencyBlock{}
	for _, block := range task.Blocks {
		blocks[block.Type] = block
	}

	if len(blocks) == 0 {
		return nil, d.NewValidationError(CodeRecommendationInvalid, "Task must contain competency blocks")
	}

	now := time.Now().UTC()
	urgencyFactor := taskUrgencyFactor(task.DeadlineAt, now)
	taskRiskFactor := riskFactor(task.Risk)

	var candidates []recommendationCandidate
	mem := blocks[CompetencyBlockMemorization]
	comp := blocks[CompetencyBlockComprehension]
	prac := blocks[CompetencyBlockPractice]

	if mem.Mastery < policy.LowMasteryThreshold && comp.Mastery >= policy.AcceptableMasteryThreshold {
		candidates = append(candidates, newCandidate(task.ID, CompetencyBlockMemorization,
			"Low memorization with acceptable comprehension",
			[]string{"Export cards for fact repetition", "Run short retention check in 24 hours"},
			mem, policy, taskRiskFactor, urgencyFactor))
	} else if mem.Mastery < policy.AcceptableMasteryThreshold && mem.RiskState >= RiskLevelHigh {
		candidates = append(candidates, newCandidate(task.ID, CompetencyBlockMemorization,
			"High risk memorization drop",
			[]string{"Export cards for fact repetition", "Run short retention check in 24 hours"},
			mem, policy, taskRiskFactor, urgencyFactor))
	}

	if comp.Mastery < policy.LowMasteryThreshold && mem.Mastery >= policy.AcceptableMasteryThreshold {
		candidates = append(candidates, newCandidate(task.ID, CompetencyBlockComprehension,
			"Low comprehension with acceptable memorization",
			[]string{"Run why/what-if dialogue", "Retest causal links with mini-case"},
			comp, policy, taskRiskFactor, urgencyFactor))
	} else if comp.Mastery < policy.AcceptableMasteryThreshold && comp.RiskState >= RiskLevelHigh {
		candidates = append(candidates, newCandidate(task.ID, CompetencyBlockComprehension,
			"High risk comprehension drop",
			[]string{"Run why/what-if dialogue", "Retest causal links with mini-case"},
			comp, policy, taskRiskFactor, urgencyFactor))
	}

	if prac.Mastery < policy.LowMasteryThreshold {
		candidates = append(candidates, newCandidate(task.ID, CompetencyBlockPractice,
			"Low practical performance",
			[]string{"Solve a sequence of applied tasks with increasing complexity", "Review and document typical mistakes"},
			prac, policy, taskRiskFactor, urgencyFactor))
	} else if prac.Mastery < policy.AcceptableMasteryThreshold && prac.RiskState >= RiskLevelHigh {
		candidates = append(candidates, newCandidate(task.ID, CompetencyBlockPractice,
			"High risk practice drop",
			[]string{"Solve a sequence of applied tasks with increasing complexity", "Review and document typical mistakes"},
			prac, policy, taskRiskFactor, urgencyFactor))
	}

	if latestSnapshot != nil {
		if err := latestSnapshot.Validate(); err != nil {
			return nil, err
		}
		if calibrationNeeded(*latestSnapshot, policy.CalibrationGapThreshold) {
			candidates = append(candidates, recommendationCandidate{
				rec: Recommendation{
					TaskID:       task.ID,
					Reason:       "High confidence-performance gap",
					NextActions:  []string{"Start calibration session", "Reduce self-assessment weight in next decision cycle"},
					PriorityHint: ProjectPriorityHigh,
				},
				score: 1.4 * urgencyFactor,
			})
		}
	}

	if len(candidates) == 0 {
		candidates = append(candidates, recommendationCandidate{
			rec: Recommendation{
				TaskID:       task.ID,
				Reason:       "Task is stable; maintain readiness",
				NextActions:  []string{"Schedule next check by risk forecast", "Keep one mixed practice iteration"},
				PriorityHint: ProjectPriorityNormal,
			},
			score: 0.2 * urgencyFactor,
		})
	}

	slices.SortStableFunc(candidates, func(a, b recommendationCandidate) int {
		if a.score == b.score {
			return 0
		}
		if a.score > b.score {
			return -1
		}
		return 1
	})

	limit := policy.MaxRecommendations
	if limit <= 0 {
		limit = 1
	}
	if limit > len(candidates) {
		limit = len(candidates)
	}

	recommendations := make([]Recommendation, 0, limit)
	for i := 0; i < limit; i++ {
		recommendations = append(recommendations, candidates[i].rec)
	}

	for _, recommendation := range recommendations {
		if err := recommendation.Validate(); err != nil {
			return nil, err
		}
	}

	return recommendations, nil
}

type CalibrationService struct{}

func NewCalibrationService() CalibrationService {
	return CalibrationService{}
}

func (s CalibrationService) ConfidenceGap(snapshot AssessmentSnapshot) (float64, error) {
	if err := snapshot.Validate(); err != nil {
		return 0, err
	}
	return snapshot.ConfidenceAfter - snapshot.NormalizedScore, nil
}

func (s CalibrationService) NeedsCalibration(snapshot AssessmentSnapshot) (bool, error) {
	return s.NeedsCalibrationWithPolicy(snapshot, DefaultCalibrationPolicy())
}

func (s CalibrationService) NeedsCalibrationWithPolicy(
	snapshot AssessmentSnapshot,
	policy CalibrationPolicy,
) (bool, error) {
	if err := policy.Validate(); err != nil {
		return false, err
	}
	gap, err := s.ConfidenceGap(snapshot)
	if err != nil {
		return false, err
	}
	return gap >= policy.GapThreshold, nil
}

func (s CalibrationService) SubjectiveWeightMultiplier(snapshot AssessmentSnapshot) (float64, error) {
	return s.SubjectiveWeightMultiplierWithPolicy(snapshot, DefaultCalibrationPolicy())
}

func (s CalibrationService) SubjectiveWeightMultiplierWithPolicy(
	snapshot AssessmentSnapshot,
	policy CalibrationPolicy,
) (float64, error) {
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	needs, err := s.NeedsCalibrationWithPolicy(snapshot, policy)
	if err != nil {
		return 0, err
	}
	if needs {
		return policy.PenaltyMultiplier, nil
	}
	return 1.0, nil
}

func calibrationNeeded(snapshot AssessmentSnapshot, gapThreshold float64) bool {
	gap := snapshot.ConfidenceAfter - snapshot.NormalizedScore
	return gap >= gapThreshold
}

type recommendationCandidate struct {
	rec   Recommendation
	score float64
}

func newCandidate(
	taskID string,
	blockType CompetencyBlockType,
	reason string,
	actions []string,
	block CompetencyBlock,
	policy RecommendationPolicy,
	taskRiskFactor float64,
	urgencyFactor float64,
) recommendationCandidate {
	weightFactor := blockWeightFactor(block.Weight)
	deficit := math.Max(0, policy.AcceptableMasteryThreshold-block.Mastery)
	riskFactor := riskFactor(block.RiskState)
	score := (deficit / 100.0) * weightFactor * riskFactor * taskRiskFactor * urgencyFactor
	priority := ProjectPriorityNormal
	if score >= 0.35 || block.RiskState >= RiskLevelHigh {
		priority = ProjectPriorityHigh
	}

	return recommendationCandidate{
		rec: Recommendation{
			TaskID:       taskID,
			BlockType:    blockType,
			Reason:       reason,
			NextActions:  actions,
			PriorityHint: priority,
		},
		score: score,
	}
}

func blockWeightFactor(weight int) float64 {
	if weight <= 0 {
		return 0.01
	}
	f := float64(weight) / 100.0
	if f < 0.01 {
		return 0.01
	}
	return f
}

func riskFactor(level RiskLevel) float64 {
	switch level {
	case RiskLevelCritical:
		return 1.4
	case RiskLevelHigh:
		return 1.25
	case RiskLevelMedium:
		return 1.1
	default:
		return 1.0
	}
}

func taskUrgencyFactor(deadlineAt *time.Time, now time.Time) float64 {
	if deadlineAt == nil {
		return 1.0
	}
	hoursToDeadline := deadlineAt.UTC().Sub(now).Hours()
	switch {
	case hoursToDeadline <= 24:
		return 1.3
	case hoursToDeadline <= 72:
		return 1.2
	case hoursToDeadline <= 168:
		return 1.1
	default:
		return 1.0
	}
}
