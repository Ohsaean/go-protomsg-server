package main

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
)

type MsgHandlerFunc func(user *User, data *gs_protocol.Message)

var msgHandler = map[gs_protocol.Type]MsgHandlerFunc{
	gs_protocol.Type_Login:          LoginHandler,
	gs_protocol.Type_Create:         CreateHandler,
	gs_protocol.Type_Join:           JoinHandler,
	gs_protocol.Type_DefinedAction1: Action1Handler,
	gs_protocol.Type_Quit:           QuitHandler,
	gs_protocol.Type_RoomList:       RoomListHandler,
}

func LoginHandler(user *User, data *gs_protocol.Message) {

	req := data.GetReqLogin()
	if req == nil {
		lib.Log("fail, GetReqLogin()")
	}
	user.userID = req.UserID

	// TODO validation logic here

	// response body marshaling
	res := &gs_protocol.Message{
		Type: gs_protocol.Type_Login,
		ResLogin: &gs_protocol.ResLogin{
			Result: 1,
			UserID: user.userID,
		},
	}

	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.recv <- NewMessage(user.userID, msg)
}

func CreateHandler(user *User, data *gs_protocol.Message) {

	req := data.GetReqCreate()
	if req == nil {
		lib.Log("fail, GetReqCreate()")
	}

	if user.userID != req.UserID {
		if DEBUG {
			lib.Log("Fail room create, user id missmatch")
		}
		return
	}

	// room create
	roomID := GetRandomRoomID()
	r := NewRoom(roomID)
	r.users.Set(user.userID, user) // insert user
	user.room = r                  // set room
	rooms.Set(roomID, r)           // set room into global shared map
	if DEBUG {
		lib.Log("Get rand room id : ", lib.Itoa64(roomID))
	}
	// response body marshaling
	res := &gs_protocol.Message{
		Type: gs_protocol.Type_Create,
		ResCreate: &gs_protocol.ResCreate{
			RoomID: roomID,
			UserID: user.userID,
		},
	}

	if DEBUG {
		lib.Log("Room create, room id : ", lib.Itoa64(roomID))
	}
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
}

func JoinHandler(user *User, data *gs_protocol.Message) {

	// request body unmarshaling
	req := data.GetReqJoin()
	if req == nil {
		lib.Log("fail, GetReqJoin()")
	}

	roomID := req.RoomID

	value, ok := rooms.Get(roomID)

	if !ok {
		if DEBUG {
			lib.Log("Fail room join, room does not exist, room id : ", lib.Itoa64(roomID))
		}
		return
	}

	r := value.(*Room)
	r.users.Set(user.userID, user)
	user.room = r

	// broadcast message
	notifyMsg := &gs_protocol.Message{
		Type: gs_protocol.Type_NotifyJoin,
		NotifyJoin: &gs_protocol.NotifyJoinMsg{
			RoomID: roomID,
			UserID: user.userID,
		},
	}
	msg, err := proto.Marshal(notifyMsg)
	lib.CheckError(err)

	user.SendToAll(NewMessage(user.userID, msg))

	// response body marshaling
	res := &gs_protocol.Message{
		Type: gs_protocol.Type_Join,
		ResJoin: &gs_protocol.ResJoin{
			RoomID: roomID,
			UserID: user.userID,
		},
	}

	msg, err = proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
}

func Action1Handler(user *User, data *gs_protocol.Message) {

	// request body unmarshaling
	req := data.GetReqAction1()
	if req == nil {
		lib.Log("fail, GetReqAction1()")
	}

	// TODO create business logic for Action1 Type

	lib.Log("Action1 userID : ", lib.Itoa64(req.UserID))

	// broadcast message
	notifyMsg := &gs_protocol.Message{
		Type: gs_protocol.Type_NotifyAction1,
		NotifyAction1: &gs_protocol.NotifyAction1Msg{
			UserID: user.userID,
		},
	}
	msg, err := proto.Marshal(notifyMsg)
	lib.CheckError(err)

	user.SendToAll(NewMessage(user.userID, msg))

	// response body marshaling
	res := &gs_protocol.Message{
		Type: gs_protocol.Type_NotifyAction1,
		ResAction1: &gs_protocol.ResAction1{
			UserID: user.userID,
		},
	}

	msg, err = proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
}

func QuitHandler(user *User, data *gs_protocol.Message) {

	// request body unmarshaling
	req := data.GetReqQuit()
	if req == nil {
		lib.Log("fail, GetReqQuit()")
	}

	// response body marshaling
	res := &gs_protocol.Message{
		Type: gs_protocol.Type_Quit,
		ResQuit: &gs_protocol.ResQuit{
			UserID:    user.userID,
			IsSuccess: 1,
		},
	}
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))

	// same act user.Leave()
	user.exit <- struct{}{}
}

func RoomListHandler(user *User, data *gs_protocol.Message) {
	// request body unmarshaling
	req := data.GetReqRoomList()
	if req == nil {
		lib.Log("fail, GetReqQuit()")
	}

	// response body marshaling
	res := &gs_protocol.Message{
		Type: gs_protocol.Type_RoomList,
		ResRoomList: &gs_protocol.ResRoomList{
			UserID:  user.userID,
			RoomIDs: rooms.GetKeys(),
		},
	}
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, msg))
	user.Push(NewMessage(user.userID, msg))
}
