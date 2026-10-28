package v1

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/controller/restapi/v1/request"
	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/internal/projects/usecase"
)

// @Summary     Add stage
// @Description Add a new stage to a project
// @ID          add-stage
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string           true "Project ID"
// @Param       request    body     request.AddStage true "Stage data"
// @Success     201        {object} entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages [post]
func (r *V1) addStage(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.AddStageParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - addStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.AddStage
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - addStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - addStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	stage, err := r.s.Add(ctx.UserContext(), usecase.AddStageInput{
		OwnerID:    userID,
		ProjectID:  params.ProjectID,
		Name:       body.Name,
		Goal:       body.Goal,
		Order:      body.Order,
		DeadlineAt: body.DeadlineAt,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - addStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(stage)
}

// @Summary     Update stage
// @Description Update stage details by ID
// @ID          update-stage
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string              true "Project ID"
// @Param       stage_id   path     string              true "Stage ID"
// @Param       request    body     request.UpdateStage true "Stage data"
// @Success     200        {object} entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id} [put]
func (r *V1) updateStage(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.UpdateStageParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - updateStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.UpdateStage
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	stage, err := r.s.Update(ctx.UserContext(), usecase.UpdateStageInput{
		OwnerID:    userID,
		ProjectID:  params.ProjectID,
		StageID:    params.StageID,
		Name:       body.Name,
		Goal:       body.Goal,
		Order:      body.Order,
		DeadlineAt: body.DeadlineAt,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - updateStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(stage)
}

// @Summary     Get stage by ID
// @Description Get stage by ID for the current user
// @ID          get-stage-by-id
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Success     200        {object} entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id} [get]
func (r *V1) getStageByID(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetStageByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getStageByID")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getStageByID")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	stage, err := r.recalculateStageMetrics(ctx.UserContext(), metrics, userID, params.ProjectID, params.StageID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getStageByID")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(stage)
}

// @Summary     List stages by project
// @Description List all stages in a project for the current user
// @ID          list-stages-by-project
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Success     200        {array}  entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages [get]
func (r *V1) listStagesByProject(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.AddStageParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - listStagesByProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listStagesByProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	stages, err := r.s.ListByProject(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listStagesByProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	updated := make([]entity.Stage, 0, len(stages))
	for _, stage := range stages {
		updatedStage, err := r.recalculateStageMetrics(ctx.UserContext(), metrics, userID, params.ProjectID, stage.ID)
		if err != nil {
			r.l.Error(err, "restapi - v1 - listStagesByProject")
			if appErr, ok := err.(*d.AppError); ok {
				return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
			}
			return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
		updated = append(updated, updatedStage)
	}

	return ctx.Status(http.StatusOK).JSON(updated)
}

// @Summary     Delete stage
// @Description Delete stage by ID for the current user
// @ID          delete-stage
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Success     204 "No Content"
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id} [delete]
func (r *V1) deleteStage(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetStageByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - deleteStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	if err := r.s.Delete(ctx.UserContext(), userID, params.ProjectID, params.StageID); err != nil {
		r.l.Error(err, "restapi - v1 - deleteStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Pause stage
// @Description Pause stage by ID
// @ID          pause-stage
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Success     200        {object} entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/pause [patch]
func (r *V1) pauseStage(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetStageByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - pauseStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	stage, err := r.s.Pause(ctx.UserContext(), userID, params.ProjectID, params.StageID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - pauseStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(stage)
}

// @Summary     Activate stage
// @Description Activate stage by ID
// @ID          activate-stage
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Success     200        {object} entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/activate [patch]
func (r *V1) activateStage(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetStageByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - activateStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	stage, err := r.s.Activate(ctx.UserContext(), userID, params.ProjectID, params.StageID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - activateStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(stage)
}

// @Summary     Archive stage
// @Description Archive stage by ID
// @ID          archive-stage
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Success     200        {object} entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/archive [patch]
func (r *V1) archiveStage(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetStageByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - archiveStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	stage, err := r.s.Archive(ctx.UserContext(), userID, params.ProjectID, params.StageID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - archiveStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(stage)
}

// @Summary     Recalculate stage readiness
// @Description Recalculate stage readiness by ID
// @ID          recalculate-stage-readiness
// @Tags        stages
// @Accept      json
// @Produce     json
// @Param       project_id path     string                             true "Project ID"
// @Param       stage_id   path     string                             true "Stage ID"
// @Param       request    body     request.RecalculateStageReadiness  false "Recalculation settings"
// @Success     200        {object} entity.Stage
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/readiness/recalculate [post]
func (r *V1) recalculateStageReadiness(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetStageByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - recalculateStageReadiness")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.RecalculateStageReadiness
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&body); err != nil {
			r.l.Error(err, "restapi - v1 - recalculateStageReadiness")
			return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
		}
	}

	at := time.Now().UTC()
	if body.At != nil {
		at = body.At.UTC()
	}

	aggregationPolicy := entity.DefaultAggregationPolicy()
	if body.AggregationPolicy != nil {
		aggregationPolicy = *body.AggregationPolicy
	}

	stage, err := r.s.RecalculateReadiness(ctx.UserContext(), usecase.RecalculateStageReadinessInput{
		OwnerID:           userID,
		ProjectID:         params.ProjectID,
		StageID:           params.StageID,
		At:                at,
		AggregationPolicy: aggregationPolicy,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - recalculateStageReadiness")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(stage)
}
