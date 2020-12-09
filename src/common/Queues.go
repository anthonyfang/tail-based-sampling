package common

import (
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

// RecordTemplate is a template for record down each line of trace record info
type RecordTemplate struct {
	HasError    bool
	BatchNo     int
	LifeTime    int
	Records     []string
	SyncRecords sync.Map
}

// CacheQueue is to store the records
var CacheQueue = sync.Map{}

// BadTraceIDList is recording down the bad trace IDs
var BadTraceIDList = []string{}

var Wg sync.WaitGroup

// BadTraceList is a list record down the bad trace
var BadTraceList = make(map[string]*RecordTemplate)

// UpdateRecord is using for updating the record in cache
func (data *RecordTemplate) UpdateRecord(record string) {
	data.Records = append(data.Records, record)
}

// SortRecords is sorting the records field
func (data *RecordTemplate) SortRecords() {
	// bubbleSort
	// len := len(data.Records)
	// for i := 0; i < len-1; i++ {
	// 	for j := 0; j < len-1-i; j++ {
	// 		arrJ, _ := strconv.Atoi(strings.Split(data.Records[j], "|")[1])
	// 		arrJ1, _ := strconv.Atoi(strings.Split(data.Records[j+1], "|")[1])

	// 		if arrJ > arrJ1 {
	// 			data.Records[j], data.Records[j+1] = data.Records[j+1], data.Records[j]
	// 		}
	// 	}
	// }
	quicksort(data.Records)
}

func quicksort(a []string) []string {
	if len(a) < 2 {
		return a
	}
	left, right := 0, len(a)-1
	pivot := rand.Int() % len(a)
	a[pivot], a[right] = a[right], a[pivot]

	for i, _ := range a {
		time_i, _ := strconv.Atoi(strings.Split(a[i], "|")[1])
		time_right, _ := strconv.Atoi(strings.Split(a[right], "|")[1])
		if time_i < time_right {
			a[left], a[i] = a[i], a[left]
			left++
		}
	}
	a[left], a[right] = a[right], a[left]
	quicksort(a[:left])
	quicksort(a[left+1:])
	return a
}

// GenCheckSumToQueue is used for generate the ckSum
func (data *RecordTemplate) GenCheckSumToQueue(traceID string, result map[string]string) {
	checkSumString := strings.Join(data.Records, "\n") + "\n"
	result[traceID] = MD5(checkSumString)
}

// GetTraceInfo is getting the traceInfo
func GetTraceInfo(traceID string) *RecordTemplate {
	traceCacheInfo, _ := CacheQueue.Load(traceID)
	traceInfo := &RecordTemplate{}
	if traceCacheInfo != nil {
		traceInfo = traceCacheInfo.(*RecordTemplate)
	}
	return traceInfo
}

// SetTraceInfo is setting the traceInfo
func SetTraceInfo(traceID string, data *RecordTemplate) {
	CacheQueue.Store(traceID, data)
}
