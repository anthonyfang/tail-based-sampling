package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/gofiber/websocket/v2"

	BackendHandler "tail-based-sampling/src/backend"
	CliendHandler "tail-based-sampling/src/client"
	Common "tail-based-sampling/src/common"
)

func main() {

	app := fiber.New()
	port := Common.GetEnvDefault("SERVER_PORT", "3000")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello Baby ~ Johnny is coming!")
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.SendString(fmt.Sprintf("Server is running on port: %v", port))
	})

	if port == "8002" {
		app.Post("/setParameter", BackendHandler.SetParameterPostHandler)
		app.Get("/setParameter", BackendHandler.SetParameterGetHandler)
	} else {
		app.Post("/setParameter", CliendHandler.SetParameterPostHandler)
		app.Get("/setParameter", CliendHandler.SetParameterGetHandler)
	}

	if port == "8002" {
		app.Post("/setWrongTraceId", BackendHandler.SetWrongTraceIDHandler)
		app.Post("/finish", BackendHandler.FinishHandler)
	} else {
		app.Get("/getWrongTrace/:traceID", CliendHandler.GetWrongTraceHandler)
	}

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// c.Locals is added to the *websocket.Conn
		log.Println(c.Locals("allowed"))  // true
		log.Println(c.Params("id"))       // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""

		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			if err = c.WriteMessage(mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}

	}))

	app.Listen(":" + port)

}
