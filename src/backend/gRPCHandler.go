package backend

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"tail-based-sampling/src/common"
	"tail-based-sampling/src/trace"
)

type Server struct{}

var gRPCstreams = []*trace.TraceService_TraceChatServer{}

func (s *Server) TraceChat(stream trace.TraceService_TraceChatServer) error {

	gRPCstreams = append(gRPCstreams, &stream)

	var clientID = ""
	// messages := []*pb.PayloadMessage{
	// 	{Action: "GetWrongTrace", ID: "123123123123", Records: []string{}},
	// 	{Action: "GetWrongTrace", ID: "3453453453453", Records: []string{}},
	// }
	// for _, msg := range messages {
	// 	if err := stream.Send(msg); err != nil {
	// 		log.Fatalf("Failed to send a note: %v", err)
	// 	}
	// }

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Fatalf("Failed to receive a message : %v", err)
		}
		if err != nil {
			log.Fatalf("Failed : %v", err)
			break
		}

		switch in.Action {
		case "SetClientID":
			clientID = in.ID
		case "SetWrongTraceID":
			batchNo, _ := strconv.Atoi(in.ID)
			// log.Printf("Got message for action %s %s", in.Action, in.ID)
			// fmt.Println(in.Records)
			gRPCProcessWrongTraceID(batchNo, in.Records)

		case "ReturnWrongTrace":

			gRPCProcessWrongTrace(clientID, in.ID, in.Records)
		case "SendFinished":
			common.FinishedChan <- in.ID
		default:
			break
		}

	}

	return nil
}

func gRPCWriteLoop() {
	var err error

	for messageToSend := range common.ServerSendWSChan {
		wg.Add(2)
		payload := &trace.PayloadMessage{
			Action:  "GetWrongTrace",
			ID:      messageToSend.(string),
			Records: []string{},
		}

		for _, stream := range gRPCstreams {
			if err = (*stream).Send(payload); err != nil {
				log.Println("write:", err)
			}
		}
	}
}

func gRPCProcessWrongTraceID(batchNo int, records []string) {

	for _, v := range records {
		BackendTraceIDQueue.Set(v, batchNo)
	}

	common.BatchReceivedCountChan <- batchNo
}

func gRPCProcessWrongTrace(clientID string, traceID string, records []string) {

	// Push into the cache server
	if common.IS_DEBUG && traceID == common.DEBUG_TRACE_ID {
		fmt.Println("-----traceInfo start----")
		fmt.Println(records)
		fmt.Println("-----traceInfo end----")
	}

	data := &common.RecordTemplate{HasError: true, BatchNo: 0, Records: []string{}}
	data.Records = records

	common.SetTraceInfo(traceID+"-"+clientID, data)

	common.ReceivedTraceInfoChan <- traceID
	// fmt.Println(data.Records)
	// Request all the clients to get all the bad trace info
}
