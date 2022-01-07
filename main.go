package main

import (
	"cheese-theif/id"
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var tokenMap = sync.Map{}

type Stat struct {
	Rooms     int      `json:"rooms"`
	Users     int      `json:"users"`
	Total     int      `json:"total"`
	RoomNames []string `json:"room_names"`
}

type MemoryStat struct {
	Alloc      string `json:"alloc"`
	TotalAlloc string `json:"total_alloc"`
	Sys        string `json:"sys"`
	NumGC      string `json:"num_gc"`
}

func haveUser(token string) bool {
	if _, ok := tokenMap.Load(token); ok {
		return true
	}
	return false
}

func updateInfo(server *socketio.Server) {
	var rooms = server.Rooms("/game")
	server.BroadcastToNamespace("/stats", "update", Stat{
		Rooms:     len(rooms),
		Users:     len(getTokensFromMap()),
		Total:     server.Count(),
		RoomNames: rooms,
	})
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func getStats() MemoryStat {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStat{
		Alloc:      fmt.Sprintf("Alloc = %v MiB", bToMb(m.Alloc)),
		TotalAlloc: fmt.Sprintf("TotalAlloc = %v MiB", bToMb(m.TotalAlloc)),
		Sys:        fmt.Sprintf("Sys = %v MiB", bToMb(m.Sys)),
		NumGC:      fmt.Sprintf("NumGC = %v", m.NumGC),
	}
}

func main() {
	serverOptions := engineio.Options{
		SessionIDGenerator: id.NewNameGenerator(),
	}

	server := socketio.NewServer(&serverOptions)
	server.OnConnect("/", func(conn socketio.Conn) error {
		conn.SetContext("")
		updateInfo(server)
		color.Green("root:client connected: %s\n", conn.ID())
		return nil
	})

	server.OnDisconnect("/", func(conn socketio.Conn, s string) {
		defer func() {
			conn.Close()
		}()
		color.Red("disconnected:%s\n%s\n", conn.ID(), s)
	})

	server.OnConnect("/stats", func(conn socketio.Conn) error {
		updateInfo(server)
		conn.LeaveAll()
		server.JoinRoom("/stats", "manager", conn)
		fmt.Printf("/stats:client connected: %s\n", conn.ID())
		return nil
	})

	server.OnDisconnect("/stats", func(conn socketio.Conn, s string) {
		defer func() {
			conn.Close()
		}()
		fmt.Printf("/stats:client disconnected:%s\n", conn.ID())
	})

	server.OnConnect("/game", func(conn socketio.Conn) error {
		defer updateInfo(server)
		url := conn.URL()
		token := url.Query().Get("token")
		if token == "null" {
			token = id.GenerateToken()
		}

		if !haveUser(token) {
			tokenMap.Store(token, "")
		}

		conn.LeaveAll()
		server.JoinRoom("/game", "default-room", conn)
		conn.Emit("auth", token)
		color.Cyan("token:%s, length:%d", token, len(token))
		color.Green("/game: connected:%s", conn.ID())
		return nil
	})

	go func() {
		for {
			time.Sleep(time.Second * 5)
			server.BroadcastToNamespace("/stats", "memory", getStats())
		}
	}()

	server.OnDisconnect("/game", func(conn socketio.Conn, s string) {
		defer func() {
			conn.Close()
		}()

		updateInfo(server)
		url := conn.URL()
		token := url.Query().Get("token")
		color.Red("/game:disconnected:%s, token:%s", conn.ID(), token)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socket listen error:%s\n", err)
		}
	}()

	router := gin.New()
	router.GET("socket.io/*any", Cors(), gin.WrapH(server))
	router.POST("socket.io/*any", Cors(), gin.WrapH(server))
	router.StaticFS("/web", http.Dir("./web"))
	router.GET("/tokens", func(context *gin.Context) {
		context.JSON(http.StatusOK, getTokensFromMap())
	})

	router.Run(":8000")
}

func getTokensFromMap() map[string]interface{} {
	tokenList := make(map[string]interface{})
	tokenMap.Range(func(key, value interface{}) bool {
		tokenList[key.(string)] = value
		return true
	})

	return tokenList
}
