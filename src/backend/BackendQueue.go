package backend

import(
    "sync"
    "tail-based-sampling/src/common"
)

var resultWorkingQueue = make(map[string]*common.RecordTemplate)
var resultQueue = make(map[string]string)

var backendChan = make(chan string)

type traceInfoStruct struct {
    traceID     string;
    batchNo     int;
}

var bufferChan = make(chan string, 15)
var wg sync.WaitGroup
var resultQueueLocker = sync.Mutex{}
