package v1

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/maximka-44js/lc/internal/app/controller/restapi/v1"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/controller/restapi/v1/request"
	_ "github.com/maximka-44js/lc/internal/storage/entity"
)

// @Summary     Create library
// @Description Create a new library for the current user
// @ID          create-library
// @Tags        libraries
// @Accept      json
// @Produce     json
// @Param       request body     request.CreateLibrary true "Library data"
// @Success     201     {object} entity.Library
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries [post]
func (r *V1) createLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	ownerUsername, _ := ctx.Locals("username").(string)
	if displayName, ok := ctx.Locals("displayName").(string); ok && displayName != "" {
		ownerUsername = displayName
	}

	var body request.CreateLibrary

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - createLibrary")

		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - createLibrary")

		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	r.l.Debug("restapi - v1 - createLibrary - userID: %s", userID)
	r.l.Debug("restapi - v1 - createLibrary - body: %v", body)

	library, err := r.lb.Create(ctx.UserContext(), userID, ownerUsername, body.Title, body.Description, body.Public, body.Accesses)

	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
	}

	return ctx.Status(http.StatusCreated).JSON(library)
}

// @Summary     Get library by ID
// @Description Get a specific library by ID for the current user
// @ID          get-library-by-id
// @Tags        libraries
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Library ID"
// @Success     200  {object} entity.Library
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/{id} [get]
func (r *V1) getLibraryByID(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetLibraryByID

	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getLibraryByID")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	r.l.Debug("restapi - v1 - getLibraryByID - userID: %s", userID)
	r.l.Debug("restapi - v1 - getLibraryByID - params: %v", params)

	library, err := r.lb.GetByID(ctx.UserContext(), userID, params.LibID)

	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(library)
}

// @Summary     Update library
// @Description Update a library by ID for the current user
// @ID          update-library
// @Tags        libraries
// @Accept      json
// @Produce     json
// @Param       id      path     string true "Library ID"
// @Param       request body     request.UpdateLibrary true "Library data"
// @Success     200     {object} entity.Library
// @Failure     400     {object} d.AppError
// @Failure     401     {object} d.AppError
// @Failure     403     {object} d.AppError
// @Failure     404     {object} d.AppError
// @Failure     500     {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/{id} [put]
func (r *V1) updateLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.UpdateLibraryParams
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - updateLibrary")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	var body request.UpdateLibrary
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updateLibrary")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updateLibrary")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	r.l.Debug("restapi - v1 - updateLibrary - userID: %s", userID)
	r.l.Debug("restapi - v1 - updateLibrary - params: %v", params)
	r.l.Debug("restapi - v1 - updateLibrary - body: %v", body)

	library, err := r.lb.Update(
		ctx.UserContext(),
		userID,
		params.LibID,
		body.Title,
		body.Description,
		body.Public,
		body.Accesses,
	)

	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(library)
}

// @Summary     Compress library
// @Description Trigger ZIP compression task for a library
// @ID          compress-library
// @Tags        libraries
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Library ID"
// @Success     202  {object} map[string]string
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/{id}/compress [post]
func (r *V1) compressLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetLibraryByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - compressLibrary")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		r.l.Error(err, "restapi - v1 - compressLibrary")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	if err := r.lb.RequestCompress(ctx.UserContext(), userID, params.LibID); err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusAccepted).JSON(map[string]string{"status": "queued"})
}

// @Summary     Delete library
// @Description Delete a library by ID for the current user
// @ID          delete-library
// @Tags        libraries
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Library ID"
// @Success     200  {object} entity.Library
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     403  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/{id} [delete]
func (r *V1) deleteLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	lib, err := r.lb.Delete(ctx.UserContext(), userID, ctx.Params("id"))
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(lib)
}

