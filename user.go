package main

import (
	"github.com/golang/protobuf/proto"
	"go-protomsg-server/lib"
	"go-protomsg-server/protobuf"
)

type User struct {
	userID int64         // 아이디
	room   *Room         // 방 객체
	recv   chan *Message // 이벤트 수신용 채널
	exit   chan struct{} // 나가기용도
}

// 생성자
func NewUser(uid int64, room *Room) *User {
	return &User{
		userID: uid,
		recv:   make(chan *Message),
		exit:   make(chan struct{}),
		room:   room,
	}
}

func (u *User) Leave() {
	// 통지(브로드캐스트) 메시지
	notifyMsg := new(gs_protocol.NotifyQuitMsg)
	gsutil.Log("user id : ", gsutil.Itoa64(u.userID))
	notifyMsg.UserID = proto.Int64(u.userID)

	if u.room != nil {
		gsutil.Log("room id : ", gsutil.Itoa64(u.room.roomID))
		notifyMsg.RoomID = proto.Int64(u.room.roomID)

		msg, err := proto.Marshal(notifyMsg)
		gsutil.CheckError(err)

		// 탈퇴 처리 (broadcast gorutine 과 race condition 발생함)
		 u.room.Leave(u.userID)
		 gsutil.Log("Leave proc end")

		// 방에 탈퇴 통지
		u.SendToAll(NewMessage(u.userID, gs_protocol.Type_NotifyQuit, msg))
		gsutil.Log("SendToAll Leave end")
	}
}

func (u *User) Push(m *Message) {
	u.recv <- m // 유저에게 이벤트 보내기
}

func (u *User) SendToAll(m *Message) {
	if u.room.users.Count() > 0 { // 방에 유저가 있을 경우에만..
		u.room.messages <- m
	}
}
