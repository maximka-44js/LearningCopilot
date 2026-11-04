package middleware

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maximka-44js/lc/pkg/logger"
)

func buildRequestMessage(ctx *fiber.Ctx) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(maskSensitiveURLQuery(ctx.OriginalURL()))
	result.WriteString(" - ")
	result.WriteString(strconv.Itoa(ctx.Response().StatusCode()))
	result.WriteString(" ")
	result.WriteString(strconv.Itoa(len(ctx.Response().Body())))

	return result.String()
}

func Logger(l logger.Interface) func(c *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		err := ctx.Next()

		l.Info("%s", buildRequestMessage(ctx))

		return err
	}
}

func maskSensitiveURLQuery(raw string) string {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return raw
	}

	q := u.Query()
	masked := false
	for key := range q {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if normalized == "access_token" || normalized == "api_key" {
			q.Set(key, "***")
			masked = true
		}
	}
	if !masked {
		return raw
	}

	u.RawQuery = q.Encode()
	return u.String()
}
