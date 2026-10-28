package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/controller/restapi/v1/request"
	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/internal/projects/usecase"
)

type setTaskPolicyResponse entity.Task //nolint:unused // Used by Swagger annotations.

type setAggregationPolicyResponse entity.AggregationPolicy //nolint:unused // Used by Swagger annotations.

type setRiskPolicyResponse entity.RiskPolicy //nolint:unused // Used by Swagger annotations.

type setRecommendationPolicyResponse entity.RecommendationPolicy //nolint:unused // Used by Swagger annotations.

type setCalibrationPolicyResponse entity.CalibrationPolicy //nolint:unused // Used by Swagger annotations.

type getUserPoliciesResponse usecase.UserPolicies //nolint:unused // Used by Swagger annotations.

type getProjectPoliciesResponse usecase.UserPolicies //nolint:unused // Used by Swagger annotations.

// @Summary     Set task policy
// @Description Set policy for a specific task
// @ID          set-task-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       project_id path     string                true "Project ID"
// @Param       stage_id   path     string                true "Stage ID"
// @Param       task_id    path     string                true "Task ID"
// @Param       request    body     request.SetTaskPolicy true "Task policy payload"
// @Success     200        {object} v1.setTaskPolicyResponse
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/policy [patch]
func (r *V1) setTaskPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.SetTaskPolicyParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - setTaskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.SetTaskPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setTaskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setTaskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	task, err := r.po.SetTaskPolicy(ctx.UserContext(), userID, params.ProjectID, params.StageID, params.TaskID, body.Policy)
	if err != nil {
		r.l.Error(err, "restapi - v1 - setTaskPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

// @Summary     Set aggregation policy
// @Description Set user-level aggregation policy
// @ID          set-aggregation-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       request body     request.SetAggregationPolicy true "Aggregation policy payload"
// @Success     200     {object} v1.setAggregationPolicyResponse
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/policies/aggregation [patch]
func (r *V1) setAggregationPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.SetAggregationPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setAggregationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setAggregationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	err := r.po.SetAggregationPolicy(ctx.UserContext(), userID, body.Policy)
	if err != nil {
		r.l.Error(err, "restapi - v1 - setAggregationPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Set risk policy
// @Description Set user-level risk policy
// @ID          set-risk-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       request body     request.SetRiskPolicy true "Risk policy payload"
// @Success     200     {object} v1.setRiskPolicyResponse
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/policies/risk [patch]
func (r *V1) setRiskPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.SetRiskPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setRiskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setRiskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	err := r.po.SetRiskPolicy(ctx.UserContext(), userID, body.Policy)
	if err != nil {
		r.l.Error(err, "restapi - v1 - setRiskPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Set recommendation policy
// @Description Set user-level recommendation policy
// @ID          set-recommendation-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       request body     request.SetRecommendationPolicy true "Recommendation policy payload"
// @Success     200     {object} v1.setRecommendationPolicyResponse
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/policies/recommendation [patch]
func (r *V1) setRecommendationPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.SetRecommendationPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setRecommendationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setRecommendationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	err := r.po.SetRecommendationPolicy(ctx.UserContext(), userID, body.Policy)
	if err != nil {
		r.l.Error(err, "restapi - v1 - setRecommendationPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Set calibration policy
// @Description Set user-level calibration policy
// @ID          set-calibration-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       request body     request.SetCalibrationPolicy true "Calibration policy payload"
// @Success     200     {object} v1.setCalibrationPolicyResponse
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/policies/calibration [patch]
func (r *V1) setCalibrationPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.SetCalibrationPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setCalibrationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setCalibrationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	err := r.po.SetCalibrationPolicy(ctx.UserContext(), userID, body.Policy)
	if err != nil {
		r.l.Error(err, "restapi - v1 - setCalibrationPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Get project policies
// @Description Get effective policies for a project (project overrides or user defaults)
// @ID          get-project-policies
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       id path string true "Project ID"
// @Success     200 {object} v1.getProjectPoliciesResponse
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/policies [get]
func (r *V1) getProjectPolicies(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.ProjectPolicyParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getProjectPolicies")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	policies, err := r.po.GetProjectPolicies(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getProjectPolicies")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(policies)
}

// @Summary     Set project aggregation policy
// @Description Set project-level aggregation policy
// @ID          set-project-aggregation-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       id path string true "Project ID"
// @Param       request body request.SetAggregationPolicy true "Aggregation policy payload"
// @Success     200 {object} v1.setAggregationPolicyResponse
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/policies/aggregation [patch]
func (r *V1) setProjectAggregationPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.ProjectPolicyParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectAggregationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.SetAggregationPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectAggregationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectAggregationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.po.SetProjectAggregationPolicy(ctx.UserContext(), userID, params.ProjectID, body.Policy); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectAggregationPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Set project risk policy
// @Description Set project-level risk policy
// @ID          set-project-risk-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       id path string true "Project ID"
// @Param       request body request.SetRiskPolicy true "Risk policy payload"
// @Success     200 {object} v1.setRiskPolicyResponse
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/policies/risk [patch]
func (r *V1) setProjectRiskPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.ProjectPolicyParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRiskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.SetRiskPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRiskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRiskPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.po.SetProjectRiskPolicy(ctx.UserContext(), userID, params.ProjectID, body.Policy); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRiskPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Set project recommendation policy
// @Description Set project-level recommendation policy
// @ID          set-project-recommendation-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       id path string true "Project ID"
// @Param       request body request.SetRecommendationPolicy true "Recommendation policy payload"
// @Success     200 {object} v1.setRecommendationPolicyResponse
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/policies/recommendation [patch]
func (r *V1) setProjectRecommendationPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.ProjectPolicyParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRecommendationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.SetRecommendationPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRecommendationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRecommendationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.po.SetProjectRecommendationPolicy(ctx.UserContext(), userID, params.ProjectID, body.Policy); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectRecommendationPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Set project calibration policy
// @Description Set project-level calibration policy
// @ID          set-project-calibration-policy
// @Tags        policies
// @Accept      json
// @Produce     json
// @Param       id path string true "Project ID"
// @Param       request body request.SetCalibrationPolicy true "Calibration policy payload"
// @Success     200 {object} v1.setCalibrationPolicyResponse
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{id}/policies/calibration [patch]
func (r *V1) setProjectCalibrationPolicy(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.ProjectPolicyParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectCalibrationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.SetCalibrationPolicy
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectCalibrationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectCalibrationPolicy")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.po.SetProjectCalibrationPolicy(ctx.UserContext(), userID, params.ProjectID, body.Policy); err != nil {
		r.l.Error(err, "restapi - v1 - setProjectCalibrationPolicy")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(body.Policy)
}

// @Summary     Get user policies
// @Description Get effective user-level policies
// @ID          get-user-policies
// @Tags        policies
// @Accept      json
// @Produce     json
// @Success     200 {object} v1.getUserPoliciesResponse
// @Failure     401 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/policies [get]
func (r *V1) getUserPolicies(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	policies, err := r.po.GetUserPolicies(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getUserPolicies")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(policies)
}
