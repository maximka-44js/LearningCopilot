package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maximka-44js/lc/internal/app/controller/restapi/v1/response"
)

func ErrorResponse(ctx *fiber.Ctx, code int, msg string) error {
	return ctx.Status(code).JSON(response.Error{Error: msg})
}
