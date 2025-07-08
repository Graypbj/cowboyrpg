package game

import (
	"log"
	"sync"
)

type Party struct {
	ID      string
	Members []*Player
	Battle  *Battle
	mu      sync.Mutex
}

func joinParty(p *Player, partyID, name, class string) {
	partiesMu.Lock()
	party, exists := parties[partyID]
	if !exists {
		party = &Party{ID: partyID}
		parties[partyID] = party
	}
	partiesMu.Unlock()

	p.Name = name
	p.Class = class
	p.PartyID = partyID

	party.mu.Lock()
	defer party.mu.Unlock()

	for _, member := range party.Members {
		if member.ID == p.ID {
			return
		}
	}

	party.Members = append(party.Members, p)
	broadcastToParty(party, OutgoingMessage{
		Type: "player_joined",
		Data: map[string]string{
			"id":    p.ID,
			"name":  p.Name,
			"class": p.Class,
		},
	})

	log.Printf("Player %s joined party %s", p.Name, party.ID)

	if len(party.Members) >= 2 {
		party.StartBattle()
	}
}

func broadcastToParty(party *Party, msg OutgoingMessage) {
	for _, member := range party.Members {
		send(member, msg)
	}
}
