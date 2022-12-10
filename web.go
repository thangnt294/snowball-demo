package main

import (
	"fmt"
	"math/rand"

	"github.com/gofiber/fiber/v2"
)

type ApiMsg struct {
	Type int // 1: createTx, 2: listChain
	Data string
}

// startWebServer starts the web server and listens for API requests.
func startWebServer(network []Node) {
	app := fiber.New()

	app.Post("/createTx/:val", func(c *fiber.Ctx) error { // create a Tx in a network
		nodeid := rand.Intn(len(network))
		fmt.Printf("creating Tx in node %d\n", nodeid)
		network[nodeid].ApiChan <- ApiMsg{1, c.Params("val")}
		return c.SendString("OK")
	})

	app.Get("/chain", func(c *fiber.Ctx) error { // list all local chain in network
		for i := 0; i < len(network); i++ {
			network[i].ApiChan <- ApiMsg{2, ""}
		}
		return c.SendString("OK")
	})

	app.Listen(":3000")
}
