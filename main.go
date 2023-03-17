package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"github.com/mdp/qrterminal/v3"
)

const (
	PORT = "8080"
)

type Vel struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	host, _ := getNetworkInterfaces()
	printQRCode(&host)

	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/public/index.html"+"?host="+host+".local")
	})
	router.GET("/ws", func(c *gin.Context) { serveWs(c.Writer, c.Request) })
	router.Static("/public", "./public")
	router.Run(":" + PORT)
}

func getNetworkInterfaces() (string, []net.IP) {
	host, err := os.Hostname()
	if err != nil {
		if _, file, line, ok := runtime.Caller(0); ok {
			log.Panic(fmt.Sprint("file ", file, ", line", line, " : ", err))
		}
		return "", nil
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		if _, file, line, ok := runtime.Caller(0); ok {
			log.Panic(fmt.Sprint("file ", file, ", line", line, " : ", err))
		}
		return "", nil
	}
	fmt.Printf("HOST %v :\n", host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Println("\t", ipv4)
		}
	}
	return host, addrs
}

func printQRCode(host *string) {
	fmt.Println("http://" + *host + ".local" + ":" + PORT)
	qrterminal.GenerateWithConfig("http://"+*host+".local"+":"+PORT, qrterminal.Config{
		Level:     qrterminal.M,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	})
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	// ? upgrade this connection to a WebSocket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		// if _, file, line, ok := runtime.Caller(0); ok {
		// 	log.Println(fmt.Sprint("file ", file, ", line", line, " : ", err))
		// }
		return
	}
	defer ws.Close()
	log.Println("new connection from : ", ws.RemoteAddr().String())
	// ? listen indefinitely for new messages coming
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			// if _, file, line, ok := runtime.Caller(0); ok {
			// 	log.Println(fmt.Sprint("file ", file, ", line", line, " : ", err))
			// }
			break
		}
		if strings.HasPrefix(string(p), "pos") { // ? mice position
			p = p[3:]
			var vel Vel
			err = json.Unmarshal(p, &vel)
			if err != nil {
				log.Println(err)
			}
			fmt.Println("vel : ", vel)
			x, y := robotgo.GetMousePos()
			robotgo.Move(x+int(vel.X), y+int(vel.Y))
		} else if strings.HasPrefix(string(p), "left") { //  ? left click
			fmt.Printf("left\n")
			robotgo.Click("left")
		} else if strings.HasPrefix(string(p), "right") { // ? right click
			fmt.Printf("right\n")
			robotgo.Click("right")
		} else {
			fmt.Println("not supported : ", string(p))
		}
	}
}
