package backend

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
)

var clientPorts = []string{}

// FinishHandler is using for trigger calculation
func FinishHandler(c *fiber.Ctx) error {
    
    type Request struct {
        Port string         `json:"port"`;
    }
    var body Request
    err := c.BodyParser(&body)
    if err != nil {
        fmt.Println(err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "message": "Cannot parse Body JSON",
        })
    }

    clientPorts = append(clientPorts, body.Port)

    msg := "OK! Waiting for other ports"

    // TODO
    // if len(clientPorts) == 2 {
    if len(clientPorts) == 1 {
        fmt.Println("============= Result ================")
        for key, value := range resultWorkingQueue {
            fmt.Println("XXXXXXXXXXXXX ", key, " XXXXXXXXXXXXX")
            fmt.Println(value)
        }
        fmt.Println("============= END ================")

        msg = "OK! Start Upload now"
    }
    return c.SendString(msg)
}
