package common

import (
	cmap "github.com/orcaman/concurrent-map"
)

var TraceInfoStore = QuickStore{}

type QuickStore struct {
	idList      []string
	infoCache   []info
	storeMap    cmap.ConcurrentMap
	enableCache bool
}

type info struct {
	id      string
	records []string
}

func (q *QuickStore) New() {
	q.idList = []string{}
	q.storeMap = cmap.New()

	//--- for cache ---//
	q.enableCache = false
	q.infoCache = []info{}
}

func (q *QuickStore) Add(id string, val interface{}) {
	var found = false
	var newData = []string{}

	// ID List for Housekeep purpose
	data, found := q.storeMap.Get(id)
	if !found {
		q.idList = append(q.idList, id)
	}
	// ID List for Housekeep purpose

	if found {
		newData = append(data.([]string), val.(string))
	} else {
		newData = append(newData, val.(string))
	}
	q.storeMap.Set(id, newData)

	// if enableCache {

	// 	for i, v := range q.infoCache{
	// 		if v.id == id {
	// 			found=true
	// 			q.infoCache[i].records = append(q.infoCache[i].records, val.(string))
	// 		}
	// 	}

	// 	if !found{
	// 		if len(q.infoCache)>=20{
	// 			q.storeMap.Set(q.infoCache[0].id, q.infoCache[0].records)

	// 			var newInfo = &info{id:id, records: []string{val}}
	// 			q.infoCache=append(q.infoCache[1:len(q.infoCache)-1], newInfo)
	// 		}

	// 	}
	// }
}

func (q *QuickStore) Get(id string, result *[]string) (ok bool) {
	val, ok := q.storeMap.Get(id)
	if !ok {
		return ok
	}
	*result = append(*result, val.([]string)...)
	return true
}

func (q *QuickStore) HouskeepTill(keepTraceNumber int) {
	var toHousekeep = len(q.idList) - keepTraceNumber

	for i := 0; i < toHousekeep; i++ {
		q.storeMap.Remove(q.idList[i])
	}
	q.idList = q.idList[toHousekeep : len(q.idList)-1]
}
