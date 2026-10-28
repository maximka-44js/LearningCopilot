package v1

//dfdsg

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/maximka-44js/lc/internal/access/entity"
	"github.com/maximka-44js/lc/internal/access/usecase"
	appv1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

// @Summary     Create API key
// @Description Create a new API key for external REST integrations
// @ID          create-api-key
// @Tags        access
// @Accept      json
// @Produce     json
// @Param       request body     createAPIKeyRequest true "API key payload"
// @Success     201     {object} createAPIKeyResponse
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     403     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /access/api-keys [post]
func (r *V1) createAPIKey(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return appv1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body createAPIKeyRequest
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - access - createAPIKey")
		return appv1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - access - createAPIKey")
		return appv1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	out, err := r.apiKeys.Create(ctx.UserContext(), usecase.CreateAPIKeyInput{
		UserID:    userID,
		Name:      body.Name,
		Scopes:    body.Scopes,
		ExpiresAt: body.ExpiresAt,
	})
	if err != nil {
		if appErr, ok := err.(*d.AppError); ok {
			return appv1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return appv1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(createAPIKeyResponse{
		APIKey: mapAPIKey(out.APIKey),
		Secret: out.PlaintextKey,
	})
}

// @Summary     List API keys
// @Description List current user API keys without secret material
// @ID          list-api-keys
// @Tags        access
// @Accept      json
// @Produce     json
// @Success     200  {array} apiKeyResponse
// @Failure     401  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /access/api-keys [get]
func (r *V1) listAPIKeys(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return appv1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	keys, err := r.apiKeys.List(ctx.UserContext(), userID)
	if err != nil {
		if appErr, ok := err.(*d.AppError); ok {
			return appv1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return appv1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	resp := make([]apiKeyResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, mapAPIKey(k))
	}

	return ctx.Status(http.StatusOK).JSON(resp)
}

// @Summary     Revoke API key
// @Description Revoke API key immediately
// @ID          revoke-api-key
// @Tags        access
// @Accept      json
// @Produce     json
// @Param       id   path     string true "API key ID"
// @Success     200  {object} apiKeyResponse
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /access/api-keys/{id} [delete]
func (r *V1) revokeAPIKey(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return appv1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	apiKeyID := ctx.Params("id")
	key, err := r.apiKeys.Revoke(ctx.UserContext(), usecase.RevokeAPIKeyInput{UserID: userID, APIKeyID: apiKeyID})
	if err != nil {
		if appErr, ok := err.(*d.AppError); ok {
			return appv1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return appv1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(mapAPIKey(key))
}

// @Summary     Rotate API key
// @Description Create replacement key and revoke old key after grace window
// @ID          rotate-api-key
// @Tags        access
// @Accept      json
// @Produce     json
// @Param       id      path     string true "API key ID"
// @Param       request body     rotateAPIKeyRequest false "Optional rename for rotated key"
// @Success     200     {object} createAPIKeyResponse
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /access/api-keys/{id}/rotate [post]
func (r *V1) rotateAPIKey(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return appv1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body rotateAPIKeyRequest
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&body); err != nil {
			r.l.Error(err, "restapi - access - rotateAPIKey")
			return appv1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
		}
	}

	out, err := r.apiKeys.Rotate(ctx.UserContext(), usecase.RotateAPIKeyInput{
		UserID:   userID,
		APIKeyID: ctx.Params("id"),
		Name:     body.Name,
	})
	if err != nil {
		if appErr, ok := err.(*d.AppError); ok {
			return appv1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return appv1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(createAPIKeyResponse{
		APIKey: mapAPIKey(out.APIKey),
		Secret: out.PlaintextKey,
	})
}

func mapAPIKey(k entity.APIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		Scopes:     k.Scopes,
		ExpiresAt:  k.ExpiresAt,
		LastUsedAt: k.LastUsedAt,
		RevokedAt:  k.RevokedAt,
		Prefix:     k.KeyPrefix,
		CreatedAt:  k.CreatedAt,
	}
}
