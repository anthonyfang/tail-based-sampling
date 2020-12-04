package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"tail-based-sampling/src/common"
	"time"
)

const (
	MAX_CONCURRENCY = 100
)

var batchReceivedCount = 0

var ch = make(chan struct{})

// TODO
var clientHosts = []string{"http://localhost:8000", "http://localhost:8001"}

//var clientHosts = []string{"http://localhost:8000"}

var tmpChan = make(chan struct{}, MAX_CONCURRENCY)

func processing() {

	for {

		if batchReceivedCount >= 2 {
			fmt.Println("processing")
			var sendNum uint64
			atomic.StoreUint64(&sendNum, 0)

			// Request all the clients to get all the bad trace info
			BackendTraceIDQueue.Range(func(k, v interface{}) bool {
				traceID := k.(string)
				fmt.Println("xxx ", traceID)
				//go func(traceID string) {
				var wgHostData sync.WaitGroup
				for _, url := range clientHosts {
					wgHostData.Add(1)
					go func(url string, traceID string, wgHostData *sync.WaitGroup) {
						defer wgHostData.Done()
						atomic.AddUint64(&sendNum, 1)
						getWrongTraceInfo(url+"/getWrongTrace", traceID)
					}(url, traceID, &wgHostData)
				}
				// Ensure all the clients return data back
				wgHostData.Wait()

				traceInfoCache := common.GetTraceInfo(traceID)

				if traceInfoCache != nil && len(traceInfoCache.Records) > 0 {

					// sort
					traceInfoCache.SortRecords()
					if traceID == "c074d0a90cd607b" {
						fmt.Println(traceInfoCache)
					}
					// generate checkSum to result queue
					resultQueueLocker.Lock()
					traceInfoCache.GenCheckSumToQueue(traceID, resultQueue)
					resultQueueLocker.Unlock()
				}
				BackendTraceIDQueue.Delete(k)
				// }(traceID)
				return true
			})

			lock := &sync.Mutex{}
			lock.Lock()
			batchReceivedCount = 0
			fmt.Println("Sent numbers: ", sendNum)
			// fmt.Println(counter)
			lock.Unlock()
		}

		if finishedSignal {
			go sendCheckSum(resultQueue)
			go func() {
				fmt.Println("============= Result ================")
				// wg.Wait()
				for key, value := range resultQueue {
					fmt.Println("XXXXXXXXXXXXX ", key, ": --------- ", value)
				}
				fmt.Println("============= END ================", time.Now())
			}()
			break
		}

		time.Sleep(500)
	}
}

func getWrongTraceInfo(URL string, traceID string) {
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
	// fmt.Println(res.Body)
	// if traceID == "c074d0a90cd607b" {
	// 	fmt.Println(*res)
	// }
	// if traceID == "c074d0a90cd607b" {
	// 	fmt.Println(traceInfo.Records)
	// }
	// Push into the cache server
	if len(traceInfo.Records) > 0 {
		traceInfoCache := common.GetTraceInfo(traceID)

		if traceInfoCache != nil && len(traceInfoCache.Records) > 0 {
			traceInfo.Records = append(traceInfoCache.Records, traceInfo.Records...)
		}
		common.SetTraceInfo(traceID, &traceInfo)
	}
	<-tmpChan
}
