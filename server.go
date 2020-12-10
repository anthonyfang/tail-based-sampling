package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soheilhy/cmux"

	"google.golang.org/grpc"

	"context"

	BackendHandler "tail-based-sampling/src/backend"
	chat "tail-based-sampling/src/chat"
	CliendHandler "tail-based-sampling/src/client"
	Common "tail-based-sampling/src/common"
)

var port string

func main() {
	port = Common.GetEnvDefault("SERVER_PORT", "3000")
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	m := cmux.New(l)

	httpl := m.Match(cmux.HTTP1Fast())
	grpcl := m.Match(cmux.Any())

	go serveGRPC(grpcl)
	go serveHTTPAndWs(httpl)

	if err := m.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		panic(err)
	}
	fmt.Println("Server listening: ", port)
}

type grpcServer struct{}

func (s *grpcServer) SayHello(ctx context.Context, in *chat.Message) (
	*chat.Message, error) {

	fmt.Printf("request:%v\n", in)
	// return &chat.Message{Message: "Hello " + in.Name + " from cmux"}, nil
	return &chat.Message{Body: "Hello From Client!"}, nil
}

func serveGRPC(l net.Listener) {
	grpcs := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcs, &grpcServer{})
	if err := grpcs.Serve(l); err != cmux.ErrListenerClosed {
		panic(err)
	}
}

func serveHTTPAndWs(l net.Listener) {
	r := gin.Default()

	// websocket echo
	r.Any("/ws/:id/:type", BackendHandler.WsServerHandler)

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello Baby ~ Johnny is coming!")
		// io.Copy(c.Writer, c.Request.Body)
	})

	if port == "8002" {
		r.GET("/ready", func(c *gin.Context) {
			c.String(200, "Server is running on port: %v", port)
		})
	} else {
		r.GET("/ready", func(c *gin.Context) {
			go CliendHandler.WsConnection()
			c.String(200, "Server is running on port: %v", port)
		})
	}

	if port == "8002" {
		r.GET("/setParameter", BackendHandler.SetParameterGetHandler)
		r.POST("/setParameter", BackendHandler.SetParameterPostHandler)
	} else {
		r.GET("/setParameter", CliendHandler.SetParameterGetHandler)
		r.POST("/setParameter", CliendHandler.SetParameterPostHandler)
	}

	s := &http.Server{
		Handler: r,
	}

	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		panic(err)
	}
}
