package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

type Vel struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	// load the folder containing our static web assets
	router.LoadHTMLGlob("public/*")

	// serve the index page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/ws", func(c *gin.Context) {
		serveWs(c.Writer, c.Request)
	})
	router.Run(":8080")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(p))

		// if the first 3 characters of the message are "mos"
		if strings.HasPrefix(string(p), "mos") {
			// remove the first 3 characters
			p = p[3:]
			// convert string to Vel struct
			var vle Vel
			json.Unmarshal(p, &vle)
			fmt.Println(vle.X, vle.Y)
		}

		// convert string to Vel struct

		// var vle Vel
		// json.Unmarshal(p, &vle)
		// fmt.Println(vle.X, vle.Y)

		// robotgo.KeyTap(string(p))
	}
}
