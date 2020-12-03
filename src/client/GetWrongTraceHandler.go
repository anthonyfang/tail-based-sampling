package client

import (
	// "fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"tail-based-sampling/src/common"
)

// GetWrongTraceHandler is use for handling the SetWrongTraceId endpoint
func GetWrongTraceHandler(c *fiber.Ctx) error {
    traceID := c.Params("traceID")
    batchNo, _ := strconv.Atoi(c.Params("batchNo"))
    data := &common.RecordTemplate{HasError: true, BatchNo: batchNo, Records:[]string{}}
    traceInfo := common.GetTraceInfo(traceID)
    if traceInfo != nil {
        data = traceInfo
    }
    common.CacheQueue.Delete(traceID)

	return c.JSON(data)
}
