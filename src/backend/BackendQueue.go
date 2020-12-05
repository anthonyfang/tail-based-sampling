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

var wg sync.WaitGroup
var resultQueueLocker = sync.Mutex{}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
