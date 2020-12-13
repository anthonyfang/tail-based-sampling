package backend

import (
	"sync"
	"tail-based-sampling/src/common"

	cmap "github.com/orcaman/concurrent-map"
)

var resultWorkingQueue = make(map[string]*common.RecordTemplate)
var resultQueue = make(map[string]string)

var wg sync.WaitGroup
var resultQueueLocker = sync.Mutex{}

var BackendTraceIDQueue = cmap.New()
var BackendBatchQueue = cmap.New()
var BackendReceivedTraceInfo = cmap.New()
var BackendReceivedTraceInfoCount = cmap.New()
var BackendTraceInfoCache = cmap.New()
