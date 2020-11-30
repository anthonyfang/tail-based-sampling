package backend

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
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
    wg.Wait()
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

        go func(traceID string) {
            defer wg.Done()
            bufferChan <- traceID
            for _, url := range clientHosts {
                go getWrongTraceInfo(url + "/getWrongTrace", batchNo, traceID)
            }
            <-bufferChan

        }(traceID)
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

    // Push into the result working queue
    if len(traceInfo.Records) > 0 {
        go pushReusltWorkingQueue(traceInfo, traceID)
    }
}

func pushReusltWorkingQueue(traceInfo common.RecordTemplate, traceID string) {
    common.CQLocker.Lock()
    if resultWorkingQueue[traceID] != nil {
        traceInfo.Records = append(resultWorkingQueue[traceID].Records, traceInfo.Records...)
    }
    resultWorkingQueue[traceID] = &traceInfo
    common.CQLocker.Unlock()
}
