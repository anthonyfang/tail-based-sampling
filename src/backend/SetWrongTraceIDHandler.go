package backend

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "sync"
    "tail-based-sampling/src/common"

    "github.com/gofiber/fiber/v2"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetWrongTraceIDHandler(c *fiber.Ctx) error {
    type Request struct {
        BatchNo int         `json:"batchNo"`;
        Records []string    `json:"records"`;
    }

    var body Request
    err := c.BodyParser(&body)
    if err != nil {
        fmt.Println(err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Cannot parse Body JSON",
        })
    }
    go processing(body.BatchNo, body.Records)

    return c.SendString("OK!")
}

// TODO
// var clientHosts = []string{"http://localhost:8000", "http://localhost:8001"}
var clientHosts = []string{"http://localhost:8000"}

func processing(batchNo int, records []string) {
    // Request all the clients to get all the bad trace info
    for _, traceID := range records {
        wg.Add(1)

        go func(traceID string, wg1 *sync.WaitGroup) {
            var wgHostData sync.WaitGroup
            for _, url := range clientHosts {
                wgHostData.Add(1)
                go func(url string, batchNo int, traceID string, wgHostData1 *sync.WaitGroup){
                    defer wgHostData1.Done()
                    getWrongTraceInfo(url + "/getWrongTrace", batchNo, traceID)
                }(url, batchNo, traceID, &wgHostData)
                // Ensure all the clients return data back
                wgHostData.Wait()

                traceInfoCache := common.GetTraceInfo(traceID)
                if traceInfoCache != nil {
                    // sort
                    traceInfoCache.SortRecords()

                    // generate checkSum to result queue
                    resultQueueLocker.Lock()
                    traceInfoCache.GenCheckSumToQueue(traceID, resultQueue)
                    defer resultQueueLocker.Unlock()
                }
            }
            defer wg1.Done()
        }(traceID, &wg)
    }
}

func getWrongTraceInfo(URL string, batchNo int, traceID string) {
    res, err := http.Get(URL + "/"+ strconv.Itoa(batchNo) + "/" + traceID)
    if err != nil {
        log.Fatal(err)
    }
    defer res.Body.Close()

    var traceInfo common.RecordTemplate
    err = json.NewDecoder(res.Body).Decode(&traceInfo)
    if err != nil {
        log.Fatalln(err)
    }

    // Push into the cache server
    if len(traceInfo.Records) > 0 {
        traceInfoCache := common.GetTraceInfo(traceID)
        if traceInfoCache != nil {
            traceInfo.Records = append(traceInfoCache.Records, traceInfo.Records...)
        }
        common.SetTraceInfo(traceID, traceInfoCache)
    }
}
