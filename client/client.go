package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/manifoldco/promptui"
)

const serverURL = "wss://50f94bcdc1ec.ngrok-free.app/ws"

var playerName string
var hasMadeMove bool

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Dial error: %v", err)
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
	playerName = strings.TrimSpace(name)

	if choice == "1" {
		conn.WriteJSON(map[string]any{
			"type": "create_party",
			"name": playerName,
		})
	} else {
		fmt.Print("Enter Party ID to join: ")
		partyID, _ := reader.ReadString('\n')
		partyID = strings.TrimSpace(partyID)
		conn.WriteJSON(map[string]any{
			"type":     "join_party",
			"party_id": partyID,
			"name":     playerName,
		})
	}

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg["type"] {
		case "party_created", "party_joined":
			fmt.Println(msg["message"])
		case "game_update":
			clearScreen()
			fmt.Println(msg["message"])

			if you, ok := msg["you"].(map[string]any); ok {
				fmt.Printf("%s HP: %.0f\n", playerName, you["hp"].(float64))
			}
			if enemy, ok := msg["enemy"].(map[string]any); ok {
				enemyMove := enemy["move"]
				enemyHP := enemy["hp"].(float64)
				fmt.Printf("Enemy Move: %s\nEnemy HP: %.0f\n", enemyMove, enemyHP)
			}
			fmt.Println("Choose your next move")
			hasMadeMove = false
		case "game_start", "next_turn":
			fmt.Println(msg["message"])
			if !hasMadeMove {
				selectMove(conn)
				hasMadeMove = true
			}
		case "game_over":
			clearScreen()
			fmt.Println("Game Over:", msg["result"])
			return
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

	if move == "attack" {
		weaponPrompt := promptui.Select{
			Label: "Choose weapon",
			Items: []string{"revolver", "shotgun", "rifle"},
		}
		_, weapon, _ := weaponPrompt.Run()
		moveData["weapon"] = weapon

		coverPrompt := promptui.Select{
			Label: "Choose where to hide",
			Items: []string{"nothing", "barrel", "trough"},
		}
		_, cover, _ := coverPrompt.Run()
		moveData["cover"] = cover
	} else if move == "hide" {
		coverPrompt := promptui.Select{
			Label: "Choose where to hide",
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

func clearScreen() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
