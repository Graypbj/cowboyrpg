package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const port = ":8080"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	playerID := uuid.New().String()
	player := &Player{
		ID:   playerID,
		Conn: ws,
		HP:   100,
	}
	playersMu.Lock()
	players[playerID] = player
	playersMu.Unlock()

	for {
		var msg map[string]any
		if err := ws.ReadJSON(&msg); err != nil {
			fmt.Printf("read error: %v", err)
			break
		}

		fmt.Println(msg)

		switch msg["type"] {
		case "create_party":
			handleCreateParty(player, msg)
		case "join_party":
			handleJoinParty(player, msg)
		case "choose_move":
			handleChooseMove(player, msg)
		default:
			ws.WriteJSON(map[string]any{
				"type":  "error",
				"error": fmt.Sprintf("Unknown message type: %s", msg["type"]),
			})
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	fmt.Printf("Server started on %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic("ListenAndServer: " + err.Error())
	}
}
