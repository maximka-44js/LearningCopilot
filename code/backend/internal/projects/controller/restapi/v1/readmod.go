package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/controller/restapi/v1/request"
	"github.com/maximka-44js/lc/internal/projects/usecase"
)

type getProjectDashboardResponse usecase.ProjectDashboardView //nolint:unused // Used by Swagger annotations.

type getStageBoardResponse usecase.StageBoardView //nolint:unused // Used by Swagger annotations.

type getTaskReadinessViewResponse usecase.TaskReadinessView //nolint:unused // Used by Swagger annotations.

// @Summary     Get project dashboard
// @Description Get dashboard read-model for a project
// @ID          get-project-dashboard
// @Tags        projects-read-model
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Project ID"
// @Success     200  {object} v1.getProjectDashboardResponse
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/dashboard [get]
func (r *V1) getProjectDashboard(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.ProjectDashboardParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getProjectDashboard")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getProjectDashboard")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	if _, err := r.recalculateProjectMetrics(ctx.UserContext(), metrics, userID, params.ProjectID); err != nil {
		r.l.Error(err, "restapi - v1 - getProjectDashboard")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	view, err := r.rm.ProjectDashboard(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getProjectDashboard")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(view)
}

// @Summary     Get stage board
// @Description Get board read-model for a project stage
// @ID          get-stage-board
// @Tags        projects-read-model
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Success     200        {object} v1.getStageBoardResponse
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/board [get]
func (r *V1) getStageBoard(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.StageBoardParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getStageBoard")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getStageBoard")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	if _, err := r.recalculateStageMetrics(ctx.UserContext(), metrics, userID, params.ProjectID, params.StageID); err != nil {
		r.l.Error(err, "restapi - v1 - getStageBoard")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	view, err := r.rm.StageBoard(ctx.UserContext(), userID, params.ProjectID, params.StageID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getStageBoard")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(view)
}

// @Summary     Get task readiness view
// @Description Get readiness read-model for a task
// @ID          get-task-readiness-view
// @Tags        projects-read-model
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Param       task_id    path     string true "Task ID"
// @Success     200        {object} v1.getTaskReadinessViewResponse
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/readiness [get]
func (r *V1) getTaskReadinessView(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskReadinessParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getTaskReadinessView")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getTaskReadinessView")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	if _, err := r.recalculateTaskMetrics(ctx.UserContext(), metrics, userID, params.ProjectID, params.StageID, params.TaskID); err != nil {
		r.l.Error(err, "restapi - v1 - getTaskReadinessView")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	view, err := r.rm.TaskReadiness(ctx.UserContext(), userID, params.ProjectID, params.StageID, params.TaskID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getTaskReadinessView")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(view)
}
