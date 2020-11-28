package common

import (
    "fmt"
    "os"
    "github.com/gofiber/fiber/v2"
)

// SetParameterHandler is use for handling the SetParameterHandler endpoint
func SetParameterHandler(c *fiber.Ctx) error {
	type Request struct {
        Port string `json:"port"`
    }
    var body Request

    err := c.BodyParser(&body)
    if err != nil {
        fmt.Println(err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Cannot parse Body JSON",
        })
    }
    os.Setenv("UPLOAD_SERVER_PORT", body.Port)
    return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", body.Port))
}
