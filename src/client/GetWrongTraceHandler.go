package client

import (
	"github.com/gofiber/fiber/v2"

	"tail-based-sampling/src/common"
)

// GetWrongTraceHandler is use for handling the SetWrongTraceId endpoint
func GetWrongTraceHandler(c *fiber.Ctx) error {
	traceID := c.Params("traceID")
	// batchNo, _ := strconv.Atoi(c.Params("batchNo"))
	data := &common.RecordTemplate{HasError: true, BatchNo: 0, Records: []string{}}
	traceInfo := common.GetTraceInfo(traceID)

	// if traceID == "c074d0a90cd607b" {
	// 	fmt.Println(traceInfo)
	// }
	if traceInfo != nil {
		data = traceInfo
	}
	common.CacheQueue.Delete(traceID)

	return c.JSON(data)
}
