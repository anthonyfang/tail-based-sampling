package common

import (
	"fmt"
    "bufio"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

// SetParameterHandler is use for handling the SetParameterHandler endpoint
func SetParameterHandler(c *fiber.Ctx) error {
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

    go fetchData(body.Port)

    return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", body.Port))
}

// fetchData is use for fetching the data file from datasource server
func fetchData(port string){
    var url string
    switch currentServerPort := GetEnvDefault("SERVER_PORT", ""); currentServerPort {
    case "8000":
        url = fmt.Sprintf("http://localhost:%v/trace1.data", port)
    case "8001":
        url = fmt.Sprintf("http://localhost:%v/trace2.data", port)
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

        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }
}
