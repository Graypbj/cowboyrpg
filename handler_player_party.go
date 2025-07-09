package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string
	Conn     *websocket.Conn
	Name     string
	PartyID  string
	HP       int
	Move     string
	HasMoved bool
}

type Party struct {
	ID      string
	Players [2]*Player
}

var (
	playersMu sync.Mutex
	players   = make(map[string]*Player)

	partiesMu sync.Mutex
	parties   = make(map[string]*Party)
)

func handleCreateParty(player *Player, msg map[string]any) {
	name, ok := msg["name"].(string)
	if !ok {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Missing or invalid name",
		})
		return
	}

	partyID := uuid.New().String()
	party := &Party{ID: partyID}
	player.Name = name
	player.PartyID = partyID
	party.Players[0] = player

	partiesMu.Lock()
	parties[partyID] = party
	partiesMu.Unlock()

	player.Conn.WriteJSON(map[string]any{
		"type":     "party_created",
		"party_id": partyID,
		"message":  fmt.Sprintf("Party %s created. Waiting for another player...", partyID),
	})
}

func handleJoinParty(player *Player, msg map[string]any) {
	partyID, ok := msg["party_id"].(string)
	name, nameOk := msg["name"].(string)
	if !ok || !nameOk {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Missing party_id or name",
		})
		return
	}

	partiesMu.Lock()
	party, exists := parties[partyID]
	partiesMu.Unlock()
	if !exists {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Party does not exist",
		})
		return
	}

	if party.Players[1] != nil {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Party is full",
		})
		return
	}

	player.Name = name
	player.PartyID = partyID
	party.Players[1] = player

	player.Conn.WriteJSON(map[string]any{
		"type":     "party_joined",
		"party_id": partyID,
		"message":  fmt.Sprintf("Joined party %s. Game starting...", partyID),
	})

	// Notify both players that the game is starting
	for _, p := range party.Players {
		if p != nil {
			p.Conn.WriteJSON(map[string]any{
				"type":    "game_start",
				"message": "Game has started! Choose your move.",
			})
		}
	}
}

func handleChooseMove(player *Player, msg map[string]any) {
	move, ok := msg["move"].(string)
	if !ok || (move != "attack" && move != "heal" && move != "hide") {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Invalid or missing move",
		})
		return
	}

	player.Move = move
	player.HasMoved = true

	partiesMu.Lock()
	party, exists := parties[player.PartyID]
	partiesMu.Unlock()
	if !exists {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Party no longer exists",
		})
		return
	}

	p1, p2 := party.Players[0], party.Players[1]
	if p1 == nil || p2 == nil || !p1.HasMoved || !p2.HasMoved {
		return // wait until both players have moved
	}

	resolveTurn(p1, p2)

	p1.HasMoved, p2.HasMoved = false, false
	p1.Move, p2.Move = "", ""

	if p1.HP <= 0 || p2.HP <= 0 {
		result := map[string]any{
			"type":   "game_over",
			"result": fmt.Sprintf("%s wins!", winnerName(p1, p2)),
		}
		p1.Conn.WriteJSON(result)
		p2.Conn.WriteJSON(result)
		return
	}

	for _, p := range party.Players {
		if p != nil {
			p.Conn.WriteJSON(map[string]any{
				"type":    "next_turn",
				"message": "Choose your next move",
			})
		}
	}
}

func resolveTurn(p1, p2 *Player) {
	summary := func(p, other *Player) string {
		switch p.Move {
		case "attack":
			if other.Move != "hide" {
				other.HP -= 10
				return fmt.Sprintf("%s attacked %s for 10 damage.", p.Name, other.Name)
			}
			return fmt.Sprintf("%s attacked but %s hid.", p.Name, other.Name)
		case "heal":
			p.HP += 5
			return fmt.Sprintf("%s healed for 5 HP.", p.Name)
		case "hide":
			return fmt.Sprintf("%s hid this turn.", p.Name)
		}
		return ""
	}

	m1 := summary(p1, p2)
	m2 := summary(p2, p1)

	status := func(p *Player, other *Player, msg string) map[string]any {
		return map[string]any{
			"type":    "game_update",
			"you":     map[string]any{"hp": p.HP, "move": p.Move},
			"enemy":   map[string]any{"hp": other.HP, "move": other.Move},
			"message": msg,
			"time":    time.Now().Format(time.RFC3339),
		}
	}

	p1.Conn.WriteJSON(status(p1, p2, m1+" "+m2))
	p2.Conn.WriteJSON(status(p2, p1, m2+" "+m1))
}

func winnerName(p1, p2 *Player) string {
	if p1.HP <= 0 && p2.HP <= 0 {
		return "No one"
	} else if p1.HP <= 0 {
		return p2.Name
	} else {
		return p1.Name
	}
}
