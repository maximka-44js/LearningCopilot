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

// @Summary     Create project
// @Description Create a new project for the current user
// @ID          create-project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       request body     request.CreateProject true "Project data"
// @Success     201     {object} entity.Project
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /projects [post]
func (r *V1) createProject(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.CreateProject
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - createProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - createProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	project, err := r.p.Create(ctx.UserContext(), usecase.CreateProjectInput{
		OwnerID:     userID,
		Name:        body.Name,
		Description: body.Description,
		DeadlineAt:  body.DeadlineAt,
		Priority:    body.Priority,
		Metadata:    body.Metadata,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - createProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(project)
}

// @Summary     Update project
// @Description Update project details by ID
// @ID          update-project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id      path     string                true "Project ID"
// @Param       request body     request.UpdateProject true "Project data"
// @Success     200     {object} entity.Project
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id} [put]
func (r *V1) updateProject(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.UpdateProjectParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - updateProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.UpdateProject
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	project, err := r.p.Update(ctx.UserContext(), usecase.UpdateProjectInput{
		OwnerID:     userID,
		ProjectID:   params.ProjectID,
		Name:        body.Name,
		Description: body.Description,
		DeadlineAt:  body.DeadlineAt,
		Priority:    body.Priority,
		Metadata:    body.Metadata,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - updateProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(project)
}

// @Summary     Get project by ID
// @Description Get project by ID for the current user
// @ID          get-project-by-id
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Project ID"
// @Success     200  {object} entity.Project
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id} [get]
func (r *V1) getProjectByID(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetProjectByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getProjectByID")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getProjectByID")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	project, err := r.recalculateProjectMetrics(ctx.UserContext(), metrics, userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getProjectByID")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(project)
}

// @Summary     List projects
// @Description List all projects for the current user
// @ID          list-projects
// @Tags        projects
// @Accept      json
// @Produce     json
// @Success     200  {array} entity.Project
// @Failure     401  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /projects [get]
func (r *V1) listProjectsByOwner(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	projects, err := r.p.ListByOwner(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listProjectsByOwner")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	updated := make([]entity.Project, 0, len(projects))
	for _, project := range projects {
		metrics, err := r.newMetricsContext(ctx.UserContext(), userID, project.ID)
		if err != nil {
			r.l.Error(err, "restapi - v1 - listProjectsByOwner")
			if appErr, ok := err.(*d.AppError); ok {
				return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
			}
			return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}

		updatedProject, err := r.recalculateProjectMetrics(ctx.UserContext(), metrics, userID, project.ID)
		if err != nil {
			r.l.Error(err, "restapi - v1 - listProjectsByOwner")
			if appErr, ok := err.(*d.AppError); ok {
				return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
			}
			return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
		updated = append(updated, updatedProject)
	}

	return ctx.Status(http.StatusOK).JSON(updated)
}

// @Summary     Delete project
// @Description Delete project by ID for the current user
// @ID          delete-project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Project ID"
// @Success     204 "No Content"
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id} [delete]
func (r *V1) deleteProject(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetProjectByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - deleteProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	if err := r.p.Delete(ctx.UserContext(), userID, params.ProjectID); err != nil {
		r.l.Error(err, "restapi - v1 - deleteProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Pause project
// @Description Pause project by ID
// @ID          pause-project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Project ID"
// @Success     200  {object} entity.Project
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/pause [patch]
func (r *V1) pauseProject(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetProjectByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - pauseProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	project, err := r.p.Pause(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - pauseProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(project)
}

// @Summary     Activate project
// @Description Activate project by ID
// @ID          activate-project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Project ID"
// @Success     200  {object} entity.Project
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/activate [patch]
func (r *V1) activateProject(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetProjectByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - activateProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	project, err := r.p.Activate(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - activateProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(project)
}

// @Summary     Archive project
// @Description Archive project by ID
// @ID          archive-project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Project ID"
// @Success     200  {object} entity.Project
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/archive [patch]
func (r *V1) archiveProject(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetProjectByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - archiveProject")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	project, err := r.p.Archive(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - archiveProject")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(project)
}

// @Summary     Recalculate project readiness
// @Description Recalculate project readiness by ID
// @ID          recalculate-project-readiness
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id      path     string                                 true "Project ID"
// @Param       request body     request.RecalculateProjectReadiness    false "Recalculation settings"
// @Success     200     {object} entity.Project
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/readiness/recalculate [post]
func (r *V1) recalculateProjectReadiness(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetProjectByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - recalculateProjectReadiness")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.RecalculateProjectReadiness
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&body); err != nil {
			r.l.Error(err, "restapi - v1 - recalculateProjectReadiness")
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

	project, err := r.p.RecalculateReadiness(ctx.UserContext(), usecase.RecalculateProjectReadinessInput{
		OwnerID:           userID,
		ProjectID:         params.ProjectID,
		At:                at,
		AggregationPolicy: aggregationPolicy,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - recalculateProjectReadiness")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(project)
}
