package backend

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetWrongTraceIDHandler(c *fiber.Ctx) error {
	type Request struct {
		BatchNo int      `json:"batchNo"`
		Records []string `json:"records"`
	}

	var body Request
	err := c.BodyParser(&body)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot parse Body JSON",
		})
	}
	go addTraceID(body.BatchNo, body.Records)

	return c.SendString("OK!")
}

func addTraceID(batchNo int, records []string) {
	for _, v := range records {
		// fmt.Println(v)
		BackendTraceIDQueue.Store(v, false)
	}
	lock := &sync.Mutex{}
	lock.Lock()
	batchReceivedCount++
	// fmt.Println(batchReceivedCount)
	lock.Unlock()
}
