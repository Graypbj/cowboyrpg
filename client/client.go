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

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to Cowboy RPG")
	fmt.Println("1. Create Party")
	fmt.Println("2. Join Party")
	fmt.Print("Choose an option: ")
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
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("Disconnected from server")
			return
		}

		var msg map[string]any
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Println("Failed to parse message:", err)
			continue
		}

		switch msg["type"] {
		case "party_created":
			fmt.Printf("Party Created: %s\n", msg["party_id"])
		case "party_joined":
			fmt.Println(msg["message"])
		case "game_start", "next_turn":
			fmt.Println(msg["message"])

			prompt := promptui.Select{
				Label: "Choose your move",
				Items: []string{"attack", "heal", "hide"},
			}

			_, result, err := prompt.Run()
			if err != nil {
				log.Println("Prompt failed:", err)
				return
			}

			conn.WriteJSON(map[string]any{
				"type": "choose_move",
				"move": result,
			})
		case "game_update":
			fmt.Printf("You: HP=%v, Move=%v\n", msg["you"].(map[string]any)["hp"], msg["you"].(map[string]any)["move"])
			fmt.Printf("Enemy: HP=%v, Move=%v\n", msg["enemy"].(map[string]any)["hp"], msg["enemy"].(map[string]any)["move"])
			fmt.Println(msg["message"])
		case "game_over":
			fmt.Println("Game Over:", msg["result"])
			return
		case "error":
			fmt.Println("Error:", msg["error"])
		default:
			fmt.Println("Unknown message:", msg)
		}
	}
}
