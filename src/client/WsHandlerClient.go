package client

import (
	"encoding/json"
	"fmt"
	"log"
	"tail-based-sampling/src/common"
	"tail-based-sampling/src/trace"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/websocket"
)

var ws1 *websocket.Conn
var ws2 *websocket.Conn

// func gRPCconnect() {

// 	var conn *grpc.ClientConn
// 	conn, err := grpc.Dial(":8002", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("did not connect: %s", err)
// 	}
// 	defer conn.Close()

// 	c := trace.NewTraceServiceClient(conn)

// 	var port = common.GetEnvDefault("SERVER_PORT", "3000")

// 	response, err := c.SetWrongTraceID(context.Background(), &trace.TraceIDsMesssage{PortID: port, Records: []string{"test1", "test2", "test3"}})
// 	if err != nil {
// 		log.Fatalf("Error when calling SayHello: %s", err)
// 	}
// 	log.Printf("Response from server: %s", response.Body)

// 	ctx := stream.Context()
// 	done := make(chan bool)

// 	// first goroutine sends random increasing numbers to stream
// 	// and closes int after 10 iterations
// 	go func() {
// 		for i := 1; i <= 10; i++ {
// 			// generate random nummber and send it to stream
// 			rnd := int32(rand.Intn(i))
// 			req := pb.Request{Num: rnd}
// 			if err := stream.Send(&req); err != nil {
// 				log.Fatalf("can not send %v", err)
// 			}
// 			log.Printf("%d sent", req.Num)
// 			time.Sleep(time.Millisecond * 200)
// 		}
// 		if err := stream.CloseSend(); err != nil {
// 			log.Println(err)
// 		}
// 	}()

// 	// second goroutine receives data from stream
// 	// and saves result in max variable
// 	//
// 	// if stream is finished it closes done channel
// }

func WsConnection() {
	var (
		err1 error
		err2 error
	)
	var port = common.GetEnvDefault("SERVER_PORT", "3000")
	var origin = "http://127.0.0.1:" + port + "/"
	var url_1 = "ws://127.0.0.1:8002/ws/" + port + "/ids"
	var url_2 = "ws://127.0.0.1:8002/ws/" + port + "/info"

	ws1, err1 = websocket.Dial(url_1, "", origin)
	if err1 != nil {
		log.Fatalln(err1)
	}
	ws2, err2 = websocket.Dial(url_2, "", origin)
	if err2 != nil {
		log.Fatalln(err2)
	}

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

		if payload.Action == "GetWrongTrace" {
			ReturnWrongTrace(ws2, payload.ID)
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

	// var payload = new(common.Payload)
	// if batchNo > 70 {
	// 	fmt.Println("traceID:", traceID)
	// 	fmt.Println(data.Records)
	// }
	payload := &trace.PayloadMessage{
		Action:  "ReturnWrongTrace",
		ID:      traceID,
		Records: data.Records,
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
