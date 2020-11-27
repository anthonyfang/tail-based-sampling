package main

import (
    "fmt"
    "os"

    "github.com/gofiber/fiber/v2"

    BackendHandler "tail-based-sampling/src/backend"
    CliendHandler "tail-based-sampling/src/client"
)

func main() {
    app := fiber.New()
    port := GetEnvDefault("SERVER_PORT", "3000")

    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello Baby ~ Johnny is coming!")
    })

    app.Get("/ready", func(c *fiber.Ctx) error {
        return c.SendString(fmt.Sprintf("Server is running on port: %v", port))
    })

    SetParamterRouter(app, port)

    if port == "8002" {
        app.Post("/setWrongTraceId", BackendHandler.SetWrongTraceIDHandler)
    } else {
        app.Get("/getWrongTrace", CliendHandler.GetWrongTraceHandler)
    }

    app.Listen(":" + port)
}

// GetEnvDefault is using for getting enviroment variable with default value
func GetEnvDefault(key string, defVal string) string {
    val, ex := os.LookupEnv(key)
    if !ex {
        return defVal
    }
    return val
}

// SetParamterRouter is a router to handle the setParamter endpoint
func SetParamterRouter(app *fiber.App, port string) {
    app.Post("/setParameter", func(c *fiber.Ctx) error {
        if port == "8002" {
            return c.SendString("Backend Service Do not support!")
        }
        return c.SendString("OK!")
    })
}
