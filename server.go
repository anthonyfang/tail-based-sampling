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

	BackendHandler "tail-based-sampling/src/backend"
	CliendHandler "tail-based-sampling/src/client"
	Common "tail-based-sampling/src/common"
	pb "tail-based-sampling/src/trace"
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

func serveGRPC(l net.Listener) {
	grpcs := grpc.NewServer()
	pb.RegisterTraceServiceServer(grpcs, &BackendHandler.Server{})
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
			go BackendHandler.StartBackendProcess()
			c.String(200, "Server is running on port: %v", port)
		})
	} else {
		r.GET("/ready", func(c *gin.Context) {
			// go CliendHandler.WsConnection()
			go CliendHandler.StartClientProcess()
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
