package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"tail-based-sampling/src/common"
	"time"
)

const (
	MAX_CONCURRENCY = 60
)

var batchReceivedCount = 0

var ch = make(chan struct{})

var clientHosts = []string{"http://localhost:8000", "http://localhost:8001"}

//var clientHosts = []string{"http://localhost:8000"}

var tmpChan = make(chan struct{}, MAX_CONCURRENCY)

var finishSignals = 0

var batchGate = 4

var tmpBatchQueue = make(map[int]bool)

var tmpCheckSumQueue = make(map[string]*common.RecordTemplate)

func processing() {

	defer close(common.FinishedChan)

	go func() {
		for traceID := range common.GenCheckSumToQueueChan {
			resultQueueLocker.Lock()
			tmpCheckSumQueue[traceID].GenCheckSumToQueue(traceID, resultQueue)
			resultQueueLocker.Unlock()

			time.Sleep(500)
		}
	}()

	for {
		select {
		case batchNo := <-common.BatchReceivedCountChan:

			var num, ok = BackendBatchQueue.Load(batchNo)
			if !ok {
				num = 0
			}
			num = num.(int) + 1
			BackendBatchQueue.Store(batchNo, num)

			tmpBatchQueue[batchNo] = false

			processAllCachedBatches(false)

		case <-common.FinishedChan:
			finishSignals++
			if finishSignals == len(clientHosts) {
				processAllCachedBatches(true)
				time.Sleep(200)
				sendCheckSum(resultQueue)
				func() {
					fmt.Println("============= Result ================")
					for key, value := range resultQueue {
						fmt.Println("XXXXXXXXXXXXX ", key, ": --------- ", value)
					}
					fmt.Println("============= END ================", time.Now())
				}()
				return
			}

		}

		time.Sleep(100)
	}
}

func processAllCachedBatches(noControl bool) {
	fmt.Println("processing")
	var sendNum uint64 = 0

	for tmpBatchNo, _ := range tmpBatchQueue {

		var num, _ = BackendBatchQueue.Load(tmpBatchNo)
		if common.IsDebug {
			fmt.Println("tmpBatchQueue[", tmpBatchNo, "] - BackendBatchQueue[", tmpBatchNo, "]: ", num)
		}
		// fmt.Println("num.(int) > len(clientHosts):", num.(int) > len(clientHosts), ", len(tmpBatchQueue)%batchGate: ", len(tmpBatchQueue)%batchGate)
		if noControl || (num.(int) >= len(clientHosts) && len(tmpBatchQueue) >= batchGate) {
			BackendTraceIDQueue.Range(func(k, v interface{}) bool {
				batchNo := v.(int)

				if tmpBatchNo == batchNo {
					traceID := k.(string)

					var wgHostData sync.WaitGroup
					for i, url := range clientHosts {
						wgHostData.Add(1)
						atomic.AddUint64(&sendNum, 1)
						go getWrongTraceInfo(url+"/getWrongTrace", traceID, i, &wgHostData)
					}
					// Ensure all the clients return data back
					wgHostData.Wait()

					// get the records from the above request calls and join them together
					newInfoArr := []string{}
					traceInfoCache := common.GetTraceInfo(traceID)
					for i, _ := range clientHosts {
						traceInfo := common.GetTraceInfo(traceID + "-" + strconv.Itoa(i))
						newInfoArr = append(newInfoArr, traceInfo.Records...)
					}
					traceInfoCache.Records = append(traceInfoCache.Records, newInfoArr...)

					if traceInfoCache != nil && len(traceInfoCache.Records) > 0 {
						// sort
						traceInfoCache.SortRecords()
						if common.IsDebug && traceID == common.DebugTraceID {
							fmt.Println(traceInfoCache)
						}
						// generate checkSum to result queue
						tmpCheckSumQueue[traceID] = traceInfoCache
						common.GenCheckSumToQueueChan <- traceID
					}

					for i, _ := range clientHosts {
						common.CacheQueue.Delete(traceID + "-" + strconv.Itoa(i))
					}
					BackendTraceIDQueue.Delete(k)
					BackendBatchQueue.Delete(v.(int))
					tmpBatchQueue[v.(int)] = true
				}
				return true
			})
			if common.IsDebug {
				fmt.Println("BatchNo", tmpBatchNo, ", Sent numbers: ", sendNum)
			}
		}
	}
	for k, v := range tmpBatchQueue {
		if v {
			delete(tmpBatchQueue, k)
		}
	}
	// Request all the clients to get all the bad trace info
}

func getWrongTraceInfo(URL string, traceID string, id int, wgHostData *sync.WaitGroup) {
	tmpChan <- struct{}{}
	url := URL + "/" + traceID
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
	if common.IsDebug && traceID == common.DebugTraceID {
		fmt.Println("-----traceInfo start----")
		fmt.Println(traceInfo.Records)
		fmt.Println("-----traceInfo end----")
	}
	if len(traceInfo.Records) > 0 {
		common.SetTraceInfo(traceID+"-"+strconv.Itoa(id), &traceInfo)
	}
	wgHostData.Done()
	<-tmpChan
}