// @Summary     List libraries
// @Description List all libraries for the current user
// @ID          list-libraries
// @Tags        libraries
// @Accept      json
// @Produce     json
// @Success     200  {array} entity.Library
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries [get]
func (r *V1) listLibraries(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	r.l.Debug("restapi - v1 - listLibraries - userID: %s", userID)
	libraries, err := r.lb.List(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(libraries)
}

// @Summary     List public libraries
// @Description List public libraries available in the global catalog
// @ID          list-public-libraries
// @Tags        libraries
// @Produce     json
// @Param       search query string false "Search in title, description, or owner username"
// @Param       sort query string false "Sort order: newest or popular"
// @Param       limit query int false "Number of libraries to return (default 20, maximum 50)"
// @Param       offset query int false "Number of libraries to skip"
// @Success     200  {object} entity.PublicLibrariesPage
// @Failure     401  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/public [get]
func (r *V1) listPublicLibraries(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	search := ctx.Query("search")
	limit := 20
	if value := ctx.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid limit")
		}
		limit = parsed
	}
	if limit > 50 {
		limit = 50
	}

	offset := 0
	if value := ctx.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid offset")
		}
		offset = parsed
	}

	sort := ctx.Query("sort", "newest")
	if sort != "newest" && sort != "popular" {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid sort")
	}
	libraries, err := r.lb.ListPublic(ctx.UserContext(), userID, search, sort, limit, offset)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(libraries)
}

// @Summary Star public library
// @Tags libraries
// @Param id path string true "Library ID"
// @Success 204
// @Security BearerAuth
// @Router /storage/libraries/{id}/star [post]
func (r *V1) starPublicLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err := r.lb.StarPublic(ctx.UserContext(), userID, ctx.Params("id")); err != nil {
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary Unstar public library
// @Tags libraries
// @Param id path string true "Library ID"
// @Success 204
// @Security BearerAuth
// @Router /storage/libraries/{id}/star [delete]
func (r *V1) unstarPublicLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err := r.lb.UnstarPublic(ctx.UserContext(), userID, ctx.Params("id")); err != nil {
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Add public library
// @Description Add a public library to the current user's library list
// @ID          add-public-library
// @Tags        libraries
// @Produce     json
// @Param       id   path     string true "Library ID"
// @Success     200  {object} entity.Library
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/{id}/add [post]
func (r *V1) addPublicLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetLibraryByID
	if err := ctx.ParamsParser(&params); err != nil {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	library, err := r.lb.AddPublic(ctx.UserContext(), userID, params.LibID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(library)
}

// @Summary     Unpin public library
// @Description Remove a public library from the current user's library list
// @ID          remove-public-library
// @Tags        libraries
// @Produce     json
// @Param       id   path     string true "Library ID"
// @Success     204
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/{id}/subscription [delete]
func (r *V1) removePublicLibrary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var params request.GetLibraryByID
	if err := ctx.ParamsParser(&params); err != nil {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.lb.RemovePublic(ctx.UserContext(), userID, params.LibID); err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Get library root contents
// @Description Get top-level folders and files for a library to render the tree root
// @ID          get-library-root-contents
// @Tags        libraries
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Library ID"
// @Success     200  {object} TreeContents
// @Failure     400  {object} d.AppError
// @Failure     401  {object} d.AppError
// @Failure     404  {object} d.AppError
// @Failure     500  {object} d.AppError
// @Security    BearerAuth
// @Router      /storage/libraries/{id}/tree [get]
func (r *V1) getLibraryContents(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return v1.ErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var params request.GetLibraryByID
	if err := ctx.ParamsParser(&params); err != nil {
		r.l.Error(err, "restapi - v1 - getLibraryContents")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}
	if err := r.v.Struct(params); err != nil {
		r.l.Error(err, "restapi - v1 - getLibraryContents")
		return v1.ErrorResponse(ctx, http.StatusBadRequest, "invalid request parameters")
	}

	contents, err := r.fd.GetLibraryContents(ctx.UserContext(), userID, params.LibID)
	if err != nil {
		r.l.Error(err)
		if appErr, ok := err.(*d.AppError); ok {
			return v1.ErrorResponse(ctx, appErr.Kind.HTTPCode(), appErr.Message)
		}
		return v1.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(contents)
}
