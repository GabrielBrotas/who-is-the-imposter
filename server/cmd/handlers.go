package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Consider validating against a list of approved origins
		},
	}
	gameManager = NewGameManager()
	words       = []string{"uva", "banana", "abacate", "abobora", "acerola"}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	player := NewPlayer(conn)

	gameManager.AddPlayer(player)
	gameManager.SendPlayerList(conn)
	gameManager.BroadcastPlayerList()

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading json: %v", err)
			break
		}

		// Handle different message types here
		switch msg["type"].(string) {
		case "changeName":
			newName, ok := msg["newName"].(string)
			if ok {
				gameManager.ChangePlayerName(player, newName)
				gameManager.BroadcastPlayerList()
			}
		case "startGame":
			dist := ImposterDistribution{
				One:   int(msg["one"].(float64)), // Convert float64 to int (common in JSON parsing)
				Two:   int(msg["two"].(float64)),
				Three: int(msg["three"].(float64)),
			}
			gameManager.StartGame(dist)
		case "resetGame":
			gameManager.ResetGame()
		}
	}

	gameManager.RemovePlayer(player)
	gameManager.BroadcastPlayerList()
}
