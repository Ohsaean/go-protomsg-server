package main

import (
	gs "github.com/Ohsaean/go-protomsg-server/lib"
	"math"
)

type Room struct {
	users    gs.SharedMap
	messages chan *Message   // broadcast message channel
	roomID   int64
}

func GetRoom(roomID int64) (r *Room) {
	value, ok := rooms.Get(roomID)

	if !ok {
		gs.Log("err: not exist room : ", roomID)
	}
	r = value.(*Room)
	return
}

func GetRandomRoomID() (uuid int64) {
	for {
		// TODO change room-id generate strategy
		uuid = int64(gs.RandInt32(1, math.MaxInt32))
		if _, ok := rooms.Get(uuid); ok {
			gs.Log("err: exist same room id")
			continue
		}
		return
	}
}

// Room construct
func NewRoom(roomID int64) (r *Room) {
	r = new(Room)
	r.messages = make(chan *Message)
	r.users = gs.NewSMap() // global shared map
	r.roomID = roomID
	go r.RoomMessageLoop() // for broadcast message
	return
}

func (r *Room) RoomMessageLoop() {
	// when messages channel is closed then "for-loop" will be break
	for m := range r.messages {
		for userID, _  := range r.users.Map() {
			if userID != m.userID {
				value, ok := r.users.Get(userID)
				if ok {
					user := value.(*User)
					gs.Log("Push message for broadcast :" + gs.Itoa64(user.userID))
					user.Push(m)
				}
			}
		}
	}
}

func (r *Room) getRoomUsers() (userIDs []int64) {
	userIDs = r.users.GetKeys()
	return userIDs
}

func (r *Room) Leave(userID int64) {
	r.users.Remove(userID)

	if r.IsEmptyRoom() {
		close(r.messages)      // close broadcast channel
		rooms.Remove(r.roomID)
	}
}

func (r *Room) IsEmptyRoom() bool {
	if r.users.Count() == 0 {
		return true
	}
	return false
}
