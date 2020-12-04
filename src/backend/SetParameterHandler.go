package backend

import (
	"github.com/gofiber/fiber/v2"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetParameterGetHandler(c *fiber.Ctx) error {

	go processing()

	return c.SendString("OK!")
}

func SetParameterPostHandler(c *fiber.Ctx) error {

	SetParameterGetHandler(c)
	return c.SendString("OK!")
}
