package common

import (
	// "fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"tail-based-sampling/src/cache"
)

// RecordTemplate is a template for record down each line of trace record info
type RecordTemplate struct {
    HasError      bool;
    BatchNo       int;
    Records       []string
}

// CacheServer is using cache the records
var CacheServer = cache.NewCache(0, 2 * time.Second)

// BadTraceIDList is recording down the bad trace IDs
var BadTraceIDList = []string{}
var badListLocker = sync.Mutex{}

var wg sync.WaitGroup

// BadTraceList is a list record down the bad trace
var BadTraceList = make(map[string]*RecordTemplate)

var cacheChan = make(chan string, 1)

// PostTraceChan is a channel for sending/receiving the signal 
var PostTraceChan = make(chan string)

// UpdateRecord is using for updating the record in cache
func(data *RecordTemplate) UpdateRecord(record string) {
    data.Records = append(data.Records, record)
}

// SortRecords is sorting the records field
func(data *RecordTemplate) SortRecords(){
    // bubbleSort
    len := len(data.Records)

    for i := 0; i < len - 1; i++ {
        for j := 0; j < len - 1 - i; j++ {
            arrJ, _:= strconv.Atoi(strings.Split(data.Records[j], "|")[1])
            arrJ1, _ := strconv.Atoi(strings.Split(data.Records[j+1], "|")[1])

            if(arrJ > arrJ1) {
                temp := data.Records[j+1];
                data.Records[j+1] = data.Records[j];
                data.Records[j] = temp;
            }
        }
    }
}

// GenCheckSumToQueue is using for generate the ckSum
func(data *RecordTemplate) GenCheckSumToQueue(traceID string, result map[string]string) {
    checkSumString := strings.Join(data.Records, "\n") + "\n"
    result[traceID] = MD5(checkSumString)
}

// GetTraceInfo is getting the traceInfo
func GetTraceInfo(traceID string) *RecordTemplate {
    traceCacheInfo := CacheServer.Get(traceID)
    traceInfo := &RecordTemplate{}
    if traceCacheInfo != nil {
        traceInfo = traceCacheInfo.(*RecordTemplate)
    }
    return traceInfo
}

// SetTraceInfo is setting the traceInfo
func SetTraceInfo(traceID string, data *RecordTemplate) {
    CacheServer.Set(traceID, data, 2 * time.Second)
}
