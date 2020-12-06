package backend

import (
	"tail-based-sampling/src/common"

	"github.com/gofiber/fiber/v2"
)

// FinishHandler is using for trigger calculation
func FinishHandler(c *fiber.Ctx) error {
	msg := "Finished!"

	// TODO
	// if len(clientPorts) == 2 {

	common.FinishedChan <- "complete"

	return c.SendString(msg)
}
