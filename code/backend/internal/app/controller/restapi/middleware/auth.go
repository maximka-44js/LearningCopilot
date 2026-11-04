package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	accessUseCase "github.com/maximka-44js/lc/internal/access/usecase"
	"github.com/maximka-44js/lc/pkg/jwt"
	"github.com/maximka-44js/lc/pkg/logger"
)

const _bearerParts = 2

const (
	ctxUserID      = "userID"
	ctxAuthType    = "authType"
	ctxAPIKeyID    = "apiKeyID"
	ctxAPIKeyScope = "apiKeyScopes"

	authTypeJWT    = "jwt"
	authTypeAPIKey = "api_key"

	apiKeyTokenPrefix = "flk_"
)

type errorResponse struct {
	Error string `json:"error"`
}

type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, in accessUseCase.AuthenticateAPIKeyInput) (accessUseCase.AuthenticateAPIKeyOutput, error)
}

// Auth returns a JWT authentication middleware for Fiber.
func Auth(jwtManager *jwt.Manager, l logger.Interface, apiKeyAuth APIKeyAuthenticator) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		token, fromQuery, err := extractToken(ctx)
		if err != nil {
			return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: err.Error()})
		}

		if fromQuery {
			userID, parseErr := jwtManager.ParseToken(token)
			if parseErr != nil {
				if isWebSocketUpgradeRequest(ctx) {
					l.Error(parseErr, "restapi - auth - parse websocket query token")
				} else {
					l.Error(parseErr, "restapi - auth - parse query token")
				}
				return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "invalid or expired token"})
			}

			ctx.Locals(ctxUserID, userID)
			ctx.Locals(ctxAuthType, authTypeJWT)

			return ctx.Next()
		}

		if isAPIKeyToken(token) {
			if isWebSocketUpgradeRequest(ctx) {
				return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "API key auth is not supported for websocket"})
			}
			if apiKeyAuth == nil {
				return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "invalid or expired token"})
			}

			out, authErr := apiKeyAuth.Authenticate(ctx.UserContext(), accessUseCase.AuthenticateAPIKeyInput{PresentedKey: token})
			if authErr != nil {
				l.Error(authErr, "restapi - auth - authenticate api key")
				return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "invalid or expired token"})
			}

			ctx.Locals(ctxUserID, out.APIKey.UserID)
			ctx.Locals(ctxAuthType, authTypeAPIKey)
			ctx.Locals(ctxAPIKeyID, out.APIKey.ID)
			ctx.Locals(ctxAPIKeyScope, out.APIKey.Scopes)

			return ctx.Next()
		}

		userID, err := jwtManager.ParseToken(token)
		if err != nil {
			l.Error(err, "restapi - auth - parse token")
			return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "invalid or expired token"})
		}

		ctx.Locals(ctxUserID, userID)
		ctx.Locals(ctxAuthType, authTypeJWT)

		return ctx.Next()
	}
}

func extractToken(ctx *fiber.Ctx) (string, bool, error) {
	header := strings.TrimSpace(ctx.Get("Authorization"))
	if header != "" {
		parts := strings.SplitN(header, " ", _bearerParts)
		if len(parts) != _bearerParts || !strings.EqualFold(parts[0], "Bearer") {
			return "", false, fiber.NewError(http.StatusUnauthorized, "invalid authorization header format")
		}

		return parts[1], false, nil
	}

	if isWebSocketUpgradeRequest(ctx) {
		queryToken := strings.TrimSpace(ctx.Query("access_token"))
		if queryToken != "" {
			return queryToken, true, nil
		}
	}

	return "", false, fiber.NewError(http.StatusUnauthorized, "missing authorization header")
}

func isAPIKeyToken(token string) bool {
	trimmed := strings.TrimSpace(token)
	if !strings.HasPrefix(trimmed, apiKeyTokenPrefix) {
		return false
	}
	parts := strings.SplitN(trimmed, ".", 2)
	return len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != ""
}

func isWebSocketUpgradeRequest(ctx *fiber.Ctx) bool {
	if websocket.IsWebSocketUpgrade(ctx) {
		return true
	}

	upgrade := strings.TrimSpace(ctx.Get("Upgrade"))
	connection := strings.ToLower(strings.TrimSpace(ctx.Get("Connection")))

	return strings.EqualFold(upgrade, "websocket") && strings.Contains(connection, "upgrade")
}
