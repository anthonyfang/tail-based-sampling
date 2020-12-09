package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"tail-based-sampling/src/common"

	// "io"
	"log"
	"net/http"
	"os"
	"time"

	// "sync"
	// "math"

	"github.com/gofiber/fiber/v2"
)

var isRunning = false
var counter uint64 = 0

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

	return c.SendString(fmt.Sprintf("Please use Get Request! Upload server port is: %v", body.Port))
}

// SetParameterGetHandler is use for handling the SetParameterHandler endpoint
func SetParameterGetHandler(c *fiber.Ctx) error {
	port := c.Query("port")

	os.Setenv("UPLOAD_SERVER_PORT", port)
	serverPort := common.GetEnvDefault("SERVER_PORT", "8002")

	if serverPort != "8002" && !isRunning {
		isRunning = true
		url := getURL(port)

		go wsConnection()

		go windowing()

		go processing()

		go fetchData(url)
	}

	return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", port))
}

// fetchData is use for fetching the data file from datasource server
func fetchData(url string) {
	startTime := time.Now()
	fmt.Println("################# : fetchingData", startTime)
	if url != "" {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, bufio.MaxScanTokenSize)

		for scanner.Scan() {
			recordString := scanner.Text()
			common.NewLineChan <- fmt.Sprintf("%s|,,|%i", recordString, int(batchNo))
		}

		close(common.NewLineChan)
		for {
			msg := <-common.FinishedChan
			if msg == "readline" {
				TimeChan <- timeWindowEnd + 1
				close(TimeChan)
			}

			if msg == "timeWindow" {
				fmt.Println("xxxxxxxxxxxxxxxxxx: pushed ", counter)
				go postFinishSignal()

				fmt.Println("################# : fetchingData END", time.Now())
				fmt.Println("################# : fetchingData Total Elapsed Time: ", time.Since(startTime))

				return
			}
			time.Sleep(500)
		}
	}
}

func getURL(port string) string {
	var url string
	port = "8080"
	switch currentServerPort := common.GetEnvDefault("SERVER_PORT", ""); currentServerPort {
	case "8000":
		url = fmt.Sprintf("http://localhost:%v/trace1-4G.data", port)
	case "8001":
		url = fmt.Sprintf("http://localhost:%v/trace2-4G.data", port)
	default:
		url = ""
	}

	return url
}

func postFinishSignal() {
	// res, err := http.Post("http://localhost:8002/finish", "application/json", bytes.NewBuffer(mjson))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer res.Body.Close()
	var payload = new(common.Payload)
	payload.SendFinishGen(common.GetEnvDefault("SERVER_PORT", "8002"))

	msg, _ := json.Marshal(payload)

	_, err := ws.Write(msg)
	if err != nil {
		log.Fatal(err)
	}
}
