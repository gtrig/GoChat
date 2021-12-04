package server

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Participant struct {
	Host bool
	Conn *websocket.Conn
}

type RoomMap struct {
	Mutex sync.RWMutex
	Map   map[string][]Participant
}

func (r *RoomMap) Init() {
	r.Map = make(map[string][]Participant)
}

func (r *RoomMap) Get(roomID string) []Participant {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	return r.Map[roomID]
}

// CreateRoom creates a new room with the random unique roomID
func (r *RoomMap) CreateRoom() string {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	roomID := uuid.NewString()
	r.Map[roomID] = make([]Participant, 0)
	return roomID
}

// JoinRoom joins the given roomID with a boolean flag for host and a connection
func (r *RoomMap) JoinRoom(roomID string, host bool, conn *websocket.Conn) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.Map[roomID] = append(r.Map[roomID], Participant{host, conn})
}

// DeleteRoom deletes the room with the given roomID
func (r *RoomMap) DeleteRoom(roomID string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	delete(r.Map, roomID)
}
