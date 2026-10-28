package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/maximka-44js/lc/internal/access/entity"
	"github.com/maximka-44js/lc/internal/app/controller/restapi/middleware"
	"github.com/maximka-44js/lc/internal/storage/repo"
	"github.com/maximka-44js/lc/internal/storage/usecase"
	"github.com/maximka-44js/lc/pkg/jwt"
	"github.com/maximka-44js/lc/pkg/logger"
)

// NewRoutes -.
func NewRoutes(storageGroup fiber.Router, lb usecase.Library, fd usecase.Folder, f usecase.File, os repo.ObjectStorage, jwtManager *jwt.Manager, apiKeyAuth middleware.APIKeyAuthenticator, l logger.Interface) {
	r := &V1{
		lb: lb,
		fd: fd,
		f:  f,
		os: os,
		l:  l,
		v:  validator.New(validator.WithRequiredStructEnabled()),
	}

	public := storageGroup.Group("")
	protected := storageGroup.Group(
		"",
		middleware.Auth(jwtManager, l, apiKeyAuth),
		middleware.RequireDomainAccess(entity.ScopeDomainStorage),
	)
	protected.Post("/libraries", r.createLibrary)
	protected.Get("/libraries/:id", r.getLibraryByID)
	protected.Get("/libraries/:id/tree", r.getLibraryContents)
	protected.Post("/libraries/:id/compress", r.compressLibrary)
	protected.Get("/libraries", r.listLibraries)
	protected.Put("/libraries/:id", r.updateLibrary)
	protected.Delete("/libraries/:id", r.deleteLibrary)

	protected.Post("/folders", r.createFolder)
	protected.Get("/folders/:id", r.getFolderByID)
	protected.Get("/folders/:id/tree", r.getFolderContents)
	protected.Post("/folders/:id/compress", r.compressFolder)
	protected.Post("/folders/:id/extract", r.extractFolder)
	protected.Put("/folders/:id", r.updateFolder)
	protected.Post("/folders/:id/copy", r.copyFolder)
	protected.Delete("/folders/:id", r.deleteFolder)

	// files
	protected.Post("/files/upload", r.uploadFile)
	protected.Post("/files/init", r.initUpload)
	protected.Put("/files/:id", r.updateFile)
	protected.Post("/files/:id/copy", r.copyFile)
	protected.Post("/files/:id/parts", r.presignUploadPart)
	protected.Post("/files/:id/complete", r.completeUpload)
	protected.Post("/files/:id/abort", r.abortUpload)
	protected.Get("/files/:id/download", r.downloadFile)
	protected.Get("/files/:id/content", r.downloadFileContent)
	protected.Delete("/files/:id", r.deleteFile)
	_ = r
	_ = public
	_ = protected
}
