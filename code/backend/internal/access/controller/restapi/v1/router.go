package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/maximka-44js/lc/internal/access/usecase"
	"github.com/maximka-44js/lc/internal/app/controller/restapi/middleware"
	"github.com/maximka-44js/lc/pkg/jwt"
	"github.com/maximka-44js/lc/pkg/logger"
)

func NewRoutes(accessGroup fiber.Router, apiKeys usecase.APIKeys, jwtManager *jwt.Manager, l logger.Interface) {
	r := &V1{
		apiKeys: apiKeys,
		l:       l,
		v:       validator.New(validator.WithRequiredStructEnabled()),
	}

	protected := accessGroup.Group("", middleware.Auth(jwtManager, l, apiKeys), middleware.RequireJWT())
	protected.Post("", r.createAPIKey)
	protected.Get("", r.listAPIKeys)
	protected.Delete("/:id", r.revokeAPIKey)
	protected.Post("/:id/rotate", r.rotateAPIKey)
}
