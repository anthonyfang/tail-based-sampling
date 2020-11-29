package common

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
        go fetchData(body.Port)
    }

    return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", body.Port))
}

// SetParameterGetHandler is use for handling the SetParameterHandler endpoint
func SetParameterGetHandler(c *fiber.Ctx) error {
	port := c.Params("port")
    os.Setenv("UPLOAD_SERVER_PORT", port)

    serverPort := GetEnvDefault("SERVER_PORT", "8002")

    if serverPort != "8002" {
        go fetchData(port)
    }

    return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", port))
}

// fetchData is use for fetching the data file from datasource server
func fetchData(port string){
    var url string
    switch currentServerPort := GetEnvDefault("SERVER_PORT", ""); currentServerPort {
    case "8000":
        url = fmt.Sprintf("http://localhost:%v/trace1-small.data", port)
    case "8001":
        url = fmt.Sprintf("http://localhost:%v/trace2-small.data", port)
    default:
        url = ""
    }
    
    if url != "" {
        resp, err := http.Get(url)
        if err != nil {
            log.Fatalln(err)
        }

        defer resp.Body.Close()
        scanner := bufio.NewScanner(resp.Body)

        i := 0
        for scanner.Scan() {
            recordString := scanner.Text()
            record := strings.Split(recordString, "|")
            traceID := record[0]

            // validate error record
            hasError := false
            if len(record) > 8 {
                hasError = isErrorRecord(record[8])
            }

            // add the line to cacheQueue
            data := &RecordTemplate{hasError, i, false, []string{}}
            if len(CacheQueue[traceID].records) > 0 {
                data = &RecordTemplate{hasError, i, false, CacheQueue[traceID].records}
            }
            data.UpdateRecord(recordString)

            fmt.Println(scanner.Text())

            i++
        }
    }
}

func isErrorRecord(tags string) bool {
    result := (!strings.Contains(tags, "http.status_code=200") || strings.Contains(tags, ""))
    return result
}
