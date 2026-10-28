package read_model

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/internal/projects/repo"
	"github.com/maximka-44js/lc/internal/projects/usecase"
)

const (
	blockDeficitThreshold = 60.0
)

type UseCase struct {
	readModelRepo   repo.ReadModelRepo
	recommendations entity.RecommendationService
	readiness       entity.ReadinessService
}

var _ usecase.ReadModel = (*UseCase)(nil)

func NewReadModelUseCase(readModelRepo repo.ReadModelRepo) *UseCase {
	return &UseCase{
		readModelRepo:   readModelRepo,
		recommendations: entity.NewRecommendationService(),
		readiness:       entity.NewReadinessService(),
	}
}

func (uc *UseCase) ProjectDashboard(ctx context.Context, ownerID, projectID string) (usecase.ProjectDashboardView, error) {
	if err := validateReadModelIDInput(ownerID, projectID); err != nil {
		return usecase.ProjectDashboardView{}, err
	}

	now := time.Now().UTC()
	data, err := uc.readModelRepo.ProjectDashboard(ctx, ownerID, projectID, now)
	if err != nil {
		return usecase.ProjectDashboardView{}, err
	}

	riskByStage := make(map[string]entity.RiskLevel, len(data.Stages))
	for _, stage := range data.Stages {
		if stage.Status == entity.StageStatusActive {
			riskByStage[stage.ID] = stage.Risk
		}
	}

	critical := make([]usecase.CriticalItem, 0, len(data.CriticalToDeadline))
	for _, item := range data.CriticalToDeadline {
		critical = append(critical, usecase.CriticalItem{
			StageID:    item.StageID,
			TaskID:     item.TaskID,
			Reason:     item.Reason,
			DeadlineAt: item.DeadlineAt,
		})
	}

	return usecase.ProjectDashboardView{
		Project:            data.Project,
		Coverage:           data.Project.Coverage,
		ReadinessNow:       data.Project.Readiness,
		RiskByStage:        riskByStage,
		CriticalToDeadline: critical,
	}, nil
}

func (uc *UseCase) StageBoard(ctx context.Context, ownerID, projectID, stageID string) (usecase.StageBoardView, error) {
	if err := validateReadModelIDInput(ownerID, projectID, stageID); err != nil {
		return usecase.StageBoardView{}, err
	}

	now := time.Now().UTC()
	data, err := uc.readModelRepo.StageBoard(ctx, ownerID, projectID, stageID, now)
	if err != nil {
		return usecase.StageBoardView{}, err
	}

	deficits := make(map[string][]entity.CompetencyBlockType, len(data.Tasks))
	nextPriority := make(map[string]entity.ProjectPriority, len(data.Tasks))

	for _, task := range data.Tasks {
		deficits[task.ID] = detectTaskDeficits(task)
		nextPriority[task.ID] = deriveNextActionPriority(task)
	}

	return usecase.StageBoardView{
		Stage:              data.Stage,
		Tasks:              data.Tasks,
		BlockDeficits:      deficits,
		NextActionPriority: nextPriority,
	}, nil
}

