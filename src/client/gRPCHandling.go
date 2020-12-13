package client

import (
	"context"
	"flag"
	"io"
	"log"
	"tail-based-sampling/src/common"
	"tail-based-sampling/src/trace"
	pb "tail-based-sampling/src/trace"
	"time"

	"google.golang.org/grpc"
)

var gRPCstream *pb.TraceService_TraceChatClient
var gRPCclient *pb.TraceServiceClient

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:8002", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
)

func runTraceChat() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewTraceServiceClient(conn)

	gRPCclient = &client

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer cancel()
	stream, err := client.TraceChat(ctx)

	gRPCstream = &stream
	if err != nil {
		log.Fatalf("%v.TraceChat(_) = _, %v", client, err)
	}
	clientIDmsg := &pb.PayloadMessage{Action: "SetClientID", ID: common.GetEnvDefault("SERVER_PORT", ""), Records: []string{}}
	// for _, msg := range messages {
	if err := stream.Send(clientIDmsg); err != nil {
		log.Fatalf("Failed to set Client ID: %v", err)
	}
	// }
	// waitc := make(chan struct{})
	// go func() {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// read done.
			// close(waitc)
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		if in.Action == "GetWrongTrace" {
			gRPCReturnWrongTrace(in.ID)
		}

		// log.Printf("Got message for action %s %s", in.Action, in.ID)
	}
	// }()
	// <-waitc
	// stream.CloseSend()
}

func gRPCReturnWrongTrace(traceID string) {

	var result = []string{}
	common.TraceInfoStore.Get(traceID, &result)

	payload := &trace.PayloadMessage{
		Action:  "ReturnWrongTrace",
		ID:      traceID,
		Records: result,
	}
	// payload.ReturnWrongTraceGen(traceID, payload)

	if err := (*gRPCstream).Send(payload); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(data.Records)
}

// if *tls {
// 	if *caFile == "" {
// 		*caFile = data.Path("x509/ca_cert.pem")
// 	}
// 	creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
// 	if err != nil {
// 		log.Fatalf("Failed to create TLS credentials %v", err)
// 	}
// 	opts = append(opts, grpc.WithTransportCredentials(creds))
// } else {
// 	opts = append(opts, grpc.WithInsecure())
// }
