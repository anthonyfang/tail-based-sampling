package backend

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"tail-based-sampling/src/common"
	"time"
)

const (
	BATCH_GATE = 6
)

var batchReceivedCount = 0
var finishSignals = 0

var tmpBatchQueue = make(map[int]bool)
var tmpCheckSumQueue = make(map[string]*common.RecordTemplate)

func StartBackendProcess() {
	port := <-common.ReadyChan

	fmt.Println(port)
	processing()
}

func processing() {
	fmt.Println("Running Backend process...")
	defer close(common.FinishedChan)

	go agregateForTraceID()

	go gRPCWriteLoop()

	for {
		select {
		case batchNo := <-common.BatchReceivedCountChan:
			// var x = new(uint32)
			// var ptValue, _ = BackendBatchQueue.LoadOrStore(batchNo, x)
			// var ptNum = ptValue.(*uint32)
			// atomic.AddUint32(ptNum, 1)
			var val = 1
			var num, ok = BackendBatchQueue.Get(strconv.Itoa(batchNo))
			if ok {
				val = num.(int) + 1
			}
			fmt.Printf("BatchReceivedCountChan %d, count %d \n", batchNo, val)
			BackendBatchQueue.Set(strconv.Itoa(batchNo), val)
			tmpBatchQueue[batchNo] = false

			processAllCachedBatches(false)

		case <-common.FinishedChan:
			finishSignals++

			if finishSignals == len(common.ClientHosts) {
				processAllCachedBatches(true)
				wg.Wait()

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

func agregateForTraceID() {
	for traceID := range common.ReceivedTraceInfoChan {

		var val = 1
		var num, ok = BackendReceivedTraceInfoCount.Get(traceID)
		if ok {
			val = num.(int) + 1
		}
		BackendReceivedTraceInfoCount.Set(traceID, val)

		if val >= len(common.ClientHosts) {

			newInfoArr := []string{}
			traceInfoCache := common.GetTraceInfo(traceID)

			for _, url := range common.ClientHosts {
				urlParts := strings.Split(url, ":")
				traceInfo := common.GetTraceInfo(traceID + "-" + urlParts[2])
				newInfoArr = append(newInfoArr, traceInfo.Records...)
				if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
					fmt.Println("--------", traceID+"-"+urlParts[2])
					fmt.Println(traceInfo.Records)
				}
				// common.CacheQueue.Remove(traceID + "-" + urlParts[2])
			}

			traceInfoCache.Records = newInfoArr
			if traceInfoCache != nil && len(traceInfoCache.Records) > 0 {
				// sort
				traceInfoCache.SortRecords()
				if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
					fmt.Println(traceInfoCache)
				}
				// generate checkSum to result queue
				tmpCheckSumQueue[traceID] = traceInfoCache
				tmpCheckSumQueue[traceID].GenCheckSumToQueue(traceID, resultQueue)
			}

			BackendReceivedTraceInfo.Remove(traceID)
			common.CacheQueue.Remove(traceID)

		}
		wg.Done()
	}
}

func processAllCachedBatches(lastBatch bool) {
	if common.IS_DEBUG {
		fmt.Println("processing")
	}
	var sendNum uint64 = 0
	var bufferToSend = []string{}

	for tmpBatchNo, _ := range tmpBatchQueue {
		var t, _ = BackendBatchQueue.Get(strconv.Itoa(tmpBatchNo))
		var num = t.(int)

		if common.IS_DEBUG {
			fmt.Printf("tmpBatchQueue[%d] - BackendBatchQueue[%d]: %d \n", tmpBatchNo, tmpBatchNo, num)
		}

		if lastBatch || (num >= len(common.ClientHosts) && len(tmpBatchQueue) >= BATCH_GATE) {
			for e := range BackendTraceIDQueue.IterBuffered() {
				batchNo := e.Val.(int)

				if tmpBatchNo == batchNo {
					traceID := e.Key

					atomic.AddUint64(&sendNum, uint64(len(common.ClientHosts)))

					bufferToSend = append(bufferToSend, traceID)

					BackendTraceIDQueue.Remove(traceID)
					BackendBatchQueue.Remove(strconv.Itoa(batchNo))
					tmpBatchQueue[tmpBatchNo] = true
				}
			}
		}
	}

	for i := 0; i < len(bufferToSend); i++ {
		common.ServerSendWSChan <- bufferToSend[i]
		if i >= 100 && i%100 == 0 {
			time.Sleep(time.Millisecond * 300)
		}
	}

	for k, v := range tmpBatchQueue {
		if v {
			delete(tmpBatchQueue, k)
		}
	}
	fmt.Println("Sent numbers: ", sendNum)
}
