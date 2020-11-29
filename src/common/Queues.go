package common

import (
    "sync"
)

// RecordTemplate is a template for record down each line of trace record info
type RecordTemplate struct {
    hasError      bool;
    startLineNO   int;
    hasReport     bool;      // Send to backend Service
    records       []string
}

// CacheQueue is using cache all the records
var CacheQueue = make(map[string]*RecordTemplate)

// UpdateRecord is using for updating the record in CacheQueue
func(data *RecordTemplate) UpdateRecord(record string) {
    data.records = append(data.records, record)
}

// CQLocker is a CacheQueue Locker
var CQLocker = sync.Mutex{}
