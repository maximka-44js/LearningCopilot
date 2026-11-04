package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/maximka-44js/lc/internal/access/entity"
)

func RequireJWT() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authType, _ := ctx.Locals(ctxAuthType).(string)
		if authType == "" {
			return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "unauthorized"})
		}
		if authType != authTypeJWT {
			return ctx.Status(http.StatusForbidden).JSON(errorResponse{Error: "forbidden"})
		}
		return ctx.Next()
	}
}

func RequireDomainAccess(domain entity.ScopeDomain) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authType, _ := ctx.Locals(ctxAuthType).(string)
		if authType != authTypeAPIKey {
			return ctx.Next()
		}

		scopes, ok := ctx.Locals(ctxAPIKeyScope).([]string)
		if !ok {
			return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "unauthorized"})
		}

		access := entity.ScopeAccessForMethod(ctx.Method())
		if !entity.HasScope(scopes, domain, access) {
			return ctx.Status(http.StatusForbidden).JSON(errorResponse{Error: "insufficient API key scope"})
		}

		return ctx.Next()
	}
}
