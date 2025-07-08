package game

import "log"

type Battle struct {
	Party     *Party
	TurnIndex int
	Started   bool
}

func (p *Party) StartBattle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Battle != nil && p.Battle.Started {
		return
	}

	p.Battle = &Battle{
		Party:     p,
		TurnIndex: 0,
		Started:   true,
	}

	log.Printf("Battle started in party %s", p.ID)

	broadcastToParty(p, OutgoingMessage{
		Type: "battle_start",
		Data: map[string]string{
			"message": "Battle has begun",
		},
	})

	p.Battle.announceTurn()
}

func (b *Battle) announceTurn() {
	current := b.Party.Members[b.TurnIndex]
	broadcastToParty(b.Party, OutgoingMessage{
		Type: "turn_start",
		Data: map[string]string{
			"playerId": current.ID,
			"name":     current.Name,
		},
	})
}

func (b *Battle) HandleAction(from *Player, actionType string) {
	if b.Party.Members[b.TurnIndex].ID != from.ID {
		sendError(from, "Not your turn")
		return
	}

	broadcastToParty(b.Party, OutgoingMessage{
		Type: "action_performed",
		Data: map[string]string{
			"playerId":   from.ID,
			"actionType": actionType,
		},
	})

	b.TurnIndex = (b.TurnIndex + 1) % len(b.Party.Members)
	b.announceTurn()
}
