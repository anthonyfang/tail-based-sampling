package main

import (
    "fmt"

    "github.com/gofiber/fiber/v2"

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

    app.Post("/setParameter", Common.SetParameterHandler)

    if port == "8002" {
        app.Post("/setWrongTraceId", BackendHandler.SetWrongTraceIDHandler)
    } else {
        app.Get("/getWrongTrace/:id", CliendHandler.GetWrongTraceHandler)
    }

    app.Listen(":" + port)
}

