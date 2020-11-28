package main

import (
    "fmt"
    "os"

    "github.com/gofiber/fiber/v2"

    BackendHandler "tail-based-sampling/src/backend"
    CliendHandler "tail-based-sampling/src/client"
    CommonHandler "tail-based-sampling/src/common"
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

    app.Post("/setParameter", CommonHandler.SetParameterHandler)

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
