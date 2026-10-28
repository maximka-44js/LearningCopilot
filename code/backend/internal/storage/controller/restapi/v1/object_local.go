package v1

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	"github.com/maximka-44js/lc/internal/storage/repo"
	storageLocal "github.com/maximka-44js/lc/internal/storage/repo/local"
)

func (r *V1) putObjectLocal(ctx *fiber.Ctx) error { //nolint:unused // Reserved for local object storage routes.
	store, ok := r.os.(*storageLocal.Storage)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusNotFound, "object endpoint is unavailable")
	}

	token := strings.TrimSpace(ctx.Query("token"))
	claims, err := store.ValidatePutToken(token)
	if err != nil {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "invalid upload token")
	}

	body := requestBodyReader(ctx)
	if err := store.WriteObject(ctx.UserContext(), claims.Key, ctx.Get("Content-Type"), body); err != nil {
		r.l.Error(err)
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "failed to store object")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func (r *V1) putObjectPartLocal(ctx *fiber.Ctx) error { //nolint:unused // Reserved for local object storage routes.
	store, ok := r.os.(*storageLocal.Storage)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusNotFound, "object endpoint is unavailable")
	}

	token := strings.TrimSpace(ctx.Query("token"))
	claims, err := store.ValidatePartToken(token)
	if err != nil {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "invalid upload token")
	}

	body := requestBodyReader(ctx)
	if err := store.WriteMultipartPart(ctx.UserContext(), claims.Key, claims.UploadID, claims.PartNumber, claims.Checksum, body); err != nil {
		r.l.Error(err)
		if errors.Is(err, repo.ErrStorageNotFound) {
			return v1.ErrorResponse(ctx, http.StatusNotFound, "upload not found")
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "failed to store multipart part")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func (r *V1) getObjectLocal(ctx *fiber.Ctx) error { //nolint:unused // Reserved for local object storage routes.
	store, ok := r.os.(*storageLocal.Storage)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusNotFound, "object endpoint is unavailable")
	}

	token := strings.TrimSpace(ctx.Query("token"))
	claims, err := store.ValidateGetToken(token)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getObjectLocal - invalid token")
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "invalid download token")
	}

	file, contentType, err := store.OpenObject(claims.Key)
	if err != nil {
		r.l.Error(err)
		if errors.Is(err, repo.ErrStorageNotFound) {
			return v1.ErrorResponse(ctx, http.StatusNotFound, "object not found")
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "failed to load object")
	}

	if strings.TrimSpace(contentType) != "" {
		ctx.Set("Content-Type", contentType)
	}

	return ctx.SendStream(file)
}

func requestBodyReader(ctx *fiber.Ctx) io.Reader { //nolint:unused // Used by reserved local object storage routes.
	if stream := ctx.Context().RequestBodyStream(); stream != nil {
		return stream
	}
	return bytes.NewReader(ctx.Body())
}
