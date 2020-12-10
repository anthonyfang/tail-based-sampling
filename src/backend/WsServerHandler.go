package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"tail-based-sampling/src/common"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/gin-gonic/gin"

	"tail-based-sampling/src/trace"
)

type WsConn struct {
	*websocket.Conn
	Mux   sync.RWMutex
	Param map[string]string
}

func (p *WsConn) send(mt int, msg []byte) error {
	p.Mux.Lock()
	err := p.Conn.WriteMessage(mt, msg)
	p.Mux.Unlock()
	return err
}

func (p *WsConn) Params(key string) string {
	return p.Param[key]
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	MAX_CONCURRENCY = 30
)

var tmpChan = make(chan struct{}, MAX_CONCURRENCY)

var Mux = sync.RWMutex{}

var wsclients = []WsConn{}

func WsServerHandler(c *gin.Context) {
	r := c.Request
	w := c.Writer
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("err = %s\n", err)
		return
	}
	log.Println(c.Param("id"))   // 123
	log.Println(c.Param("type")) // 123
	log.Println(c.Query("v"))    // 1.0

	var wsc = WsConn{}
	wsc.Conn = conn
	wsc.Param = make(map[string]string)
	wsc.Param["id"] = c.Param("id")
	wsc.Param["type"] = c.Param("type")
	wsc.Param["v"] = c.Param("v")

	wsclients = append(wsclients, wsc)

	fmt.Sprintf("Websocket connected with %i successfully.", c.Param("id"))

	defer func() {
		// 发送websocket结束包
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		// 真正关闭conn
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
		}

		go func(msg []byte) {
			tmpChan <- struct{}{}
			//fmt.Println(msg)
			var payload2 common.Payload
			// err = json.Unmarshal(msg, &trace.PayloadMessage)
			// // fmt.Println(payload)
			// if err != nil {
			// 	log.Println("error:", err)
			// }

			payload := &trace.PayloadMessage{}
			err = proto.Unmarshal(msg, payload)

			if err != nil {
				err = json.Unmarshal(msg, &payload2)
				fmt.Println(payload2)
				fmt.Println(err)
				// log.Println("unmarshaling error: ", err)
			} else {
				switch payload.Action {
				case "SetWrongTraceID":
					batchNo, _ := strconv.Atoi(payload.ID)
					// var recordSlice []string
					// if payload.Records != nil {
					// 	for _, el := range payload.Records {
					// 		recordSlice = append(recordSlice, el.(string))
					// 	}
					// }
					processWrongTraceID(conn, batchNo, payload.Records)

				case "ReturnWrongTrace":
					// var recordSlice []string
					// if payload.Records != nil {
					// 	for _, el := range payload.Records.([]interface{}) {
					// 		recordSlice = append(recordSlice, el.(string))
					// 	}
					// }
					processWrongTrace(&wsc, payload.ID, payload.Records)
				case "SendFinished":
					common.FinishedChan <- payload.ID
				default:
					break
				}
			}
			<-tmpChan
		}(msg)
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

func processWrongTrace(ws *WsConn, traceID string, records []string) {

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
