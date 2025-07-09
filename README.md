# 🐎 Cowboy RPG - CLI Client

This is the official Go CLI client for the **Cowboy RPG** WebSocket-based multiplayer game. Players can create or join a party and engage in turn-based cowboy combat using a revolver, shotgun, or rifle. Take cover behind environmental objects and plan your actions wisely to outsmart your opponent!

---

## 📦 Requirements

* Go 1.20+ installed ([Install Go](https://golang.org/doc/install))
* Terminal emulator that supports ANSI escape sequences (e.g., Linux terminal, macOS Terminal, Windows Terminal or WSL)
* Active Cowboy RPG server running (can be hosted locally or remotely on a device like a Raspberry Pi)
* Internet access (if connecting to a remote server)

---

## 🧪 Features

* ✅ Create or join a 2-player party
* ✅ Real-time WebSocket communication with the server
* ✅ Select moves using an interactive CLI (`promptui`)
* ✅ Choose weapon and cover types for attacks
* ✅ Turn-based combat system with synchronized state updates
* ✅ Clean screen refreshes and combat summaries every round

---

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/Graypbj/cowboyrpg.git
cd cowboy-rpg/client
```

### 2. Install Dependencies

This project uses two external libraries:

```bash
go get github.com/gorilla/websocket
go get github.com/manifoldco/promptui
```

---

## 🕹️ How to Play

### Step 1: Run the Client

```bash
go run .
```

### Step 2: Choose an Option

```
Welcome to Cowboy RPG
1. Create Party
2. Join Party
Choose an option: 
```

* **Create Party**: You'll get a party ID to share with a friend.
* **Join Party**: Enter a valid party ID to join your friend.

### Step 3: Combat Begins!

Once both players have joined, the game will start:

* You'll be shown your HP and your opponent's HP after each round.
* Choose your move:

  * **Attack**: Choose a weapon (revolver, shotgun, rifle) and cover (nothing, barrel, trough).
  * **Heal**: Recover some HP.
  * **Hide**: Take cover to reduce incoming damage.

> Your move is locked once submitted. The game waits for both players to respond before resolving and displaying the results.

---

## 🧼 Troubleshooting

* **Nothing happens after joining**: Wait for another player to join your party.
* **Screen doesn't clear**: Try running in a compatible terminal emulator (WSL/Linux/macOS/Windows Terminal).
* **Can't connect**: Make sure your server is running and reachable at the correct address and port.

---

## 🌐 Hosting the Server

Want to play with friends on different networks?

* Host the Cowboy RPG server on a VPS or Raspberry Pi.
* Use port forwarding to expose `:8080`.
* Ensure the `serverURL` in the client points to your public IP:

```go
const serverURL = "ws://<your-public-ip>:8080/ws"
```

---

## 🛠️ Contributing

Contributions are welcome! Submit issues or pull requests to help improve the Cowboy RPG experience.

---

## 📄 License

MIT License. See [`LICENSE`](../LICENSE) for more information.