func (uc *UseCase) TaskReadiness(ctx context.Context, ownerID, projectID, stageID, taskID string) (usecase.TaskReadinessView, error) {
	if err := validateReadModelIDInput(ownerID, projectID, stageID, taskID); err != nil {
		return usecase.TaskReadinessView{}, err
	}

	now := time.Now().UTC()
	data, err := uc.readModelRepo.TaskReadiness(ctx, ownerID, projectID, stageID, taskID, now)
	if err != nil {
		return usecase.TaskReadinessView{}, err
	}

	profiles := entity.DefaultDecayProfiles()
	blockScores := make([]usecase.BlockScoreNow, 0, len(data.Task.Blocks))
	lastAssessments := make(map[entity.CompetencyBlockType]*time.Time, len(data.Task.Blocks))

	for _, block := range data.Task.Blocks {
		freshness := 0.0
		if block.LastAssessedAt != nil {
			freshness = profiles[block.Type].Freshness(*block.LastAssessedAt, now)
		}
		scoreNow := block.Mastery * freshness
		blockScores = append(blockScores, usecase.BlockScoreNow{
			Type:      block.Type,
			Weight:    block.Weight,
			Mastery:   block.Mastery,
			Freshness: round2(freshness),
			ScoreNow:  round2(scoreNow),
			Risk:      block.RiskState,
		})
		lastAssessments[block.Type] = block.LastAssessedAt
	}

	sort.Slice(blockScores, func(i, j int) bool {
		return string(blockScores[i].Type) < string(blockScores[j].Type)
	})

	latestSnapshot := latestSnapshot(data.LatestSnapshots)
	recommendations, err := uc.recommendations.GenerateForTask(data.Task, latestSnapshot)
	if err != nil {
		return usecase.TaskReadinessView{}, err
	}

	forecast, err := uc.buildDropForecast(data.Task, profiles, now)
	if err != nil {
		return usecase.TaskReadinessView{}, err
	}

	return usecase.TaskReadinessView{
		Task:            data.Task,
		BlockScoresNow:  blockScores,
		LastAssessments: lastAssessments,
		DropForecast:    forecast,
		Recommendations: recommendations,
	}, nil
}

func (uc *UseCase) buildDropForecast(task entity.Task, profiles map[entity.CompetencyBlockType]entity.DecayProfile, now time.Time) ([]usecase.ForecastPoint, error) {
	horizons := []time.Duration{24 * time.Hour, 72 * time.Hour, 7 * 24 * time.Hour}
	forecast := make([]usecase.ForecastPoint, 0, len(horizons))
	policy := entity.DefaultRiskPolicy()

	for _, horizon := range horizons {
		pointTime := now.Add(horizon)
		readiness, err := uc.readiness.RecalculateTaskReadiness(task, profiles, pointTime)
		if err != nil {
			return nil, err
		}
		forecast = append(forecast, usecase.ForecastPoint{
			At:        pointTime,
			Readiness: float64(readiness),
			Risk:      riskFromReadiness(float64(readiness), policy),
		})
	}

	return forecast, nil
}

func validateReadModelIDInput(ids ...string) error {
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			return d.NewValidationError(entity.CodeInvalidID, "UUID is required")
		}
		if _, err := uuid.Parse(trimmed); err != nil {
			return d.NewValidationError(entity.CodeInvalidID, "UUID is invalid")
		}
	}
	return nil
}

func detectTaskDeficits(task entity.Task) []entity.CompetencyBlockType {
	deficits := make([]entity.CompetencyBlockType, 0, len(task.Blocks))
	for _, block := range task.Blocks {
		if block.Mastery < blockDeficitThreshold {
			deficits = append(deficits, block.Type)
		}
	}

	sort.Slice(deficits, func(i, j int) bool {
		return string(deficits[i]) < string(deficits[j])
	})
	return deficits
}

func deriveNextActionPriority(task entity.Task) entity.ProjectPriority {
	switch task.Risk {
	case entity.RiskLevelCritical, entity.RiskLevelHigh:
		return entity.ProjectPriorityHigh
	case entity.RiskLevelMedium:
		return entity.ProjectPriorityNormal
	default:
		return task.Priority
	}
}

func latestSnapshot(snapshots []entity.AssessmentSnapshot) *entity.AssessmentSnapshot {
	if len(snapshots) == 0 {
		return nil
	}

	latest := snapshots[0]
	for i := 1; i < len(snapshots); i++ {
		if snapshots[i].RecordedAt.After(latest.RecordedAt) {
			latest = snapshots[i]
		}
	}

	return &latest
}

func riskFromReadiness(readiness float64, policy entity.RiskPolicy) entity.RiskLevel {
	switch {
	case readiness < policy.TaskCriticalReadinessBelow:
		return entity.RiskLevelCritical
	case readiness < policy.TaskHighReadinessBelow:
		return entity.RiskLevelHigh
	case readiness < policy.TaskMediumReadinessBelow:
		return entity.RiskLevelMedium
	default:
		return entity.RiskLevelLow
	}
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
