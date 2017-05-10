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
	notifyMsg := new(gs_protocol.NotifyQuitMsg)
	if DEBUG {
		lib.Log("Leave user id : ", lib.Itoa64(u.userID))
	}
	notifyMsg.UserID = proto.Int64(u.userID)

	if u.room != nil {
		if DEBUG {
			lib.Log("Leave room id : ", lib.Itoa64(u.room.roomID))
		}
		notifyMsg.RoomID = proto.Int64(u.room.roomID)

		msg, err := proto.Marshal(notifyMsg)
		lib.CheckError(err)

		// race condition by broadcast goroutine and onClientWrite goroutine
		u.room.Leave(u.userID)

		// notify all members in the room
		u.SendToAll(NewMessage(u.userID, gs_protocol.Type_NotifyQuit, msg))
		if DEBUG {
			lib.Log("NotifyQuit message send")
		}
	}

	if DEBUG {
		lib.Log("Leave func end")
	}
}

func (u *User) Push(m *Message) {
	u.recv <- m // send message to user
}

func (u *User) SendToAll(m *Message) {
	if u.room.IsEmptyRoom() == false {
		u.room.messages <- m
	}
}
