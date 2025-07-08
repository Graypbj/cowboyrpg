package game

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

type Player struct {
	ID      string
	Name    string
	Class   string
	Conn    *websocket.Conn
	PartyID string
	IsReady bool
}

var (
	playersMu sync.Mutex
	players   = make(map[string]*Player)

	partiesMu sync.Mutex
	parties   = make(map[string]*Party)
)

// Incoming JSON format
type IncomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Outgoing
type OutgoingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func HandleWebSocket(ws *websocket.Conn) {
	playerID := uuid.New().String()
	player := &Player{
		ID:   playerID,
		Conn: ws,
	}

	playersMu.Lock()
	players[playerID] = player
	playersMu.Unlock()

	log.Printf("Player %s connected", playerID)

	defer func() {
		playersMu.Lock()
		delete(players, playerID)
		playersMu.Unlock()
		ws.Close()
		log.Printf("Player %s disconnected", playerID)
	}()

	for {
		var msg IncomingMessage
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			log.Printf("Receive error: %v", err)
			break
		}

		handleMessage(player, msg)
	}
}

func handleMessage(p *Player, msg IncomingMessage) {
	switch msg.Type {
	case "join_party":
		var data struct {
			PartyID string `json:"partyId"`
			Name    string `json:"name"`
			Class   string `json:"class"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			sendError(p, "invalid join_party payload")
			return
		}
		joinParty(p, data.PartyID, data.Name, data.Class)

	case "player_action":
		var data struct {
			ActionType string `json:"action_type"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			sendError(p, "invalid player_action payload")
			return
		}
		partiesMu.Lock()
		party, ok := parties[p.PartyID]
		partiesMu.Unlock()
		if !ok || party.Battle == nil {
			sendError(p, "No active battle in your party")
			return
		}
		party.Battle.HandleAction(p, data.ActionType)
	default:
		sendError(p, "unknown message type")
	}
}

func send(p *Player, msg OutgoingMessage) {
	websocket.JSON.Send(p.Conn, msg)
}

func sendError(p *Player, errorText string) {
	send(p, OutgoingMessage{
		Type: "error",
		Data: map[string]string{"message": errorText},
	})
}
