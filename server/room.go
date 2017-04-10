package main

import (
	"github.com/ohsaean/gogpd/lib"
	"math"
)

type Room struct {
	users    lib.SharedMap
	messages chan *Message // broadcast message channel
	roomID   int64
}

func GetRoom(roomID int64) (r *Room) {
	value, ok := rooms.Get(roomID)

	if !ok {
		if DEBUG {
			lib.Log("err: not exist room : ", roomID)
		}
	}
	r = value.(*Room)
	return
}

func GetRandomRoomID() (uuid int64) {
	for {
		// TODO change room-id generate strategy
		uuid = int64(lib.RandInt32(1, math.MaxInt32))
		if _, ok := rooms.Get(uuid); ok {
			if DEBUG {
				lib.Log("err: exist same room id")
			}
			continue
		}
		return
	}
}

// Room construct
func NewRoom(roomID int64) (r *Room) {
	r = new(Room)
	r.messages = make(chan *Message)
	r.users = lib.NewSMap(lib.Channel) // global shared map
	r.roomID = roomID
	go r.RoomMessageLoop() // for broadcast message
	return
}

func (r *Room) RoomMessageLoop() {
	// when messages channel is closed then "for-loop" will be break
	for m := range r.messages {
		for userID, _ := range r.users.Map() {
			if userID != m.userID {
				value, ok := r.users.Get(userID)
				if ok {
					user := value.(*User)
					if DEBUG {
						lib.Log("Push message for broadcast :" + lib.Itoa64(user.userID))
					}
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
		close(r.messages) // close broadcast channel
		rooms.Remove(r.roomID)
	}
}

func (r *Room) IsEmptyRoom() bool {
	if r.users.Count() == 0 {
		return true
	}
	return false
}
