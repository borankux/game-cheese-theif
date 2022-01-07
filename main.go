package main

import (
	"cheese-theif/id"
	"fmt"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"log"
	"net/http"
	"sync"
)

var tokenMap = sync.Map{}

func haveUser(token string) bool {
	if _, ok := tokenMap.Load(token); ok {
		return true
	}
	return false
}


func main() {
	serverOptions := engineio.Options{
		SessionIDGenerator: id.NewNameGenerator(),
	}

	server := socketio.NewServer(&serverOptions)
	server.OnConnect("/", func(conn socketio.Conn) error {
		url := conn.URL()
		token := url.Query().Get("token")
		if !haveUser(token) {
			conn.Emit("/token")
		}
		userID := conn.ID()
		fmt.Printf("connected:%s\n", userID)
		return nil
	})

	server.OnDisconnect("/", func(conn socketio.Conn, reason string) {
		userID := conn.ID()
		fmt.Printf("disconnected:%s, reason:%s\n", userID, reason)
	})

	server.OnEvent("/", "msg", func(s socketio.Conn, msg string) {
		fmt.Printf("%s, %s\n", msg, s.ID())
		server.BroadcastToNamespace("/", "broadcast", fmt.Sprintf("%s says:%s", s.ID(), msg))
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socket listen error:%s\n", err)
		}
	}()

	router := gin.New()
	router.GET("socket.io/*any", gin.WrapH(server))
	router.POST("socket.io/*any", gin.WrapH(server))
	router.StaticFS("/web", http.Dir("./web"))
	router.Run(":8000")
}
