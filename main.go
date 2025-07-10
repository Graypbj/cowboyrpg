package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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
			fmt.Printf("read error from %s: %v", playerID, err)
			break
		}

		msgTypeRaw, ok := msg["type"]
		if !ok {
			ws.WriteJSON(map[string]any{
				"type":  "error",
				"error": "Missing 'type' field",
			})
			continue
		}

		msgType, ok := msgTypeRaw.(string)
		if !ok {
			ws.WriteJSON(map[string]any{
				"type":  "error",
				"error": "'type' must be a string",
			})
			continue
		}
		msgType = strings.ToLower(msgType)

		fmt.Println(msg)

		switch msgType {
		case "create_party":
			handleCreateParty(player, msg)
		case "join_party":
			handleJoinParty(player, msg)
		case "choose_move":
			handleChooseMove(player, msg)
		default:
			ws.WriteJSON(map[string]any{
				"type":  "error",
				"error": fmt.Sprintf("Unknown message type: %s", msgType),
			})
		}
	}
	removePlayer(player)
}

func removePlayer(player *Player) {
	playersMu.Lock()
	delete(players, player.ID)
	playersMu.Unlock()

	if player.PartyID == "" {
		return
	}

	partiesMu.Lock()
	defer partiesMu.Unlock()

	party, ok := parties[player.PartyID]
	if !ok {
		return
	}

	for i, p := range party.Players {
		if p != nil && p.ID == player.ID {
			party.Players[i] = nil
		}
	}

	// If both players are nil, delete the party
	if party.Players[0] == nil && party.Players[1] == nil {
		delete(parties, party.ID)
		log.Printf("Party %s removed due to disconnection", party.ID)
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
