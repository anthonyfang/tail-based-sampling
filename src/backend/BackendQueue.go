package backend

import(
    "tail-based-sampling/src/common"
)

var resultWorkingQueue = make(map[string]*common.RecordTemplate)

var backendChan = make(chan string)

type traceInfoStruct struct {
    traceID     string;
    batchNo     int;
}
