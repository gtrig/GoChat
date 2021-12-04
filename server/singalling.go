package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var AllRooms RoomMap

// CreateRoomRequestHandler Create a Room and return roomID
func CreateRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	roomID := AllRooms.CreateRoom()

	type response struct {
		RoomID string `json:"roomID"`
	}

	json.NewEncoder(w).Encode(response{RoomID: roomID})
}

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type broadcastMsg struct {
	Msg    map[string]interface{} `json:"msg"`
	RoomID string                 `json:"roomID"`
	Client *websocket.Conn        `json:"-"`
}

var broadcast = make(chan broadcastMsg)

func broadcaster() {
	for {
		msg := <-broadcast
		room := AllRooms.Get(msg.RoomID)
		if room == nil {
			continue
		}

		for _, client := range AllRooms.Map[msg.RoomID] {
			if client.Conn == msg.Client {
				continue
			}

			err := client.Conn.WriteJSON(msg.Msg)
			if err != nil {
				log.Println("Error sending message to client", err)
				client.Conn.Close()
			}
		}
	}
}

// JoinRoomRequestHandler Join a Room
func JoinRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")

	room := AllRooms.Get(roomID)
	if room == nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("WebSocket upgrade error:", err)
	}

	AllRooms.JoinRoom(roomID, false, ws)

	go broadcaster()

	for {
		var msg broadcastMsg
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading message from client", err)
			ws.Close()
			break
		}

		msg.Client = ws
		msg.RoomID = roomID
		broadcast <- msg
	}
}
