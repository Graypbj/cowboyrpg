package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/manifoldco/promptui"
)

const serverURL = "ws://cowboyrpg.duckdns.org:8080/ws"

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Connection error: %v", err)
	}
	defer conn.Close()

	fmt.Println("Welcome to Cowboy RPG")
	fmt.Println("1. Create Party")
	fmt.Println("2. Join Party")
	fmt.Print("Choose an option: ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if choice == "1" {
		conn.WriteJSON(map[string]any{
			"type": "create_party",
			"name": name,
		})
	} else if choice == "2" {
		fmt.Print("Enter Party ID to join: ")
		partyID, _ := reader.ReadString('\n')
		partyID = strings.TrimSpace(partyID)
		conn.WriteJSON(map[string]any{
			"type":     "join_party",
			"party_id": partyID,
			"name":     name,
		})
	} else {
		fmt.Println("Invalid option")
		return
	}

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg["type"] {
		case "party_created", "party_joined", "game_update", "game_over":
			fmt.Println(msg["message"])
			if msg["type"] == "game_over" {
				fmt.Println("Game Over:", msg["result"])
				return
			}
		case "game_start", "next_turn":
			fmt.Println(msg["message"])
			selectMove(conn)
		case "error":
			fmt.Println("Error:", msg["error"])
		}
	}
}

func selectMove(conn *websocket.Conn) {
	prompt := promptui.Select{
		Label: "Choose your action",
		Items: []string{"attack", "heal", "hide"},
	}

	_, move, err := prompt.Run()
	if err != nil {
		log.Println("Prompt failed:", err)
		return
	}

	moveData := map[string]string{}

	switch move {
	case "attack":
		weaponPrompt := promptui.Select{
			Label: "Choose weapon",
			Items: []string{"revolver", "shotgun", "rifle"},
		}
		_, weapon, _ := weaponPrompt.Run()
		moveData["weapon"] = weapon
	case "hide":
		coverPrompt := promptui.Select{
			Label: "Choose cover",
			Items: []string{"nothing", "barrel", "trough"},
		}
		_, cover, _ := coverPrompt.Run()
		moveData["cover"] = cover
	}

	err = conn.WriteJSON(map[string]any{
		"type":      "choose_move",
		"move":      move,
		"move_data": moveData,
	})
	if err != nil {
		log.Println("Write error:", err)
	}
}
