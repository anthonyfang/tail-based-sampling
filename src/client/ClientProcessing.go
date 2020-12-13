package client

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"tail-based-sampling/src/common"
	"tail-based-sampling/src/trace"
	"time"

	cmap "github.com/orcaman/concurrent-map"
)

const (
	TIME_WINDOW = 0.05
	ROLLING     = 0.05
	BATCH_GATE  = 3
)

var batchNo int32 = 0
var currentTime int64 = 0
var timeWindowStart int64 = 0
var timeWindowEnd int64 = 0

var batchMap = cmap.New()

func StartClientProcess() {
	port := <-common.ReadyChan
	time.Sleep(time.Millisecond * 10)
	url := getURL("8080")
	fmt.Println(port)

	common.TraceInfoStore.New()

	go runTraceChat()

	go processing()

	fetchData(url)
}

func processing() {
	fmt.Println("Running Readline process...")
	for newline := range common.NewLineChan {
		// atomic.AddUint64(&counter, 1)
		// if counter%100000 == 0 {
		// 	fmt.Println(newline.Line[:48])
		// }
		pushToCache(newline.Line, newline.BatchNo)

		if counter > 0 && counter%5000 == 0 {
			record := strings.Split(newline.Line, "|")

			if len(record) > 8 {

				currentTime, _ = strconv.ParseInt(record[1], 10, 64)
				rollOver := false
				if timeWindowStart == 0 {
					timeWindowStart = currentTime
					timeWindowEnd = timeWindowStart + int64(TIME_WINDOW*1000000)
				} else {
					if currentTime > timeWindowEnd {
						rollOver = true
						timeWindowStart = currentTime + int64(ROLLING*1000000)
						timeWindowEnd = timeWindowStart + int64(TIME_WINDOW*1000000)
					}
				}

				if rollOver {
					common.Wg.Add(1)
					atomic.AddInt32(&batchNo, 1)

					if common.IS_DEBUG {
						fmt.Printf("triggered send IDs, batch %d - current count %d \n", batchNo, counter)
					}

					var badListLocker = sync.Mutex{}
					badListLocker.Lock()
					badTraceIDList := common.BadTraceIDList
					common.BadTraceIDList = []string{}
					badListLocker.Unlock()

					go postTraceIDs(int(batchNo), badTraceIDList, false)
				}
			}
		}

		if counter > 0 && counter%1000000 == 0 {
			common.TraceInfoStore.HouskeepTill(20000)

			for e := range batchMap.IterBuffered() {
				key, _ := strconv.Atoi(e.Key)
				if key < int(batchNo)-BATCH_GATE*5 {
					batchMap.Remove(e.Key)
				}
			}
		}
	}
	common.Wg.Wait()

	common.Wg.Add(1)
	var badListLocker = sync.Mutex{}
	badListLocker.Lock()
	badTraceIDList := common.BadTraceIDList
	common.BadTraceIDList = []string{}
	badListLocker.Unlock()

	atomic.AddInt32(&batchNo, 1)
	postTraceIDs(int(batchNo), badTraceIDList, true)

	common.Wg.Wait()
	common.FinishedChan <- "timeWindow"
}

func pushToCache(recordString string, batchNo int) {
	record := strings.Split(recordString, "|")

	if len(record) > 8 {
		traceID := record[0]

		// validate error record
		hasError := false

		hasError = isErrorRecord(record[8])
		// add the line to cache server

		common.TraceInfoStore.Add(traceID, recordString)

		atomic.AddUint64(&counter, 1)

		if hasError {
			common.BadTraceIDList = append(common.BadTraceIDList, traceID)
		}

	}
}

func postTraceIDs(batchNo int, badTraceIDList []string, last bool) {

	batchMap.Set(strconv.Itoa(batchNo), badTraceIDList)

	if batchNo > 1 {
		previousBatch := batchNo - 1
		badTraceIDs, _ := batchMap.Get(strconv.Itoa(previousBatch))

		if last {
			badTraceIDs = append(badTraceIDs.([]string), badTraceIDList...)
		}

		var added = len(badTraceIDs.([]string))
		atomic.AddInt32(&badTraceCounter, int32(added))

		if badTraceIDs != nil && len(badTraceIDs.([]string)) > 0 {
			payload := &trace.PayloadMessage{
				Action:  "SetWrongTraceID",
				ID:      strconv.Itoa(previousBatch),
				Records: badTraceIDs.([]string),
			}
			if err := (*gRPCstream).Send(payload); err != nil {
				log.Fatal(err)

			}
		}

	}
	common.Wg.Done()
}

func isErrorRecord(tags string) bool {
	result := (strings.Contains(tags, "http.status_code=") && !strings.Contains(tags, "http.status_code=200")) || strings.Contains(tags, "error=1")
	return result
}
