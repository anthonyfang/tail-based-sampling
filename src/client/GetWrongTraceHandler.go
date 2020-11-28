package client

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
)

// GetWrongTraceHandler is use for handling the SetWrongTraceId endpoint
func GetWrongTraceHandler(c *fiber.Ctx) error {
	return c.SendString(fmt.Sprintf("GetWrongTraceHandler: %v", c.Params("id")))
}
