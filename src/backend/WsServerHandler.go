package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"tail-based-sampling/src/common"
	"time"

	"github.com/gofiber/websocket/v2"
)

type WsConn struct {
	*websocket.Conn
	Mux sync.RWMutex
}

func (p *WsConn) send(mt int, msg []byte) error {
	p.Mux.Lock()
	err := p.Conn.WriteMessage(mt, msg)
	p.Mux.Unlock()
	return err
}

// func (p *WsConn) read() (int, []byte, error) {
// 	p.Mux.Lock()
// 	mt, msg, err := p.ReadMessage()
// 	p.Mux.Unlock()
// 	return mt, msg, err
// }

var Mux = sync.RWMutex{}

var wsclients = []WsConn{}

func WsServerHandler(c *websocket.Conn) {
	// c.Locals is added to the *websocket.Conn
	log.Println(c.Locals("allowed"))  // true
	log.Println(c.Params("id"))       // 123
	log.Println(c.Params("type"))     // 123
	log.Println(c.Query("v"))         // 1.0
	log.Println(c.Cookies("session")) // ""

	var wsc = new(WsConn)
	wsc.Conn = c
	wsclients = append(wsclients, *wsc)

	fmt.Sprintf("Websocket connected with %i successfully.", c.Params("id"))

	// var (
	// 	// mt  int
	// 	msg []byte
	// 	err error
	// )

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
		}
		//fmt.Println(msg)
		var payload common.Payload
		err = json.Unmarshal(msg, &payload)
		// fmt.Println(payload)
		if err != nil {
			log.Println("error:", err)
		}

		switch payload.Action {
		case "SetWrongTraceID":
			batchNo, _ := strconv.Atoi(payload.ID)

			var recordSlice []string
			if payload.Records != nil {
				for _, el := range payload.Records.([]interface{}) {
					recordSlice = append(recordSlice, el.(string))
				}
			}
			processWrongTraceID(c, batchNo, recordSlice)

		case "ReturnWrongTrace":
			var recordSlice []string
			if payload.Records != nil {
				for _, el := range payload.Records.([]interface{}) {
					recordSlice = append(recordSlice, el.(string))
				}
			}
			processWrongTrace(c, payload.ID, recordSlice)
		case "SendFinished":
			common.FinishedChan <- payload.ID
		default:
			break
		}

		// log.Printf("recv: %s", msg)
		time.Sleep(100)
	}

}

func wsWriteLoop() {
	var (
		msg []byte
		err error
	)

	for messageToSend := range common.ServerSendWSChan {
		wg.Add(2)
		var payload common.Payload
		traceID := messageToSend.(string)
		payload.GetWrongTraceGen(traceID)

		msg, err = json.Marshal(payload)

		for _, client := range wsclients {
			if client.Params("type") == "info" {
				if err = client.send(websocket.BinaryMessage, msg); err != nil {
					log.Println("write:", err)
				}
			}
		}
	}
}

func processWrongTraceID(ws *websocket.Conn, batchNo int, records []string) {
	for _, v := range records {
		BackendTraceIDQueue.Store(v, batchNo)
	}

	common.BatchReceivedCountChan <- batchNo
}

func processWrongTrace(ws *websocket.Conn, traceID string, records []string) {

	// Push into the cache server
	if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
		fmt.Println("-----traceInfo start----")
		fmt.Println(records)
		fmt.Println("-----traceInfo end----")
	}

	data := &common.RecordTemplate{HasError: true, BatchNo: 0, Records: []string{}}
	data.Records = records

	if len(records) > 0 {
		common.SetTraceInfo(traceID+"-"+ws.Params("id"), data)
	}

	common.ReceivedTraceInfoChan <- traceID
	// fmt.Println(data.Records)
	// Request all the clients to get all the bad trace info
}
