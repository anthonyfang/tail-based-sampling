package client

import (
	"encoding/json"
	"fmt"
	"log"
	"tail-based-sampling/src/common"
	"time"

	"golang.org/x/net/websocket"
)

var ws *websocket.Conn

func wsConnection() {

	var port = common.GetEnvDefault("SERVER_PORT", "3000")
	var origin = "http://127.0.0.1:" + port + "/"
	var url = "ws://127.0.0.1:8002/ws/" + port

	ws, _ = websocket.Dial(url, "", origin)

	fmt.Println("Websocket connected with backend successfully.")
	// go wsProcessing()

	defer ws.Close()

	for {
		// message, _ := <-common.ClientSendWSChan
		//common.ClientRecvWSChan <- string(msg)

		//------ Receive msg -------//
		var msg = make([]byte, 128)
		total, err := ws.Read(msg)
		if err != nil {
			log.Fatal(err)
		}

		var payload common.Payload
		// fmt.Println(msg[:total])

		err = json.Unmarshal(msg[:total], &payload)
		if err != nil {
			log.Fatalln(err)
		}

		switch payload.Action {
		case "GetWrongTrace":
			ReturnWrongTrace(ws, payload.ID)

		default:
			break
		}

		time.Sleep(100)
	}
}

func ReturnWrongTrace(ws *websocket.Conn, traceID string) {
	// batchNo, _ := strconv.Atoi(c.Params("batchNo"))
	data := &common.RecordTemplate{HasError: true, BatchNo: 0, Records: []string{}}
	traceInfo := common.GetTraceInfo(traceID)

	if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
		fmt.Println(traceInfo)
	}
	if traceInfo != nil {
		traceInfo.SyncRecords.Range(func(k, v interface{}) bool {
			traceInfo.Records = append(traceInfo.Records, k.(string))
			return true
		})
		data = traceInfo
	}

	defer common.CacheQueue.Delete(traceID)

	var payload = new(common.Payload)
	payload.ReturnWrongTraceGen(traceID, data.Records)

	msg, _ := json.Marshal(payload)

	_, err := ws.Write(msg)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(data.Records)
}
