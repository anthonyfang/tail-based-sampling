package client

// import (
// 	"fmt"

// 	"github.com/gofiber/fiber/v2"

// 	"tail-based-sampling/src/common"
// )

// // GetWrongTraceHandler is use for handling the SetWrongTraceId endpoint
// func GetWrongTraceHandler(c *fiber.Ctx) error {
// 	traceID := c.Params("traceID")
// 	// batchNo, _ := strconv.Atoi(c.Params("batchNo"))
// 	data := &common.RecordTemplate{HasError: true, BatchNo: 0, Records: []string{}}
// 	traceInfo := common.GetTraceInfo(traceID)

// 	if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
// 		fmt.Println(traceInfo)
// 	}
// 	if traceInfo != nil {
// 		traceInfo.SyncRecords.Range(func(k, v interface{}) bool {
// 			traceInfo.Records = append(traceInfo.Records, k.(string))
// 			return true
// 		})
// 		data = traceInfo
// 	}

// 	defer common.CacheQueue.Delete(traceID)

// 	return c.JSON(data)
// }
