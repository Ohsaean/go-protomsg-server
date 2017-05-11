package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
)

type User struct {
	userID int64
	room   *Room
	recv   chan *Message
	exit   chan struct{} // signal
}

func NewUser(uid int64, room *Room) *User {
	return &User{
		userID: uid,
		recv:   make(chan *Message),
		exit:   make(chan struct{}),
		room:   room,
	}
}

func (u *User) Leave() {

	lib.Log("Leave user id : ", lib.Itoa64(u.userID))

	if u.room == nil {
		lib.Log("Error, room is nil")
		return
	}

	lib.Log("Leave room id : ", lib.Itoa64(u.room.roomID))

	// broadcast message
	notifyMsg := &gs_protocol.Message{
		Type: gs_protocol.Type_NotifyQuit,
		NotifyQuit: &gs_protocol.NotifyQuitMsg{
			UserID: u.userID,
			RoomID: u.room.roomID,
		},
	}
	msg, err := proto.Marshal(notifyMsg)
	lib.CheckError(err)
	u.SendToAll(NewMessage(u.userID, msg))

	lib.Log("NotifyQuit message send")

	lib.Log("Leave func end")
}

func (u *User) Push(m *Message) {
	u.recv <- m // send message to user
}

func (u *User) SendToAll(m *Message) {
	if u.room.IsEmptyRoom() == false {
		u.room.messages <- m
	}
}
