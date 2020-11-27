package backend

import (
    "github.com/gofiber/fiber/v2"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetWrongTraceIDHandler(c *fiber.Ctx) error {
	return c.SendString("setWrongTraceId")
}
