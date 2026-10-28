package entity

import (
	"math"
	"time"

	d "github.com/maximka-44js/lc/internal/app/domain"
)

type ReadinessScore float64

func NewReadinessScore(value float64) (ReadinessScore, error) {
	if value < 0 || value > 100 {
		return 0, d.NewValidationError(
			CodeReadinessInvalid,
			"Readiness score must be in range [0, 100]",
		)
	}
	return ReadinessScore(value), nil
}

type WeightDistribution map[CompetencyBlockType]int

func (wd WeightDistribution) Validate() error {
	if len(wd) == 0 {
		return d.NewValidationError(
			CodeWeightDistributionInvalid,
			"Weight distribution cannot be empty",
		)
	}
	total := 0
	for blockType, weight := range wd {
		if !isCompetencyBlockTypeValid(blockType) {
			return d.NewValidationError(
				CodeWeightDistributionInvalid,
				"Weight distribution has unsupported competency type",
			)
		}
		if weight < 1 || weight > 100 {
			return d.NewValidationError(
				CodeWeightDistributionInvalid,
				"Weight distribution value must be in range [1, 100]",
			)
		}
		total += weight
	}
	if total != 100 {
		return d.NewValidationError(
			CodeWeightDistributionInvalid,
			"Weight distribution must sum to 100",
		)
	}
	return nil
}

type DecayProfile struct {
	HalfLifeHours float64 `json:"half_life_hours"`
}

func DefaultDecayProfiles() map[CompetencyBlockType]DecayProfile {
	return map[CompetencyBlockType]DecayProfile{
		CompetencyBlockComprehension: {HalfLifeHours: 120},
		CompetencyBlockMemorization:  {HalfLifeHours: 72},
		CompetencyBlockPractice:      {HalfLifeHours: 96},
	}
}

func (dp DecayProfile) Validate() error {
	if dp.HalfLifeHours <= 0 {
		return d.NewValidationError(
			CodeDecayProfileInvalid,
			"Decay profile half_life_hours must be greater than 0",
		)
	}
	return nil
}

func (dp DecayProfile) Freshness(lastAssessedAt, now time.Time) float64 {
	if err := dp.Validate(); err != nil {
		return 0
	}
	if lastAssessedAt.IsZero() {
		return 0
	}
	now = now.UTC()
	lastAssessedAt = lastAssessedAt.UTC()
	if now.Before(lastAssessedAt) {
		return 1
	}
	deltaHours := now.Sub(lastAssessedAt).Hours()
	return math.Pow(2, -deltaHours/dp.HalfLifeHours)
}
