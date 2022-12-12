package web

import (
	"ava/network"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"ava/config"

	"github.com/gofiber/fiber/v2"
)

type ApiMsg struct {
	Type int // 1: createTx, 2: listChain
	Data string
}

// StartWebServer starts the web server and listens for API requests.
func StartWebServer(network []network.Node) {
	app := fiber.New()

	app.Post("/createTx/:val", func(c *fiber.Ctx) error { // create a tx in network
		randomNodeAddr := network[rand.Intn(len(network))].Addr

		url := fmt.Sprintf("http://localhost:%d/createTx", randomNodeAddr)
		type CreateTxRequest struct {
			Tx int `json:"tx"`
		}
		tx, _ := strconv.Atoi(c.Params("val"))
		req := CreateTxRequest{
			Tx: tx,
		}
		fmt.Printf("creating Tx %d in node %d\n", tx, randomNodeAddr)

		reqJson, err := json.Marshal(req)
		if err != nil {
			log.Printf("Error marshalling request: %v", err)
			return c.SendString("Error")
		}

		_, err = http.Post(url, "application/json", bytes.NewReader(reqJson))
		if err != nil {
			log.Printf("Error sending create tx request: %v", err)
			return c.SendString("Error")
		}
		return c.SendString("OK")
	})

	app.Get("/chains", func(c *fiber.Ctx) error { // list all local chain in network
		for i := 0; i < len(network); i++ {
			url := fmt.Sprintf("http://localhost:%d/listChain", network[i].Addr)
			go http.Get(url)
		}
		return c.SendString("OK")
	})

	app.Get("/neighbors/:nodeAddr", func(c *fiber.Ctx) error { // get neighbors of a node in network
		url := fmt.Sprintf("http://localhost:%s/neighbors", c.Params("nodeAddr"))
		go http.Get(url)
		return c.SendString("OK")
	})

	app.Get("/chain/:nodeAddr", func(c *fiber.Ctx) error { // list chain of a node in network
		url := fmt.Sprintf("http://localhost:%s/listChain", c.Params("nodeAddr"))
		go http.Get(url)
		return c.SendString("OK")
	})

	app.Listen(fmt.Sprintf(":%d", config.GlobalConfig.NetworkConf.WebPort))
}
