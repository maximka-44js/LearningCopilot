package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/controller/restapi/v1/request"
	_ "github.com/maximka-44js/lc/internal/storage/entity"
)

// @Summary     Create folder
// @Description Create a new folder for the current user
// @ID          create-folder
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       request body     request.CreateFolder true "Folder data"
// @Success     201     {object} entity.Folder
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/ [post]
func (r *V1) createFolder(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.CreateFolder

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - createFolder")

		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - createFolder")

		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	r.l.Debug("restapi - v1 - createFolder - userID: %s", userID)
	r.l.Debug("restapi - v1 - createFolder - body: %v", body)

	folder, err := r.fd.Create(ctx.UserContext(), userID, body.LibID, body.Name, body.ParentID)

	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
	}

	return ctx.Status(http.StatusCreated).JSON(folder)
}

// @Summary     Get folder by ID
// @Description Get a specific folder by ID for the current user
// @ID          get-folder-by-id
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Folder ID"
// @Success     200  {object} entity.Folder
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/{id} [get]
func (r *V1) getFolderByID(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetFolderByID

	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getLibraryByID")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	r.l.Debug("restapi - v1 - getFolderByID - userID: %s", userID)
	r.l.Debug("restapi - v1 - getFolderByID - params: %v", params)

	folder, err := r.fd.GetByID(ctx.UserContext(), userID, params.FdID)

	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(folder)
}

// @Summary     Update folder
// @Description Update folder name or parent ID for the current user
// @ID          update-folder
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id      path     string true "Folder ID"
// @Param       request body     request.UpdateFolder true "Folder data"
// @Success     200     {object} entity.Folder
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     403     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/{id} [put]
func (r *V1) updateFolder(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.UpdateFolderParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - updateFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.UpdateFolder
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	r.l.Debug("restapi - v1 - updateFolder - userID: %s", userID)
	r.l.Debug("restapi - v1 - updateFolder - params: %v", params)
	r.l.Debug("restapi - v1 - updateFolder - body: %v", body)

	folder, err := r.fd.Update(ctx.UserContext(), userID, params.FdID, body.Name, body.ParentID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(folder)
}

// @Summary     Copy folder
// @Description Copy folder with nested folders and files inside the same library
// @ID          copy-folder
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id      path     string true "Folder ID"
// @Param       request body     request.CopyFolder true "Copy folder request"
// @Success     201     {object} entity.Folder
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     403     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/{id}/copy [post]
func (r *V1) copyFolder(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.CopyFolderParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - copyFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		r.l.Error(err, "restapi - v1 - copyFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.CopyFolder
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - copyFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - copyFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	folder, err := r.fd.Copy(ctx.UserContext(), userID, params.FdID, body.TargetParentID, body.Name)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(folder)
}

// @Summary     Compress folder
// @Description Trigger ZIP compression task for a folder
// @ID          compress-folder
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Folder ID"
// @Success     202  {object} map[string]string
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/{id}/compress [post]
func (r *V1) compressFolder(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetFolderByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - compressFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		r.l.Error(err, "restapi - v1 - compressFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	if err := r.fd.RequestCompress(ctx.UserContext(), userID, params.FdID); err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusAccepted).JSON(map[string]string{"status": "queued"})
}

// @Summary     Extract folder archive
// @Description Trigger ZIP extraction task for a folder
// @ID          extract-folder
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Folder ID"
// @Success     202  {object} map[string]string
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/{id}/extract [post]
func (r *V1) extractFolder(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetFolderByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - extractFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		r.l.Error(err, "restapi - v1 - extractFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	if err := r.fd.RequestExtract(ctx.UserContext(), userID, params.FdID); err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusAccepted).JSON(map[string]string{"status": "queued"})
}

// @Summary     Delete folder
// @Description Delete a folder by ID for the current user
// @ID          delete-folder
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Folder ID"
// @Success     200  {object} entity.Folder
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/{id} [delete]
func (r *V1) deleteFolder(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.DeleteFolderParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - deleteFolder")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	folder, err := r.fd.Delete(ctx.UserContext(), userID, params.FdID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(folder)
}

// @Summary     Get folder contents
// @Description Get immediate children for a folder to render the tree view
// @ID          get-folder-contents
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Folder ID"
// @Success     200  {object} TreeContents
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/folders/{id}/tree [get]
func (r *V1) getFolderContents(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetFolderByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getFolderContents")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		r.l.Error(err, "restapi - v1 - getFolderContents")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	contents, err := r.fd.GetFolderContents(ctx.UserContext(), userID, params.FdID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(contents)
}
