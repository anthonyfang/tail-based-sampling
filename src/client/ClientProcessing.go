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

	"github.com/golang/protobuf/proto"
)

const (
	TIME_WINDOW = 0.05
	ROLLING     = 0.05
	BATCH_GATE  = 3
)

var batchNo int32 = 1
var currentTime int64 = 0
var timeWindowStart int64 = 0
var timeWindowEnd int64 = 0

var TimeChan = make(chan int64)

func processing() {
	fmt.Println("Running Readline process...")
	for newline := range common.NewLineChan {
		pushToCache(newline.Line, newline.BatchNo)
	}
	common.FinishedChan <- "readline"
}

func windowing() {
	fmt.Println("Running Windowing process...")
	for val := range TimeChan { // tigger time window
		rollOver := false
		if timeWindowStart == 0 {
			timeWindowStart = val
			timeWindowEnd = timeWindowStart + int64(TIME_WINDOW*1000000)
		} else {
			if val > timeWindowEnd {
				rollOver = true
				timeWindowStart = val + int64(ROLLING*1000000)
				timeWindowEnd = timeWindowStart + int64(TIME_WINDOW*1000000)
			}
		}

		if rollOver {
			common.Wg.Add(1)
			postTraceIDs(int(batchNo))
			atomic.AddInt32(&batchNo, 1)

			go func() {
				if batchNo%10 == 0 {
					common.CacheQueue.Range(func(k, v interface{}) bool {
						if len(k.(string)) > 8 {
							traceInfo := v.(*common.RecordTemplate)
							if traceInfo.LifeTime > BATCH_GATE {
								common.CacheQueue.Delete(k)
							} else {
								traceInfo.LifeTime++
							}
						} else {
							num := k.(string)
							numInt, _ := strconv.Atoi(num)
							if numInt < int(batchNo)-BATCH_GATE*5 {
								common.CacheQueue.Delete(k)
							}
						}
						return true
					})
				}
			}()
			// fmt.Println("batchNo: ", batchNo)
		}
	}
	common.Wg.Wait()

	common.Wg.Add(1)
	postTraceIDs(int(batchNo))
	common.Wg.Wait()
	common.FinishedChan <- "timeWindow"
}

func postTraceIDs(batchNo int) {
	var badListLocker = sync.Mutex{}

	if common.IS_DEBUG {
		fmt.Println("triggered send IDs, current count: ", counter)
	}
	go func(batchNo int) {
		badListLocker.Lock()
		badTraceIDList := common.BadTraceIDList
		common.BadTraceIDList = []string{}

		common.CacheQueue.Store(strconv.Itoa(batchNo), badTraceIDList)

		if batchNo > 1 {
			// var payload = new(common.Payload)
			previousBatch := batchNo - 1
			badTraceIDs, _ := common.CacheQueue.Load(strconv.Itoa(previousBatch))

			// payload.SetWrongTraceIDGen(strconv.Itoa(previousBatch), badTraceIDs.([]string))

			payload := &trace.PayloadMessage{
				Action:  "SetWrongTraceID",
				ID:      strconv.Itoa(previousBatch),
				Records: badTraceIDs.([]string),
			}
			// payload.ReturnWrongTraceGen(traceID, payload)

			// msg, _ := json.Marshal(payload)

			msg, err := proto.Marshal(payload)
			if err != nil {
				log.Fatal("marshaling error: ", err)
			}

			// msg, _ := json.Marshal(payload)
			_, err = ws1.Write(msg)
			if err != nil {
				log.Fatal(err)

			}
		}
		badTraceIDList = []string{}
		badListLocker.Unlock()
		common.Wg.Done()
	}(batchNo)
}

func pushToCache(recordString string, batchNo int) {
	record := strings.Split(recordString, "|")
	traceID := record[0]

	// validate error record
	hasError := false
	if len(record) > 8 {
		currentTime, _ = strconv.ParseInt(record[1], 10, 64)
		hasError = isErrorRecord(record[8])
		// add the line to cache server
		traceCacheInfo, _ := common.CacheQueue.Load(traceID)
		data := &common.RecordTemplate{hasError, batchNo, 0, []string{}, sync.Map{}}

		if traceCacheInfo != nil {
			traceInfo := traceCacheInfo.(*common.RecordTemplate)
			newHasError := traceInfo.HasError
			if !newHasError {
				newHasError = hasError
			}
			traceInfo.HasError = newHasError

			traceInfo.SyncRecords.Store(recordString, batchNo)

			if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
				fmt.Println("Add Trace: ", recordString)
			}
		} else {
			data.BatchNo = batchNo
			data.SyncRecords.Store(recordString, batchNo)
			common.CacheQueue.Store(traceID, data)

			if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
				fmt.Println("New Trace: ", recordString)
			}
		}

		atomic.AddUint64(&counter, 1)

		if hasError {
			common.BadTraceIDList = append(common.BadTraceIDList, traceID)
		}

		TimeChan <- currentTime
	}
}

func isErrorRecord(tags string) bool {
	result := (strings.Contains(tags, "http.status_code=") && !strings.Contains(tags, "http.status_code=200")) || strings.Contains(tags, "error=1")
	return result
}
