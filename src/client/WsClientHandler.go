package client

import (
	"encoding/json"
	"fmt"
	"log"
	"tail-based-sampling/src/common"
	"tail-based-sampling/src/trace"
	"time"

	"golang.org/x/net/websocket"
	"google.golang.org/protobuf/proto"
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
	var result = []string{}
	common.TraceInfoStore.Get(traceID, &result)
	// defer common.CacheQueue.Delete(traceID)
	payload := &trace.PayloadMessage{
		Action:  "ReturnWrongTrace",
		ID:      traceID,
		Records: result,
	}
	// payload.ReturnWrongTraceGen(traceID, payload)
	// msg, _ := json.Marshal(payload)
	msg, err := proto.Marshal(payload)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	_, err = ws.Write(msg)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(data.Records)
}
