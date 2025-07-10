package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string
	Conn     *websocket.Conn
	Name     string
	PartyID  string
	HP       int
	Move     string
	MoveData map[string]string
	HasMoved bool
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 4

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

func generatePartycode() string {
	code := make([]byte, codeLength)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func generateUniquePartyCode(existing map[string]*Party) string {
	for {
		code := generatePartycode()
		if _, exists := existing[code]; !exists {
			return code
		}
	}
}

func handleCreateParty(player *Player, msg map[string]any) {
	name, ok := msg["name"].(string)
	if !ok {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Missing or invalid name",
		})
		return
	}

	partiesMu.Lock()
	partyID := generateUniquePartyCode(parties)
	party := &Party{ID: partyID}
	parties[partyID] = party
	partiesMu.Unlock()

	player.Name = name
	player.PartyID = partyID
	player.HP = 100
	party.Players[0] = player

	player.Conn.WriteJSON(map[string]any{
		"type":     "party_created",
		"party_id": partyID,
		"message":  fmt.Sprintf("Party %s created.\nWaiting for another player...", partyID),
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
	player.HP = 100
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
	moveData, _ := msg["move_data"].(map[string]string)
	if !ok || (move != "attack" && move != "heal" && move != "hide") {
		player.Conn.WriteJSON(map[string]any{
			"type":  "error",
			"error": "Invalid or missing move",
		})
		return
	}

	player.Move = move
	player.MoveData = moveData
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
	p1.MoveData, p2.MoveData = nil, nil

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
	damage := calculateDamage(p1, p2)
	damage2 := calculateDamage(p2, p1)

	p2.HP -= damage
	p1.HP -= damage2

	status := func(p *Player, other *Player, dmg int, otherDmg int) map[string]any {
		return map[string]any{
			"type": "game_update",
			"you": map[string]any{
				"hp":   p.HP,
				"move": p.Move,
			},
			"enemy": map[string]any{
				"hp":   other.HP,
				"move": other.Move,
			},
			"message": fmt.Sprintf("You took %d damage. Enemy took %d.", otherDmg, dmg),
			"time":    time.Now().Format(time.RFC3339),
		}
	}

	p1.Conn.WriteJSON(status(p1, p2, damage, damage2))
	p2.Conn.WriteJSON(status(p2, p1, damage2, damage))
}

func calculateDamage(attacker *Player, defender *Player) int {
	if attacker.Move == "heal" {
		attacker.HP += 5
		return 0
	}

	if attacker.Move == "hide" {
		return 0 // hide does no damage
	}

	weapon := attacker.MoveData["weapon"]
	hiding := defender.MoveData["cover"]

	// base damage
	var base int
	switch weapon {
	case "revolver":
		base = 15
	case "shotgun":
		base = 20
	case "rifle":
		base = 25
	default:
		base = 10
	}

	// cover modifier
	var mod float64
	switch hiding {
	case "nothing":
		mod = 1.0
	case "barrel":
		switch weapon {
		case "revolver":
			mod = 0.8
		case "shotgun":
			mod = 0.6
		case "rifle":
			mod = 0.4
		}
	case "trough":
		switch weapon {
		case "revolver":
			mod = 0.6
		case "shotgun":
			mod = 0.4
		case "rifle":
			mod = 0.2
		}
	default:
		mod = 1.0
	}

	return int(float64(base) * mod)
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
