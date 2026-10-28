package v1

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/controller/restapi/v1/request"
	_ "github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/internal/storage/repo"
	usecase "github.com/maximka-44js/lc/internal/storage/usecase"
)

// @Summary     Upload file (server-side)
// @Description Upload a file directly to backend using multipart/form-data and finalize it in one request
// @ID          upload-file-direct
// @Tags        files
// @Accept      multipart/form-data
// @Produce     json
// @Param       library_id         formData string true  "Library ID"
// @Param       folder_id          formData string false "Folder ID"
// @Param       file               formData file   true  "File payload"
// @Param       checksum_declared  formData string false "Optional SHA-256 checksum to verify after upload"
// @Success     201  {object} entity.File
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/upload [post]
func (r *V1) uploadFile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	contentType := strings.ToLower(strings.TrimSpace(ctx.Get("Content-Type")))
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return v1.ErrorResponse(ctx, http.StatusUnsupportedMediaType, "content type must be multipart/form-data")
	}

	libraryID := strings.TrimSpace(ctx.FormValue("library_id"))
	if libraryID == "" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "library_id is required")
	}

	fh, err := ctx.FormFile("file")
	if err != nil || fh == nil {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "file is required")
	}

	fileBody, err := fh.Open()
	if err != nil {
		r.l.Error(err, "restapi - v1 - uploadFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "failed to open uploaded file")
	}
	defer func() {
		_ = fileBody.Close()
	}()

	mimeType := strings.TrimSpace(fh.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploadedFile, err := r.f.Upload(ctx.UserContext(), userID, usecase.UploadFileInput{
		LibraryID:        libraryID,
		FolderID:         strings.TrimSpace(ctx.FormValue("folder_id")),
		Name:             fh.Filename,
		MimeType:         mimeType,
		Size:             fh.Size,
		ChecksumDeclared: strings.TrimSpace(ctx.FormValue("checksum_declared")),
		Body:             fileBody,
	})
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(uploadedFile)
}

// @Summary     Update file
// @Description Update file name or folder for the current user
// @ID          update-file
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id      path     string true "File ID"
// @Param       request body     request.UpdateFile true "File data"
// @Success     200     {object} entity.File
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     403     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id} [put]
func (r *V1) updateFile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.UpdateFileParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - updateFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.UpdateFile
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	file, err := r.f.Update(ctx.UserContext(), userID, params.FileID, body.Name, strings.TrimSpace(body.FolderID))
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(file)
}

// @Summary     Copy file
// @Description Copy file inside the same library
// @ID          copy-file
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id      path     string true "File ID"
// @Param       request body     request.CopyFile true "Copy file request"
// @Success     201     {object} entity.File
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     403     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id}/copy [post]
func (r *V1) copyFile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.CopyFileParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - copyFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		r.l.Error(err, "restapi - v1 - copyFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.CopyFile
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - copyFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - copyFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	file, err := r.f.Copy(ctx.UserContext(), userID, params.FileID, strings.TrimSpace(body.TargetFolderID), strings.TrimSpace(body.Name))
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(file)
}

// @Summary     Init upload
// @Description Initialize file upload (presign or multipart)
// @ID          init-upload
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       request body     request.InitUpload true "Init upload"
// @Success     201     {object} usecase.InitUploadResult
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     403     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/init [post]
func (r *V1) initUpload(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var body request.InitUpload
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - initUpload")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - initUpload")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	res, err := r.f.InitUpload(ctx.UserContext(), userID, body.LibraryID, body.FolderID, body.Name, body.MimeType, body.Size, body.ChecksumDeclared)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(res)
}

// @Summary     Presign upload part
// @Description Get presigned URL for multipart upload part
// @ID          presign-upload-part
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id   path     string true "File ID"
// @Param       request body     request.PresignPart true "Part info"
// @Success     200  {object} map[string]string
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id}/parts [post]
func (r *V1) presignUploadPart(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params struct {
		ID string `params:"id"`
	}
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - presignUploadPart")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid path param")
	}

	var body request.PresignPart
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - presignUploadPart")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - presignUploadPart")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	url, err := r.f.PresignUploadPart(ctx.UserContext(), userID, params.ID, body.UploadID, body.ChecksumSHA256, body.PartNumber, body.Size)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(map[string]string{"url": url})
}

// @Summary     Confirm upload
// @Description Confirm upload after client finished uploading parts; perform integrity checks and finalize file
// @ID          confirm-upload
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id   path     string true "File ID"
// @Param       request body     request.ConfirmUpload true "Confirm upload"
// @Success     200  {object} entity.File
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id}/complete [post]
func (r *V1) completeUpload(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params struct {
		ID string `params:"id"`
	}
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - completeUpload")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid path param")
	}

	var body request.ConfirmUpload
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - completeUpload")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - completeUpload")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	completedParts := make([]repo.CompletedPart, 0, len(body.CompletedParts))
	for _, part := range body.CompletedParts {
		completedParts = append(completedParts, repo.CompletedPart{ETag: part.ETag, PartNumber: part.PartNumber})
	}

	file, err := r.f.ConfirmUpload(ctx.UserContext(), userID, params.ID, body.UploadID, completedParts)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(file)
}

// @Summary     Abort upload
// @Description Abort unfinished upload and cleanup temporary objects
// @ID          abort-upload
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id      path     string true  "File ID"
// @Param       request body     request.AbortUpload false "Abort upload payload"
// @Success     204
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id}/abort [post]
func (r *V1) abortUpload(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params struct {
		ID string `params:"id"`
	}
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - abortUpload")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid path param")
	}

	var body request.AbortUpload
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&body); err != nil {
			r.l.Error(err, "restapi - v1 - abortUpload")
			return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
		}
	}

	if err := r.f.AbortUpload(ctx.UserContext(), userID, params.ID, strings.TrimSpace(body.UploadID)); err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Get file download URL
// @Description Get a presigned URL for downloading an existing file
// @ID          get-file-download
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id   path     string true "File ID"
// @Success     200  {object} usecase.DownloadFileResult
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id}/download [get]
func (r *V1) downloadFile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params struct {
		ID string `params:"id"`
	}
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - downloadFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid path param")
	}

	result, err := r.f.GetDownload(ctx.UserContext(), userID, params.ID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(result)
}

// @Summary     Download file content
// @Description Redirect to presigned URL and start file download
// @ID          download-file-content
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id   path     string true "File ID"
// @Success     307
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id}/content [get]
func (r *V1) downloadFileContent(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params struct {
		ID string `params:"id"`
	}
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - downloadFileContent")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid path param")
	}

	result, err := r.f.GetDownload(ctx.UserContext(), userID, params.ID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Redirect(result.URL, http.StatusTemporaryRedirect)
}

// @Summary     Delete file
// @Description Delete file and all its parts from storage
// @ID          delete-file
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id   path     string true "File ID"
// @Success     200  {object} entity.File
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/files/{id} [delete]
func (r *V1) deleteFile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params struct {
		ID string `params:"id"`
	}
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - deleteFile")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid path param")
	}

	file, err := r.f.Delete(ctx.UserContext(), userID, params.ID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(file)
}
