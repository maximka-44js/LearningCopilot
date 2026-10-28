package v1

import (
	"context"
	"time"

	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/internal/projects/usecase"
)

type metricsContext struct {
	policies usecase.UserPolicies
	profiles map[entity.CompetencyBlockType]entity.DecayProfile
	now      time.Time
}

func (r *V1) newMetricsContext(ctx context.Context, ownerID, projectID string) (metricsContext, error) {
	policies, err := r.po.GetProjectPolicies(ctx, ownerID, projectID)
	if err != nil {
		return metricsContext{}, err
	}

	return metricsContext{
		policies: policies,
		profiles: entity.DefaultDecayProfiles(),
		now:      time.Now().UTC(),
	}, nil
}

func (r *V1) recalculateTaskMetrics(ctx context.Context, m metricsContext, ownerID, projectID, stageID, taskID string) (entity.Task, error) {
	return r.t.RecalculateReadiness(ctx, usecase.RecalculateTaskReadinessInput{
		OwnerID:    ownerID,
		ProjectID:  projectID,
		StageID:    stageID,
		TaskID:     taskID,
		At:         m.now,
		Profiles:   m.profiles,
		RiskPolicy: m.policies.Risk,
	})
}

func (r *V1) recalculateStageMetrics(ctx context.Context, m metricsContext, ownerID, projectID, stageID string) (entity.Stage, error) {
	tasks, err := r.t.ListByStage(ctx, ownerID, projectID, stageID)
	if err != nil {
		return entity.Stage{}, err
	}

	for _, task := range tasks {
		if _, err := r.recalculateTaskMetrics(ctx, m, ownerID, projectID, stageID, task.ID); err != nil {
			return entity.Stage{}, err
		}
	}

	return r.s.RecalculateReadiness(ctx, usecase.RecalculateStageReadinessInput{
		OwnerID:           ownerID,
		ProjectID:         projectID,
		StageID:           stageID,
		At:                m.now,
		AggregationPolicy: m.policies.Aggregation,
	})
}

func (r *V1) recalculateProjectMetrics(ctx context.Context, m metricsContext, ownerID, projectID string) (entity.Project, error) {
	stages, err := r.s.ListByProject(ctx, ownerID, projectID)
	if err != nil {
		return entity.Project{}, err
	}

	for _, stage := range stages {
		if _, err := r.recalculateStageMetrics(ctx, m, ownerID, projectID, stage.ID); err != nil {
			return entity.Project{}, err
		}
	}

	return r.p.RecalculateReadiness(ctx, usecase.RecalculateProjectReadinessInput{
		OwnerID:           ownerID,
		ProjectID:         projectID,
		At:                m.now,
		AggregationPolicy: m.policies.Aggregation,
	})
}
