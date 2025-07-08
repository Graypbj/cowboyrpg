package main

import (
	"log"
	"net/http"

	"github.com/Graypbj/game"
	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/ws", websocket.Handler(game.HandleWebSocket))

	log.Println("Listening on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
