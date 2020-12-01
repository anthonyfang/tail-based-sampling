package common

import (
    "strconv"
    "strings"
    "sync"
)

// RecordTemplate is a template for record down each line of trace record info
type RecordTemplate struct {
    HasError      bool;
    BatchNo       int;
    Records       []string
}

// CacheQueue is using cache all the records
var CacheQueue = make(map[string]*RecordTemplate)
var CacheQueueBk = make(map[string]*RecordTemplate)
var BadTraceList = make(map[string]*RecordTemplate)

var badListChan = make(chan string, 1)
var cacheQueueChan = make(chan string, 1)
var wg sync.WaitGroup


// BackupCacheQueue is moving the cache queue to bk queue
func BackupCacheQueue() {
    CQLocker.Lock()
    CacheQueueBk = CacheQueue
    CacheQueue = make(map[string]*RecordTemplate)
    BadTraceList = make(map[string]*RecordTemplate)

    for key, record := range CacheQueueBk {
        if record.HasError {
            data := &RecordTemplate{record.HasError, record.BatchNo, []string{}}
            if BadTraceList[string(record.BatchNo)] != nil {
                data = &RecordTemplate{record.HasError, record.BatchNo, BadTraceList[string(record.BatchNo)].Records}
            }
            data.UpdateRecord(key)
            BadTraceList[string(record.BatchNo)] = data
        }
    }
    CQLocker.Unlock()
}

// PostTraceChan is a channel for sending/receiving the signal 
var PostTraceChan = make(chan string)

// UpdateRecord is using for updating the record in CacheQueue
func(data *RecordTemplate) UpdateRecord(record string) {
    data.Records = append(data.Records, record)
}

// SortRecords is sorting the records field
func(data *RecordTemplate) SortRecords(){
    CQLocker.Lock()
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
    CQLocker.Unlock()
}

// GenCheckSumToQueue is using for generate the ckSum
func(data *RecordTemplate) GenCheckSumToQueue(traceID string, result map[string]string) {
    checkSumString := strings.Join(data.Records, "\n") + "\n"
    result[traceID] = MD5(checkSumString)
}

// CQLocker is a CacheQueue Locker
var CQLocker = sync.Mutex{}
