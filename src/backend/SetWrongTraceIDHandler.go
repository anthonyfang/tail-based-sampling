package backend

// import (
// 	"fmt"
// 	"tail-based-sampling/src/common"

// 	"github.com/gofiber/fiber/v2"
// )

// // SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
// func SetWrongTraceIDHandler(c *fiber.Ctx) error {
// 	type Request struct {
// 		Server  string   `json:"server"`
// 		BatchNo int      `json:"batchNo"`
// 		Records []string `json:"records"`
// 	}

// 	var body Request
// 	err := c.BodyParser(&body)
// 	if err != nil {
// 		fmt.Println(err)
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "Cannot parse Body JSON",
// 		})
// 	}

// 	for _, v := range body.Records {
// 		// fmt.Println(v)
// 		BackendTraceIDQueue.Store(v, body.BatchNo)
// 	}

// 	common.BatchReceivedCountChan <- body.BatchNo

// 	return c.SendString("OK!")
// }
