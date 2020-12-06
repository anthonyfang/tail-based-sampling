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
		BatchNo int      `json:"batchNo"`
		Records []string `json:"records"`
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
var clientHosts = []string{"http://localhost:8000", "http://localhost:8001"}

//var clientHosts = []string{"http://localhost:8000"}

func processing(batchNo int, records []string) {
	finishWg.Add(1)
	fmt.Println("start to  process batch: " + strconv.Itoa(batchNo))

	// Request all the clients to get all the bad trace info
	for _, traceID := range records {
		if contains(common.BadTraceIDList, traceID) {
			continue
		}
		common.BadTraceIDList = append(common.BadTraceIDList, traceID)
		go func(traceID string, _wg *sync.WaitGroup) {
			var wgHostData sync.WaitGroup
			for _, url := range clientHosts {
				wgHostData.Add(1)

				go func(url string, batchNo int, traceID string, _wgHostData *sync.WaitGroup) {
					defer _wgHostData.Done()
					getWrongTraceInfo(url+"/getWrongTrace", batchNo, traceID)
				}(url, batchNo, traceID, &wgHostData)
			}
			// Ensure all the clients return data back
			wgHostData.Wait()

			traceInfoCache := common.GetTraceInfo(traceID)
			if traceInfoCache != nil && len(traceInfoCache.Records) > 0 {
				// sort
				traceInfoCache.SortRecords()

				// generate checkSum to result queue
				resultQueueLocker.Lock()

				traceInfoCache.GenCheckSumToQueue(traceID, resultQueue)
				defer resultQueueLocker.Unlock()
			}
		}(traceID, &wg)
	}
	fmt.Println("finish process batch: " + strconv.Itoa(batchNo))
	finishWg.Done()
}

func getWrongTraceInfo(URL string, batchNo int, traceID string) {

	url := URL + "/" + strconv.Itoa(batchNo) + "/" + traceID

	res, err := http.Get(url)
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

		if traceInfoCache != nil && len(traceInfoCache.Records) > 0 {
			traceInfo.Records = append(traceInfoCache.Records, traceInfo.Records...)
		}

		common.SetTraceInfo(traceID, &traceInfo)
	}
}
