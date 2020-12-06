package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"tail-based-sampling/src/common"
	"time"
)

var batchNo int32 = 1

var timeWindow = 0.05
var timeRolling = 0.05
var currentTime int64 = 0
var timeWindowStart int64 = 0
var timeWindowEnd int64 = 0

var TimeChan = make(chan int64)

func processing() {

	for {
		select {
		case newline, ok := <-common.NewLineChan:

			if !ok {
				common.FinishedChan <- "readline"
				return
			}
			lineParts := strings.Split(newline, ",")

			batchNo, _ := strconv.Atoi(lineParts[1])

			pushToCache(lineParts[0], batchNo)

		}
		time.Sleep(100)
	}
}

func windowing() {
	for {
		val, ok := <-TimeChan // tigger time window
		if !ok {
			common.FinishedChan <- "timeWindow"
			return
		}
		rollOver := false
		if timeWindowStart == 0 {
			timeWindowStart = val
			timeWindowEnd = timeWindowStart + int64(timeWindow*1000000)
		} else {
			if val > timeWindowEnd {
				rollOver = true
				timeWindowStart = val + int64(timeRolling*1000000)
				timeWindowEnd = timeWindowStart + int64(timeWindow*1000000)
			}
		}

		if rollOver {
			postTraceIDs(int(batchNo))
			// fmt.Println("batchNo: ", batchNo)
			atomic.AddInt32(&batchNo, 1)
		}
		time.Sleep(100)
	}
}

func postTraceIDs(batchNo int) {
	var badListLocker = sync.Mutex{}

	fmt.Println("triggered send IDs, current count: ", counter)
	badListLocker.Lock()
	badTraceIDList := common.BadTraceIDList
	common.BadTraceIDList = []string{}

	common.CacheQueue.Store(strconv.Itoa(batchNo), badTraceIDList)

	mjson, err := json.Marshal(common.RecordTemplate{
		Server:  common.GetEnvDefault("SERVER_PORT", "3000"),
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
		data := &common.RecordTemplate{"", hasError, batchNo, []string{}, sync.Map{}}

		if traceCacheInfo != nil {
			traceInfo := traceCacheInfo.(*common.RecordTemplate)
			newHasError := traceInfo.HasError
			if !newHasError {
				newHasError = hasError
			}
			traceInfo.HasError = newHasError
			// data = &common.RecordTemplate{traceInfo.Server, newHasError, batchNo, traceInfo.Records}
			traceInfo.SyncRecords.Store(recordString, batchNo)

			if common.IsDebug && traceID == common.DebugTraceID {
				fmt.Println("Add Trace: ", recordString)
			}
		} else {
			data.BatchNo = batchNo
			data.SyncRecords.Store(recordString, batchNo)
			common.CacheQueue.Store(traceID, data)

			if common.IsDebug && traceID == common.DebugTraceID {
				fmt.Println("New Trace: ", recordString)
			}
		}

		atomic.AddUint64(&counter, 1)

		if hasError {
			go func() { common.BadTraceIDList = append(common.BadTraceIDList, traceID) }()
		}

		TimeChan <- currentTime
	}
}

func isErrorRecord(tags string) bool {
	result := (strings.Contains(tags, "http.status_code=") && !strings.Contains(tags, "http.status_code=200")) || strings.Contains(tags, "error=1")
	return result
}
