package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	// "io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// "sync"
	// "math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

var isRunning = false
var counter uint64 = 0
var batchNo int32 = 1

var timeWindow = 0.05
var timeRolling = 0.05
var currentTime int64 = 0
var timeWindowStart int64 = 0
var timeWindowClose int64 = 0

var TimeChan = make(chan int64)

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
	serverPort := GetEnvDefault("SERVER_PORT", "8002")

	if serverPort != "8002" && !isRunning {
		isRunning = true
		url := getURL(port)
		go fetchData(url)

		// tigger time window
		go func() {
			for val := range TimeChan {
				rollOver := false
				if timeWindowStart == 0 {
					timeWindowStart = val
					timeWindowClose = timeWindowStart + int64(timeWindow*1000000)
				} else {
					if val > timeWindowClose {
						rollOver = true
						timeWindowStart = val + int64(timeRolling*1000000)
						timeWindowClose = timeWindowStart + int64(timeWindow*1000000)
					}
				}

				if rollOver {
					fmt.Println("triggered send IDs, current count: ", counter)
					postTraceIDs(int(batchNo))
					// fmt.Println("batchNo: ", batchNo)
					atomic.AddInt32(&batchNo, 1)
				}
				// time.Sleep(2 * time.Second)
			}
		}()
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

		var httpWg sync.WaitGroup
		for scanner.Scan() {
			recordString := scanner.Text()
			httpWg.Add(1)
			go pushToCache(recordString, int(batchNo), &httpWg)
			// time.Sleep(2 * time.Second)
		}
		httpWg.Wait()
		close(TimeChan)

		fmt.Println("xxxxxxxxxxxxxxxxxx: pushed ", counter)
		go postFinishSignal()
	}
	fmt.Println("################# : fetchingData END", time.Now())
	fmt.Println("################# : fetchingData Total Elapsed Time: ", time.Since(startTime))
}

func getURL(port string) string {
	var url string
	switch currentServerPort := GetEnvDefault("SERVER_PORT", ""); currentServerPort {
	case "8000":
		url = fmt.Sprintf("http://localhost:%v/trace1.data", port)
	case "8001":
		url = fmt.Sprintf("http://localhost:%v/trace2.data", port)
	default:
		url = ""
	}

	return url
}

func pushToCache(recordString string, batchNo int, httpWg *sync.WaitGroup) {
	defer httpWg.Done()

	record := strings.Split(recordString, "|")
	traceID := record[0]

	// validate error record
	hasError := false
	if len(record) > 8 {
		currentTime, _ = strconv.ParseInt(record[1], 10, 64)
		TimeChan <- currentTime
		hasError = isErrorRecord(record[8])
		// add the line to cache server
		traceCacheInfo, _ := CacheQueue.Load(traceID)
		data := &RecordTemplate{hasError, batchNo, []string{}}
		if traceCacheInfo != nil {
			traceInfo := traceCacheInfo.(*RecordTemplate)
			newHasError := traceInfo.HasError
			if !newHasError {
				newHasError = hasError
			}
			data = &RecordTemplate{newHasError, batchNo, traceInfo.Records}
		}

		data.UpdateRecord(recordString)
		// if traceID == "c074d0a90cd607b" {
		// 	fmt.Println("Set Trace: ", data)
		// }
		SetTraceInfo(traceID, data)

		atomic.AddUint64(&counter, 1)

		if hasError {
			go func() { BadTraceIDList = append(BadTraceIDList, traceID) }()
		}
	}
}

func isErrorRecord(tags string) bool {
	result := (strings.Contains(tags, "http.status_code=") && !strings.Contains(tags, "http.status_code=200")) || strings.Contains(tags, "error=1")
	return result
}

func postTraceIDs(batchNo int) {
	badListLocker.Lock()
	badTraceIDList := BadTraceIDList
	BadTraceIDList = []string{}

	CacheQueue.Store(strconv.Itoa(batchNo), badTraceIDList)

	mjson, err := json.Marshal(RecordTemplate{
		BatchNo: batchNo,
		Records: badTraceIDList,
	})
	if err != nil {
		log.Fatal(err)
	}

	badTraceIDList = []string{}
	badListLocker.Unlock()

	go func(mjson []byte) {
		res, err := http.Post("http://localhost:8002/setWrongTraceId", "application/json", bytes.NewBuffer(mjson))
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
	}(mjson)
}

func postFinishSignal() {
	type finishSignalTemplate struct {
		Port string
	}

	mjson, err := json.Marshal(finishSignalTemplate{
		Port: GetEnvDefault("SERVER_PORT", "8002"),
	})
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post("http://localhost:8002/finish", "application/json", bytes.NewBuffer(mjson))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}
