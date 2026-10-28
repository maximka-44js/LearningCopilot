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

// @Summary     Add task
// @Description Add a new task to a stage
// @ID          add-task
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string          true "Project ID"
// @Param       stage_id   path     string          true "Stage ID"
// @Param       request    body     request.AddTask true "Task data"
// @Success     201        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks [post]
func (r *V1) addTask(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.AddTaskParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - addTask")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.AddTask
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - addTask")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - addTask")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	policy := entity.DefaultTaskPolicy()
	if body.Policy != nil {
		policy = *body.Policy
	}

	blocks := make([]entity.CompetencyBlock, 0, len(body.Blocks))
	for _, block := range body.Blocks {
		blocks = append(blocks, entity.CompetencyBlock{
			Type:      block.Type,
			Weight:    block.Weight,
			Mastery:   0,
			RiskState: entity.RiskLevelMedium,
		})
	}

	task, err := r.t.Add(ctx.UserContext(), usecase.AddTaskInput{
		OwnerID:    userID,
		ProjectID:  params.ProjectID,
		StageID:    params.StageID,
		Outcome:    body.Outcome,
		Difficulty: body.Difficulty,
		Priority:   body.Priority,
		Policy:     policy,
		DeadlineAt: body.DeadlineAt,
		Blocks:     blocks,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - addTask")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(task)
}

// @Summary     Update task
// @Description Update task details by ID
// @ID          update-task
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string             true "Project ID"
// @Param       stage_id   path     string             true "Stage ID"
// @Param       task_id    path     string             true "Task ID"
// @Param       request    body     request.UpdateTask true "Task data"
// @Success     200        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id} [put]
func (r *V1) updateTask(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - updateTask")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.UpdateTask
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateTask")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateTask")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	task, err := r.t.Update(ctx.UserContext(), usecase.UpdateTaskInput{
		OwnerID:    userID,
		ProjectID:  params.ProjectID,
		StageID:    params.StageID,
		TaskID:     params.TaskID,
		Outcome:    body.Outcome,
		Difficulty: body.Difficulty,
		Priority:   body.Priority,
		DeadlineAt: body.DeadlineAt,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - updateTask")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

// @Summary     Get task by ID
// @Description Get task by ID for the current user
// @ID          get-task-by-id
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Param       task_id    path     string true "Task ID"
// @Success     200        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id} [get]
func (r *V1) getTaskByID(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getTaskByID")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getTaskByID")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	task, err := r.recalculateTaskMetrics(ctx.UserContext(), metrics, userID, params.ProjectID, params.StageID, params.TaskID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getTaskByID")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

// @Summary     List tasks by stage
// @Description List all tasks in a stage for the current user
// @ID          list-tasks-by-stage
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Success     200        {array}  entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks [get]
func (r *V1) listTasksByStage(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.AddTaskParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - listTasksByStage")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	metrics, err := r.newMetricsContext(ctx.UserContext(), userID, params.ProjectID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listTasksByStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	tasks, err := r.t.ListByStage(ctx.UserContext(), userID, params.ProjectID, params.StageID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listTasksByStage")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	updated := make([]entity.Task, 0, len(tasks))
	for _, task := range tasks {
		updatedTask, err := r.recalculateTaskMetrics(ctx.UserContext(), metrics, userID, params.ProjectID, params.StageID, task.ID)
		if err != nil {
			r.l.Error(err, "restapi - v1 - listTasksByStage")
			if appErr, ok := err.(*d.AppError); ok {
				return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
			}
			return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
		updated = append(updated, updatedTask)
	}

	return ctx.Status(http.StatusOK).JSON(updated)
}

// @Summary     Delete task
// @Description Delete task by ID for the current user
// @ID          delete-task
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Param       task_id    path     string true "Task ID"
// @Success     204 "No Content"
// @Failure     400 {object} d.AppError
// @Failure     401 {object} d.AppError
// @Failure     404 {object} d.AppError
// @Failure     500 {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id} [delete]
func (r *V1) deleteTask(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - deleteTask")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	if err := r.t.Delete(ctx.UserContext(), userID, params.ProjectID, params.StageID, params.TaskID); err != nil {
		r.l.Error(err, "restapi - v1 - deleteTask")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Configure task weights
// @Description Configure competency block weights for a task
// @ID          configure-task-weights
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string                        true "Project ID"
// @Param       stage_id   path     string                        true "Stage ID"
// @Param       task_id    path     string                        true "Task ID"
// @Param       request    body     request.ConfigureTaskWeights  true "Weight distribution"
// @Success     200        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/weights [patch]
func (r *V1) configureTaskWeights(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - configureTaskWeights")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.ConfigureTaskWeights
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - configureTaskWeights")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - configureTaskWeights")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	weights := entity.WeightDistribution{
		entity.CompetencyBlockComprehension: body.Weights.Comprehension,
		entity.CompetencyBlockMemorization:  body.Weights.Memorization,
		entity.CompetencyBlockPractice:      body.Weights.Practice,
	}

	task, err := r.t.ConfigureWeights(ctx.UserContext(), usecase.ConfigureTaskWeightsInput{
		OwnerID:   userID,
		ProjectID: params.ProjectID,
		StageID:   params.StageID,
		TaskID:    params.TaskID,
		Weights:   weights,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - configureTaskWeights")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

// @Summary     Start task session
// @Description Start a learning session for a task
// @ID          start-task-session
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Param       task_id    path     string true "Task ID"
// @Success     200        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/sessions/start [post]
func (r *V1) startTaskSession(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - startTaskSession")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	task, err := r.t.StartSession(ctx.UserContext(), usecase.StartTaskSessionInput{
		OwnerID:   userID,
		ProjectID: params.ProjectID,
		StageID:   params.StageID,
		TaskID:    params.TaskID,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - startTaskSession")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

// @Summary     Record assessment snapshot
// @Description Record assessment evidence for a task competency block
// @ID          record-assessment-snapshot
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string                            true "Project ID"
// @Param       stage_id   path     string                            true "Stage ID"
// @Param       task_id    path     string                            true "Task ID"
// @Param       request    body     request.RecordAssessmentSnapshot  true "Assessment payload"
// @Success     200        {object} usecase.RecordAssessmentSnapshotOutput
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/assessments [post]
func (r *V1) recordAssessmentSnapshot(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - recordAssessmentSnapshot")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.RecordAssessmentSnapshot
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - recordAssessmentSnapshot")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - recordAssessmentSnapshot")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	recordedAt := time.Now().UTC()
	if body.RecordedAt != nil {
		recordedAt = body.RecordedAt.UTC()
	}

	out, err := r.t.RecordAssessmentSnapshot(ctx.UserContext(), usecase.RecordAssessmentSnapshotInput{
		OwnerID:             userID,
		ProjectID:           params.ProjectID,
		StageID:             params.StageID,
		TaskID:              params.TaskID,
		BlockType:           body.BlockType,
		Source:              body.Source,
		RawScore:            body.RawScore,
		NormalizedScore:     body.NormalizedScore,
		ConfidenceBefore:    body.ConfidenceBefore,
		ConfidenceAfter:     body.ConfidenceAfter,
		RecordedAt:          recordedAt,
		TraceID:             body.TraceID,
		SourceReference:     body.SourceReference,
		SourceReferenceType: body.SourceReferenceType,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - recordAssessmentSnapshot")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(out)
}

// @Summary     Recalculate task readiness
// @Description Recalculate task readiness by ID
// @ID          recalculate-task-readiness
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string                           true "Project ID"
// @Param       stage_id   path     string                           true "Stage ID"
// @Param       task_id    path     string                           true "Task ID"
// @Param       request    body     request.RecalculateTaskReadiness false "Recalculation settings"
// @Success     200        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/readiness/recalculate [post]
func (r *V1) recalculateTaskReadiness(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - recalculateTaskReadiness")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.RecalculateTaskReadiness
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&body); err != nil {
			r.l.Error(err, "restapi - v1 - recalculateTaskReadiness")
			return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
		}
	}

	at := time.Now().UTC()
	if body.At != nil {
		at = body.At.UTC()
	}

	profiles := defaultTaskDecayProfiles()
	if body.Profiles != nil {
		profiles = body.Profiles
	}

	riskPolicy := entity.DefaultRiskPolicy()
	if body.RiskPolicy != nil {
		riskPolicy = *body.RiskPolicy
	}

	task, err := r.t.RecalculateReadiness(ctx.UserContext(), usecase.RecalculateTaskReadinessInput{
		OwnerID:    userID,
		ProjectID:  params.ProjectID,
		StageID:    params.StageID,
		TaskID:     params.TaskID,
		At:         at,
		Profiles:   profiles,
		RiskPolicy: riskPolicy,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - recalculateTaskReadiness")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

// @Summary     Generate task recommendations
// @Description Generate recommendations for a task
// @ID          generate-task-recommendations
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string                                true "Project ID"
// @Param       stage_id   path     string                                true "Stage ID"
// @Param       task_id    path     string                                true "Task ID"
// @Param       request    body     request.GenerateRecommendations       false "Recommendation settings"
// @Success     200        {array}  entity.Recommendation
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/recommendations/generate [post]
func (r *V1) generateTaskRecommendations(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - generateTaskRecommendations")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.GenerateRecommendations
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&body); err != nil {
			r.l.Error(err, "restapi - v1 - generateTaskRecommendations")
			return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
		}
	}

	recommendationPolicy := entity.DefaultRecommendationPolicy()
	if body.RecommendationPolicy != nil {
		recommendationPolicy = *body.RecommendationPolicy
	}

	calibrationPolicy := entity.DefaultCalibrationPolicy()
	if body.CalibrationPolicy != nil {
		calibrationPolicy = *body.CalibrationPolicy
	}

	recommendations, err := r.t.GenerateRecommendations(ctx.UserContext(), usecase.GenerateRecommendationsInput{
		OwnerID:              userID,
		ProjectID:            params.ProjectID,
		StageID:              params.StageID,
		TaskID:               params.TaskID,
		RecommendationPolicy: recommendationPolicy,
		CalibrationPolicy:    calibrationPolicy,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - generateTaskRecommendations")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(recommendations)
}

// @Summary     Mark task exam ready
// @Description Mark task as exam ready
// @ID          mark-task-exam-ready
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string true "Project ID"
// @Param       stage_id   path     string true "Stage ID"
// @Param       task_id    path     string true "Task ID"
// @Success     200        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/exam-ready [patch]
func (r *V1) markTaskExamReady(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - markTaskExamReady")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	task, err := r.t.MarkExamReady(ctx.UserContext(), usecase.MarkTaskExamReadyInput{
		OwnerID:   userID,
		ProjectID: params.ProjectID,
		StageID:   params.StageID,
		TaskID:    params.TaskID,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - markTaskExamReady")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

// @Summary     Transition task status
// @Description Transition task to a new status
// @ID          transition-task-status
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       project_id path     string                           true "Project ID"
// @Param       stage_id   path     string                           true "Stage ID"
// @Param       task_id    path     string                           true "Task ID"
// @Param       request    body     request.TransitionTaskStatus     true "Transition payload"
// @Success     200        {object} entity.Task
// @Failure     400        {object} d.AppError
// @Failure     401        {object} d.AppError
// @Failure     404        {object} d.AppError
// @Failure     500        {object} d.AppError
// @Security    BearerAuth
// @Router      /projects/{project_id}/stages/{stage_id}/tasks/{task_id}/status [patch]
func (r *V1) transitionTaskStatus(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.TaskIDParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - transitionTaskStatus")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if params.ProjectID == "" || params.StageID == "" || params.TaskID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.TransitionTaskStatus
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - transitionTaskStatus")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - transitionTaskStatus")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	task, err := r.t.TransitionStatus(ctx.UserContext(), usecase.TransitionTaskStatusInput{
		OwnerID:    userID,
		ProjectID:  params.ProjectID,
		StageID:    params.StageID,
		TaskID:     params.TaskID,
		NextStatus: body.NextStatus,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - transitionTaskStatus")
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(task)
}

func defaultTaskDecayProfiles() map[entity.CompetencyBlockType]entity.DecayProfile {
	return map[entity.CompetencyBlockType]entity.DecayProfile{
		entity.CompetencyBlockComprehension: {HalfLifeHours: 120},
		entity.CompetencyBlockMemorization:  {HalfLifeHours: 72},
		entity.CompetencyBlockPractice:      {HalfLifeHours: 96},
	}
}
