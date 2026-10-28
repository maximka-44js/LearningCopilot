package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/maximka-44js/lc/internal/access/entity"
	"github.com/maximka-44js/lc/internal/app/controller/restapi/middleware"
	"github.com/maximka-44js/lc/internal/projects/usecase"
	"github.com/maximka-44js/lc/pkg/jwt"
	"github.com/maximka-44js/lc/pkg/logger"
)

// NewRoutes configures Projects BC endpoints.
func NewRoutes(projectsGroup fiber.Router, p usecase.Project, s usecase.Stage, t usecase.Task, po usecase.Policy, rm usecase.ReadModel, jwtManager *jwt.Manager, apiKeyAuth middleware.APIKeyAuthenticator, l logger.Interface) {
	r := &V1{
		p:  p,
		s:  s,
		t:  t,
		po: po,
		rm: rm,
		l:  l,
		v:  validator.New(validator.WithRequiredStructEnabled()),
	}

	protected := projectsGroup.Group(
		"",
		middleware.Auth(jwtManager, l, apiKeyAuth),
		middleware.RequireDomainAccess(entity.ScopeDomainProjects),
	)

	policyGroup := protected.Group("/policies")
	{
		policyGroup.Get("", r.getUserPolicies)
		policyGroup.Patch("/aggregation", r.setAggregationPolicy)
		policyGroup.Patch("/risk", r.setRiskPolicy)
		policyGroup.Patch("/recommendation", r.setRecommendationPolicy)
		policyGroup.Patch("/calibration", r.setCalibrationPolicy)
	}

	projectGroup := protected.Group("")
	{
		projectGroup.Post("", r.createProject)
		projectGroup.Put("/:id", r.updateProject)
		projectGroup.Delete("/:id", r.deleteProject)
		projectGroup.Get("/:id", r.getProjectByID)
		projectGroup.Get("/:id/dashboard", r.getProjectDashboard)
		projectGroup.Get("/:id/policies", r.getProjectPolicies)
		projectGroup.Patch("/:id/policies/aggregation", r.setProjectAggregationPolicy)
		projectGroup.Patch("/:id/policies/risk", r.setProjectRiskPolicy)
		projectGroup.Patch("/:id/policies/recommendation", r.setProjectRecommendationPolicy)
		projectGroup.Patch("/:id/policies/calibration", r.setProjectCalibrationPolicy)
		projectGroup.Get("", r.listProjectsByOwner)
		projectGroup.Patch("/:id/pause", r.pauseProject)
		projectGroup.Patch("/:id/activate", r.activateProject)
		projectGroup.Patch("/:id/archive", r.archiveProject)
		projectGroup.Post("/:id/readiness/recalculate", r.recalculateProjectReadiness)
	}

	stageGroup := protected.Group("/:project_id/stages")
	{
		stageGroup.Post("", r.addStage)
		stageGroup.Put("/:stage_id", r.updateStage)
		stageGroup.Delete("/:stage_id", r.deleteStage)
		stageGroup.Get("/:stage_id", r.getStageByID)
		stageGroup.Get("/:stage_id/board", r.getStageBoard)
		stageGroup.Get("", r.listStagesByProject)
		stageGroup.Patch("/:stage_id/pause", r.pauseStage)
		stageGroup.Patch("/:stage_id/activate", r.activateStage)
		stageGroup.Patch("/:stage_id/archive", r.archiveStage)
		stageGroup.Post("/:stage_id/readiness/recalculate", r.recalculateStageReadiness)
	}

	taskGroup := protected.Group("/:project_id/stages/:stage_id/tasks")
	{
		taskGroup.Post("", r.addTask)
		taskGroup.Put("/:task_id", r.updateTask)
		taskGroup.Delete("/:task_id", r.deleteTask)
		taskGroup.Get("/:task_id", r.getTaskByID)
		taskGroup.Get("/:task_id/readiness", r.getTaskReadinessView)
		taskGroup.Get("", r.listTasksByStage)
		taskGroup.Patch("/:task_id/policy", r.setTaskPolicy)
		taskGroup.Patch("/:task_id/weights", r.configureTaskWeights)
		taskGroup.Post("/:task_id/sessions/start", r.startTaskSession)
		taskGroup.Post("/:task_id/assessments", r.recordAssessmentSnapshot)
		taskGroup.Post("/:task_id/readiness/recalculate", r.recalculateTaskReadiness)
		taskGroup.Post("/:task_id/recommendations/generate", r.generateTaskRecommendations)
		taskGroup.Patch("/:task_id/exam-ready", r.markTaskExamReady)
		taskGroup.Patch("/:task_id/status", r.transitionTaskStatus)
	}
}
