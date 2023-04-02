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
	"text/template"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"github.com/mdp/qrterminal/v3"
)

const (
	PORT  = "40444"
	DEBUG = false
)

type Vel struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	if DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	host, _ := getNetworkInterfaces()
	printQRCode(&host)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	// tmpl["index"] = template.Must(template.ParseFiles("templates/index.gohtml"))
	router.GET("/", func(c *gin.Context) {
		sendTemplate(
			c,
			template.Must(template.ParseFiles("templates/index.gohtml")),
			host,
		)
	})
	router.GET("/ws", func(c *gin.Context) { serveWs(c.Writer, c.Request) })

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

func sendTemplate(c *gin.Context, template *template.Template, host string) {
	data := struct {
		SERVER_URL  string
		SERVER_PORT string
	}{
		SERVER_URL:  host + ".local",
		SERVER_PORT: PORT,
	}
	err := template.Execute(c.Writer, data)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	// ? upgrade this connection to a WebSocket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		debLog(err)
		return
	}
	defer ws.Close()
	debPrintln("New connexion : " + ws.RemoteAddr().String())
	// ? listen indefinitely for new messages coming
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			debLog(err)
			return
		}
		if strings.HasPrefix(string(p), "pos") { // ? mice position
			p = p[3:]
			var vel Vel
			err = json.Unmarshal(p, &vel)
			if err != nil {
				debLog(err)
				return
			}
			debPrintln("vel : " + string(p))
			x, y := robotgo.GetMousePos()
			robotgo.Move(x+int(vel.X), y+int(vel.Y))
		} else if strings.HasPrefix(string(p), "left") { //  ? left click
			debPrintln("left")
			robotgo.Click("left")
		} else if strings.HasPrefix(string(p), "right") { // ? right click
			debPrintln("right")
			robotgo.Click("right")
		} else if strings.HasPrefix(string(p), "hello") { // ? hello
		} else {
			debPrintln("unknown message : " + string(p))
		}
	}
}

// ? could be nice to make a debug package
func debLog(err error) {
	if DEBUG {
		_, file, line, _ := runtime.Caller(1)
		log.Println(fmt.Sprint("file ", file, ", line ", line, " : ", err))
	}
}

func debPrintln(err string) {
	if DEBUG {
		log.Println(err)
	}
}
