package main

import (
	"github.com/gin-gonic/gin"
	socket "github.com/googollee/go-socket.io"
	"log"
	"net/http"
)

func main() {
	router := gin.New()


	server := socket.NewServer(nil)
	server.OnConnect("/", func(s socket.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})


	server.OnEvent("/", "notice", func(s socket.Conn, msg string) {
		log.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/chat", "msg", func(s socket.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socket.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socket.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socket.Conn, reason string) {
		log.Println("closed", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socket listen error: %s\n", err)
		}
	}()
	defer server.Close()

	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))
	router.StaticFS("/public", http.Dir("./public"))
	if err := router.Run(":8000"); err != nil {
		log.Fatal("failed run app: ", err)
	}
}
