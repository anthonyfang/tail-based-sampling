package trace

import (
	"log"

	"golang.org/x/net/context"
)

type Server struct {
}

func (s *Server) SetWrongTraceID(ctx context.Context, in *PayloadMessage) (*Response, error) {
	log.Printf("Receive message body from client: %s", in.ID)
	return &Response{Body: "Hello From the Server!"}, nil
}

func (s *Server) FindTraceInfo(ctx context.Context, in *PayloadMessage) (*PayloadMessage, error) {
	log.Printf("Receive message body from client: %s", in.ID)
	return &PayloadMessage{Records: []string{"Hello From the Server! --- 1", "Hello From the Server! --- 2", "Hello From the Server! --- 3"}}, nil
}
