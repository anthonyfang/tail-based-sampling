package backend

import (
	"github.com/gofiber/fiber/v2"
)

// FinishHandler is using for trigger calculation
func FinishHandler(c *fiber.Ctx) error {
	msg := "Finished!"

	// TODO
	// if len(clientPorts) == 2 {

	finishedSignal = true

	return c.SendString(msg)
}
