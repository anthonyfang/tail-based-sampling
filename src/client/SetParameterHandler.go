package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"tail-based-sampling/src/common"
	"time"

	"github.com/gofiber/fiber/v2"
)

var isRunning = false
var counter uint64 = 0
var downloadChan = make(chan bool)

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
		go startClientProcess(port)
		// // go fetchData(url)
	}

	return c.SendString(fmt.Sprintf("OK! Upload server port is: %v", port))
}

// DownloadFile is download file with Range header
func MultiPartsDownload(url string) {
	startTime := time.Now()
	fmt.Println("################# : fetchingData", startTime)
	res, _ := http.Head(url)
	maps := res.Header
	length, _ := strconv.Atoi(maps["Content-Length"][0])
	limit := 2
	lenSub := length / limit
	diff := length % limit

	var wg sync.WaitGroup

	for i := 0; i < limit; i++ {
		wg.Add(1)
		min := lenSub * i       // Min range
		max := lenSub * (i + 1) // Max range

		if i == limit-1 {
			max += diff // Add the remaining bytes in the last request
		}

		go func(min int, max int, i int) {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", url, nil)
			rangeHeader := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1) // Add the data for the Range header of the form "bytes=0-100"
			req.Header.Add("Range", rangeHeader)
			resp, _ := client.Do(req)
			defer func() { resp.Body.Close() }()

			scanner := bufio.NewScanner(resp.Body)
			buf := make([]byte, 64*1024)
			scanner.Buffer(buf, bufio.MaxScanTokenSize)

			for scanner.Scan() {
				recordString := scanner.Text()
				common.NewLineChan <- common.NewLine{Line: recordString, BatchNo: 0}
				time.Sleep(300)
			}
			defer func() {
				wg.Done()
			}()
		}(min, max, i)
	}
	wg.Wait()

	close(common.NewLineChan)
	for msg := range common.FinishedChan {
		// if msg == "readline" {
		// 	TimeChan <- timeWindowEnd + 1
		// 	common.Wg.Wait()
		// 	close(TimeChan)
		// }

		if msg == "timeWindow" {
			fmt.Println("xxxxxxxxxxxxxxxxxx: pushed ", counter)
			go postFinishSignal()

			fmt.Println("################# : fetchingData END", time.Now())
			fmt.Println("################# : fetchingData Total Elapsed Time: ", time.Since(startTime))
			// return
		}
	}
}

func startClientProcess(port string) {
	url := getURL(port)

	// go windowing()

	go processing()

	if url != "" {
		fmt.Println("Start download...")
		MultiPartsDownload(url)
	}
	// go readData()
}

func readData() {
	fmt.Println("Start to read file")
	startTime := time.Now()
	fmt.Println("################# : fetchingData", startTime)

	file, err := os.Open("/tmp/datafile")

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, bufio.MaxScanTokenSize)
	//scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		recordString := scanner.Text()
		common.NewLineChan <- common.NewLine{Line: recordString, BatchNo: 0}
	}

	close(common.NewLineChan)
	for msg := range common.FinishedChan {
		if msg == "readline" {
			TimeChan <- timeWindowEnd + 1
			common.Wg.Wait()
			close(TimeChan)
		}

		if msg == "timeWindow" {
			fmt.Println("xxxxxxxxxxxxxxxxxx: pushed ", counter)
			go postFinishSignal()

			fmt.Println("################# : fetchingData END", time.Now())
			fmt.Println("################# : fetchingData Total Elapsed Time: ", time.Since(startTime))
			// return
		}
	}
}

// fetchData is use for fetching the data file from datasource server
// func fetchData(url string) {
// 	startTime := time.Now()
// 	fmt.Println("################# : fetchingData", startTime)
// 	if url != "" {
// 		resp, err := http.Get(url)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}

// 		defer resp.Body.Close()
// 		scanner := bufio.NewScanner(resp.Body)
// 		buf := make([]byte, 64*1024)
// 		scanner.Buffer(buf, bufio.MaxScanTokenSize)

// 		for scanner.Scan() {
// 			recordString := scanner.Text()
// 			common.NewLineChan <- common.NewLine{Line: recordString, BatchNo: int(batchNo)}
// 		}

// 		close(common.NewLineChan)
// 		for msg := range common.FinishedChan {
// 			if msg == "readline" {
// 				TimeChan <- timeWindowEnd + 1
// 				common.Wg.Wait()
// 				close(TimeChan)
// 			}

// 			if msg == "timeWindow" {
// 				fmt.Println("xxxxxxxxxxxxxxxxxx: pushed ", counter)
// 				go postFinishSignal()

// 				fmt.Println("################# : fetchingData END", time.Now())
// 				fmt.Println("################# : fetchingData Total Elapsed Time: ", time.Since(startTime))
// 				// return
// 			}
// 		}
// 	}
// }

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func getURL(port string) string {
	var url string
	// port = "8080"
	switch currentServerPort := common.GetEnvDefault("SERVER_PORT", ""); currentServerPort {
	case "8000":
		url = fmt.Sprintf("http://localhost:%v/trace1.data", port)
	case "8001":
		url = fmt.Sprintf("http://localhost:%v/trace2.data", port)
	default:
		url = ""
	}

	return url
}

func postFinishSignal() {
	var payload = new(common.Payload)
	payload.SendFinishGen(common.GetEnvDefault("SERVER_PORT", "8002"))

	msg, _ := json.Marshal(payload)

	_, err := ws1.Write(msg)
	if err != nil {
		log.Fatal(err)
	}
}
