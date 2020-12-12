package backend

import (
	"fmt"
	"strconv"
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
	// processing()
}

func processing() {
	fmt.Println("Running Backend process...")
	defer close(common.FinishedChan)

	go agregateForTraceID()

	// go wsWriteLoop()

	go gRPCWriteLoop()

	for {
		select {
		case batchNo := <-common.BatchReceivedCountChan:
			// var x = new(uint32)
			// var ptValue, _ = BackendBatchQueue.LoadOrStore(batchNo, x)
			// var ptNum = ptValue.(*uint32)
			// atomic.AddUint32(ptNum, 1)
			var val = 1
			var num, ok = BackendBatchQueue.Load(batchNo)
			if ok {
				val = num.(int) + 1
			}
			fmt.Println("BatchReceivedCountChan", batchNo)
			BackendBatchQueue.Store(batchNo, val)
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
		var x = new(uint32)
		var ptValue, _ = BackendReceivedTraceInfo.LoadOrStore(traceID, x)
		var ptNum = ptValue.(*uint32)
		atomic.AddUint32(ptNum, 1)
		// fmt.Println(*ptNum)

		// var x, ok = BackendReceivedTraceInfo.Load(traceID)
		// var num = 1
		// if ok {
		// 	num = x.(int) + 1
		// }
		// BackendReceivedTraceInfo.Store(traceID, *ptNum)

		if int(*ptNum) >= len(common.ClientHosts) {

			newInfoArr := []string{}
			traceInfoCache := common.GetTraceInfo(traceID)

			for i, _ := range common.ClientHosts {
				traceInfo := common.GetTraceInfo(traceID + "-" + strconv.Itoa(i))
				newInfoArr = append(newInfoArr, traceInfo.Records...)
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
				// common.GenCheckSumToQueueChan <- traceID
			}
			for i, _ := range common.ClientHosts {
				common.CacheQueue.Delete(traceID + "-" + strconv.Itoa(i))
			}

			BackendReceivedTraceInfo.Delete(traceID)
			common.CacheQueue.Delete(traceID)

		}
		wg.Done()
	}
}

func processAllCachedBatches(lastBatch bool) {
	if common.IS_DEBUG {
		fmt.Println("processing")
	}
	var sendNum uint64 = 0

	for tmpBatchNo, _ := range tmpBatchQueue {
		// var t, _ = BackendBatchQueue.Load(tmpBatchNo)
		// var num = t.(*uint32)
		var t, _ = BackendBatchQueue.Load(tmpBatchNo)
		var num = t.(int)

		if common.IS_DEBUG {
			fmt.Println("tmpBatchQueue[", tmpBatchNo, "] - BackendBatchQueue[", tmpBatchNo, "]: ", num)
		}

		if lastBatch || (num >= len(common.ClientHosts) && len(tmpBatchQueue) >= BATCH_GATE) {
			BackendTraceIDQueue.Range(func(k, v interface{}) bool {
				batchNo := v.(int)

				if tmpBatchNo == batchNo {
					traceID := k.(string)

					atomic.AddUint64(&sendNum, uint64(len(common.ClientHosts)))
					common.ServerSendWSChan <- traceID

					BackendTraceIDQueue.Delete(k)
					BackendBatchQueue.Delete(v.(int))
					tmpBatchQueue[v.(int)] = true
				}
				return true
			})
			if common.IS_DEBUG {
				fmt.Println("BatchNo", tmpBatchNo, ", Sent numbers: ", sendNum)
			}
		}
	}
	for k, v := range tmpBatchQueue {
		if v {
			delete(tmpBatchQueue, k)
		}
	}
	if lastBatch {
		close(common.GenCheckSumToQueueChan)
	}
}
