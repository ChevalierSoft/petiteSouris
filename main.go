package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"

	"github.com/gorilla/websocket"
)

type Vel struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

func printIp() {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	fmt.Printf("HOST %v :\n", host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Println("\t", ipv4)
		}
	}
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/ws", func(c *gin.Context) {
		go serveWs(c.Writer, c.Request)
	})

	router.Static("/public", "./public")

	go printIp()
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
		// mice position
		if strings.HasPrefix(string(p), "pos") {
			p = p[3:]
			var vel Vel
			err = json.Unmarshal(p, &vel)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("vel : ", vel)
				x, y := robotgo.GetMousePos()
				robotgo.Move(x+int(vel.X), y+int(vel.Y))
			}
		} else if strings.HasPrefix(string(p), "left") {
			fmt.Printf("left\n")
			robotgo.Click("left")
		} else if strings.HasPrefix(string(p), "right") {
			fmt.Printf("right\n")
			robotgo.Click("right")
		} else {
			fmt.Println("not supported : ", string(p))
		}
	}
}
