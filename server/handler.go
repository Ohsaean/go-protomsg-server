package main

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
)

// MessageHandler 여기서 각 proto message 에 대한 적절한 프로시저를 할당함
func messageHandler(user *User, msg *gs_protocol.Message) {
	// type switch 말고는 방법이 없나??
	switch msg.Payload.(type) {

	case *gs_protocol.Message_ReqLogin:
		loginHandler(user, msg)

	case *gs_protocol.Message_ReqCreate:
		createHandler(user, msg)

	case *gs_protocol.Message_ReqJoin:
		joinHandler(user, msg)

	case *gs_protocol.Message_ReqAction1:
		action1Handler(user, msg)

	case *gs_protocol.Message_ReqRoomList:
		roomListHandler(user, msg)

	case *gs_protocol.Message_ReqQuit:
		quitHandler(user, msg)

	default:
		lib.Log("Error, not defined handler")
	}
}

func loginHandler(user *User, data *gs_protocol.Message) {

	req := data.GetReqLogin()
	if req == nil {
		lib.Log("fail, GetReqLogin()")
	} else {
		lib.Log("GetReqLogin() : ", req)
	}
	user.userID = req.UserID

	// TODO validation logic here

	// response body marshaling
	res := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ResLogin{
			ResLogin: &gs_protocol.ResLogin{
				UserID: user.userID,
			},
		},
	}

	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.recv <- NewMessage(user.userID, msg)
}

func createHandler(user *User, data *gs_protocol.Message) {

	req := data.GetReqCreate()
	if req == nil {
		lib.Log("fail, GetReqCreate()")
	}

	lib.Log("GetReqCreate() : ", req)

	if user.userID != req.UserID {
		lib.Log("Fail room create, user id missmatch")
		return
	}

	// room create
	roomID := GetAutoIncRoomID()
	r := NewRoom(roomID)
	r.users.Set(user.userID, user) // insert user
	user.room = r                  // set room
	lib.Log("user ", user)
	rooms.Set(roomID, r) // set room into global shared map
	lib.Log("Get rand room id : ", lib.Itoa64(roomID))

	// response body marshaling
	res := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ResCreate{
			ResCreate: &gs_protocol.ResCreate{
				RoomID: roomID,
				UserID: user.userID,
			},
		},
	}

	lib.Log("Room create, room id : ", lib.Itoa64(roomID))

	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
}

func joinHandler(user *User, data *gs_protocol.Message) {

	// request body unmarshaling
	req := data.GetReqJoin()
	if req == nil {
		lib.Log("fail, GetReqJoin()")
	}

	roomID := req.RoomID

	value, ok := rooms.Get(roomID)

	if !ok {

		lib.Log("Fail room join, room does not exist, room id : ", lib.Itoa64(roomID))

		return
	}

	r := value.(*Room)
	r.users.Set(user.userID, user)
	user.room = r

	// broadcast message
	notifyMsg := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqJoin{
			ReqJoin: &gs_protocol.ReqJoin{
				UserID: 1,
				RoomID: roomID,
			},
		},
	}
	msg, err := proto.Marshal(notifyMsg)
	lib.CheckError(err)

	user.SendToAll(NewMessage(user.userID, msg))

	// response body marshaling
	res := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ResJoin{
			ResJoin: &gs_protocol.ResJoin{
				RoomID: roomID,
				UserID: user.userID,
			},
		},
	}

	msg, err = proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
}

func action1Handler(user *User, data *gs_protocol.Message) {

	// request body unmarshaling
	req := data.GetReqAction1()
	if req == nil {
		lib.Log("fail, GetReqAction1()")
	}

	// TODO create business logic for Action1 Type

	lib.Log("Action1 userID : ", lib.Itoa64(req.UserID))

	// broadcast message
	notifyMsg := &gs_protocol.Message{
		Payload: &gs_protocol.Message_NotifyAction1{
			NotifyAction1: &gs_protocol.NotifyAction1Msg{
				UserID: user.userID,
			},
		},
	}
	msg, err := proto.Marshal(notifyMsg)
	lib.CheckError(err)

	user.SendToAll(NewMessage(user.userID, msg))

	// response body marshaling
	res := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ResAction1{
			ResAction1: &gs_protocol.ResAction1{
				UserID: user.userID,
			},
		},
	}

	msg, err = proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
}

func quitHandler(user *User, data *gs_protocol.Message) {

	// request body unmarshaling
	req := data.GetReqQuit()
	if req == nil {
		lib.Log("fail, GetReqQuit()")
	}

	// response body marshaling
	res := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ResQuit{
			ResQuit: &gs_protocol.ResQuit{
				UserID:    user.userID,
				IsSuccess: 1,
			},
		},
	}
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))

	// same act user.Leave()
	user.exit <- true
}

func roomListHandler(user *User, data *gs_protocol.Message) {
	// request body unmarshaling
	req := data.GetReqRoomList()
	if req == nil {
		lib.Log("fail, GetReqQuit()")
	}
	lib.Log("GetReqRoomList() : ", req)

	// response body marshaling
	res := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ResRoomList{
			ResRoomList: &gs_protocol.ResRoomList{
				UserID:  user.userID,
				RoomIDs: rooms.GetKeys(),
			},
		},
	}
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
}
