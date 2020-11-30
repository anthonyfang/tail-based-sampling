package client

import (
    "strconv"
    "github.com/gofiber/fiber/v2"

    "tail-based-sampling/src/common"
)

// Fake data will be removed
func generateFakeData() {
    common.CacheQueueBk["c074d0a90cd607b"] = &common.RecordTemplate{
        HasError:      true,
        BatchNo:       1,
        Records:       []string{
            "c074d0a90cd607b|1589285991244547|39950c516d123a42|3bd7a959290a5f69|830|PromotionCenter|db.UserDao.getUser(..)|192.168.164.40|db.instance=db&component=java-jdbc&db.type=h2&span.kind=client&__sql_id=1x7lx2l&peer.address=localhost:8082",
            "c074d0a90cd607b|1589285991244551|312e7ae8c1f2dd98|3bd7a959290a5f69|836|OrderCenter|db.AlertDao.countByTitleAndUserIdAndFilterStr(..)|192.168.164.42|db.instance=db&component=java-jdbc&db.type=h2&span.kind=client&__sql_id=1af9zkk&peer.address=localhost:8082",
        },
    }
}

// GetWrongTraceHandler is use for handling the SetWrongTraceId endpoint
func GetWrongTraceHandler(c *fiber.Ctx) error {
    traceID := c.Params("traceID")
    batchNo, _ := strconv.Atoi(c.Params("batchNo"))
    data := &common.RecordTemplate{HasError: true, BatchNo: batchNo, Records:[]string{}}
    // generateFakeData()
    common.CQLocker.Lock()
    if common.CacheQueueBk[traceID] != nil && common.CacheQueueBk[traceID].BatchNo == batchNo {
        data = common.CacheQueueBk[traceID]
    }
    common.CQLocker.Unlock()


	return c.JSON(data)
}
