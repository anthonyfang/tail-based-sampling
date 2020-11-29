package common

import (
    "bufio"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
)

// SetParameterPostHandler is use for handling the SetParameterHandler endpoint
func SetParameterPostHandler(c *fiber.Ctx) error {
    type Request struct {
        Port string `json:"port"`
    }
    var body Request

    err := c.BodyParser(&body)
    if err != nil {
        fmt.Println(err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Cannot parse Body JSON",
        })
    }
    os.Setenv("UPLOAD_SERVER_PORT", body.Port)

    serverPort := GetEnvDefault("SERVER_PORT", "8002")

    if serverPort != "8002" {
        url := getURL(body.Port)
        go fetchData(url)
    }

    return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", body.Port))
}

// SetParameterGetHandler is use for handling the SetParameterHandler endpoint
func SetParameterGetHandler(c *fiber.Ctx) error {
    port := c.Query("port")
    
    os.Setenv("UPLOAD_SERVER_PORT", port)

    serverPort := GetEnvDefault("SERVER_PORT", "8002")

    if serverPort != "8002" {
        url := getURL(port)
        go fetchData(url)
    }

    return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", port))
}

// fetchData is use for fetching the data file from datasource server
func fetchData(url string){
    fmt.Println("################# : fetchingData", time.Now())
    if url != "" {
        resp, err := http.Get(url)
        if err != nil {
            log.Fatalln(err)
        }

        defer resp.Body.Close()
        scanner := bufio.NewScanner(resp.Body)
        buf := make([]byte, 64*1024)
        scanner.Buffer(buf, bufio.MaxScanTokenSize)
        // process(resp.Body)

        // scanner := bufio.NewScanner(f)

        i := 0
        lastCleanupIndex := i
        for scanner.Scan() {
            recordString := scanner.Text()
            pushToCacheQueue(recordString, i)
            // fmt.Println(scanner.Text())

            // tigger the cleanup worker every 20000 record
            if i - lastCleanupIndex > 20000 {
                go cleanUpWorker(i)
            }
            i++
        }
        
        fmt.Println("xxxxxxxxxxxxxxxxxx: ", i)
        for key, record := range CacheQueue {
            fmt.Println("Key:", key, "=>", "Element:",record)
        }
    }
    fmt.Println("################# : fetchingData END", time.Now())
}

func getURL(port string) string {
    var url string
    switch currentServerPort := GetEnvDefault("SERVER_PORT", ""); currentServerPort {
    case "8000":
        url = fmt.Sprintf("http://localhost:%v/trace1-small.data", port)
    case "8001":
        url = fmt.Sprintf("http://localhost:%v/trace2-small.data", port)
    default:
        url = ""
    }

    return url
}

func pushToCacheQueue(recordString string, currentLineNo int) {
    CQLocker.Lock()
    record := strings.Split(recordString, "|")
    traceID := record[0]

    // validate error record
    hasError := false
    if len(record) > 8 {
        hasError = isErrorRecord(record[8])
    }

    // add the line to cacheQueue
    data := &RecordTemplate{hasError, currentLineNo, false, []string{}}
    if CacheQueue[traceID] != nil {
        data = &RecordTemplate{hasError, currentLineNo, false, CacheQueue[traceID].records}
    }
    data.UpdateRecord(recordString)
    CacheQueue[traceID] = data
    CQLocker.Unlock()
}

func isErrorRecord(tags string) bool {
    result := (!strings.Contains(tags, "http.status_code=200") || strings.Contains(tags, "error=1"))
    return result
}

func cleanUpWorker(currentLineNo int) {
    CQLocker.Lock()
    for key, record := range CacheQueue {
        if (currentLineNo - record.startLineNO > 2000 && !record.hasError || record.hasReport) {
            // fmt.Println("---------------------------- Delete:", key)
            delete(CacheQueue, key)
        }
    }
    CQLocker.Unlock()
}
