package backend

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetParameterGetHandler(c *fiber.Ctx) error {

	port := c.Query("port")

	os.Setenv("UPLOAD_SERVER_PORT", port)

	go processing()

	return c.SendString("OK!")
}

func SetParameterPostHandler(c *fiber.Ctx) error {

	return c.SendString("Please use GET request instead!")
}
