package client

import (
	"encoding/json"
	"fmt"
	"log"
	"tail-based-sampling/src/common"
	"time"

	"golang.org/x/net/websocket"
)

var ws1 *websocket.Conn
var ws2 *websocket.Conn

func WsConnection() {

	var port = common.GetEnvDefault("SERVER_PORT", "3000")
	var origin = "http://127.0.0.1:" + port + "/"
	var url_1 = "ws://127.0.0.1:8002/ws/" + port + "/ids"
	var url_2 = "ws://127.0.0.1:8002/ws/" + port + "/info"

	ws1, _ = websocket.Dial(url_1, "", origin)
	ws2, _ = websocket.Dial(url_2, "", origin)

	fmt.Println("Websocket connected with backend successfully.")
	// go wsProcessing()

	defer ws1.Close()
	defer ws2.Close()

	for {
		// message, _ := <-common.ClientSendWSChan
		//common.ClientRecvWSChan <- string(msg)

		//------ Receive msg -------//
		var msg = make([]byte, 128)
		total, err := ws2.Read(msg)
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
			ReturnWrongTrace(ws2, payload.ID)

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
