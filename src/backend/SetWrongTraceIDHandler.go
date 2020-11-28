package backend

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetWrongTraceIDHandler(c *fiber.Ctx) error {
    type Request struct {
        ID string `json:"id"`
    }

    var body Request
    err := c.BodyParser(&body)
    if err != nil {
        fmt.Println(err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Cannot parse Body JSON",
        })
    }

	return c.SendString(fmt.Sprintf("setWrongTraceId: %v", body.ID))
}
