package common

import (
	"bufio"
	"bytes"
	"encoding/json"
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

        i := 0
        batchNo := 0
        for scanner.Scan() {
            recordString := scanner.Text()
            pushToCacheQueue(recordString, batchNo)

            // tigger the cleanup worker every 20000 record
            if i % 20000 == 0 {
                BackupCacheQueue()
                // Post traceIDs to backend
                go postTraceIDs(batchNo)

                batchNo++
            }
            i++
        }
        
        fmt.Println("xxxxxxxxxxxxxxxxxx: ", i)
        for key, record := range CacheQueue {
            if record.HasError == true {
                fmt.Println("Key:", key, "--", record.HasError)
            }
        }
    }
    fmt.Println("################# : fetchingData END", time.Now())
}

func getURL(port string) string {
    var url string
    switch currentServerPort := GetEnvDefault("SERVER_PORT", ""); currentServerPort {
    case "8000":
        url = fmt.Sprintf("http://localhost:%v/trace2-small.data", port)
    case "8001":
        url = fmt.Sprintf("http://localhost:%v/trace2-small.data", port)
    default:
        url = ""
    }

    return url
}

func pushToCacheQueue(recordString string, batchNo int) {
    CQLocker.Lock()
    record := strings.Split(recordString, "|")
    traceID := record[0]

    // validate error record
    hasError := false
    if len(record) > 8 {
        hasError = isErrorRecord(record[8])
    }

    // add the line to cacheQueue
    data := &RecordTemplate{hasError, batchNo, []string{}}
    if CacheQueue[traceID] != nil {
        newHasError := CacheQueue[traceID].HasError
        if !newHasError {
            newHasError = hasError
        }
        data = &RecordTemplate{newHasError, batchNo, CacheQueue[traceID].Records}
    }
    data.UpdateRecord(recordString)
    CacheQueue[traceID] = data
    CQLocker.Unlock()
}

func isErrorRecord(tags string) bool {
    result := (strings.Contains(tags, "http.status_code=") && !strings.Contains(tags, "http.status_code=200")) || strings.Contains(tags, "error=1")
    return result
}

func postTraceIDs(batchNo int) {
    mjson, err := json.Marshal(RecordTemplate {
        BatchNo: batchNo,
        Records: BadTraceList[string(batchNo)].Records,
    })
    if err != nil {
        log.Fatal(err)
    }

    res, err := http.Post("http://localhost:8002/setWrongTraceId", "application/json", bytes.NewBuffer(mjson))
    if err != nil {
        log.Fatal(err)
    }
    defer res.Body.Close()
}
