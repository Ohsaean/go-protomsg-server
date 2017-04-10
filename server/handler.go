package main

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
)

type MsgHandlerFunc func(user *User, data []byte)

var msgHandler = map[gs_protocol.Type]MsgHandlerFunc{
	gs_protocol.Type_Login:          LoginHandler,
	gs_protocol.Type_Create:         CreateHandler,
	gs_protocol.Type_Join:           JoinHandler,
	gs_protocol.Type_DefinedAction1: Action1Handler,
	gs_protocol.Type_Quit:           QuitHandler,
	gs_protocol.Type_RoomList:       RoomListHandler,
}

func LoginHandler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqLogin)
	err := proto.Unmarshal(data, req)
	lib.CheckError(err)
	user.userID = req.GetUserID()

	// TODO validation logic here

	// response body marshaling
	res := new(gs_protocol.ResLogin)
	res.Result = proto.Int32(1)
	res.UserID = proto.Int64(user.userID)

	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.recv <- NewMessage(user.userID, gs_protocol.Type_Login, msg)
}

func CreateHandler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqCreate)
	err := proto.Unmarshal(data, req)
	lib.CheckError(err)

	if user.userID != req.GetUserID() {
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
	res := new(gs_protocol.ResCreate)
	res.RoomID = proto.Int64(roomID)
	res.UserID = proto.Int64(user.userID)

	if DEBUG {
		lib.Log("Room create, room id : ", lib.Itoa64(roomID))
	}
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_Create, msg))
}

func JoinHandler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqJoin)
	err := proto.Unmarshal(data, req)
	lib.CheckError(err)

	roomID := req.GetRoomID()

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
	notifyMsg := new(gs_protocol.NotifyJoinMsg)
	notifyMsg.UserID = proto.Int64(user.userID)
	notifyMsg.RoomID = proto.Int64(roomID)
	msg, err := proto.Marshal(notifyMsg)
	lib.CheckError(err)

	user.SendToAll(NewMessage(user.userID, gs_protocol.Type_NotifyJoin, msg))

	// response body marshaling
	res := new(gs_protocol.ResJoin)
	res.UserID = proto.Int64(user.userID)
	res.RoomID = proto.Int64(roomID)
	res.Members = r.getRoomUsers()

	msg, err = proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_Join, msg))
}

func Action1Handler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqAction1)
	err := proto.Unmarshal(data, req)
	lib.CheckError(err)

	// TODO create business logic for Action1 Type
	if DEBUG {
		lib.Log("Action1 userID : ", lib.Itoa64(req.GetUserID()))
	}

	// broadcast message
	notifyMsg := new(gs_protocol.NotifyAction1Msg)
	notifyMsg.UserID = proto.Int64(user.userID)
	msg, err := proto.Marshal(notifyMsg)
	lib.CheckError(err)

	user.SendToAll(NewMessage(user.userID, gs_protocol.Type_NotifyAction1, msg))

	// response body marshaling
	res := new(gs_protocol.ResAction1)
	res.UserID = proto.Int64(user.userID)
	res.Result = proto.Int32(1) // is success?
	msg, err = proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_DefinedAction1, msg))
}

func QuitHandler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqQuit)
	err := proto.Unmarshal(data, req)
	lib.CheckError(err)

	res := new(gs_protocol.ResQuit)
	res.IsSuccess = proto.Int32(1) // is success?
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_Quit, msg))

	// same act user.Leave()
	user.exit <- struct{}{}
}

func RoomListHandler(user *User, data []byte) {
	// request body unmarshaling
	req := new(gs_protocol.ReqRoomList)
	err := proto.Unmarshal(data, req)
	lib.CheckError(err)

	res := new(gs_protocol.ResRoomList)
	res.RoomIDs = rooms.GetKeys()
	msg, err := proto.Marshal(res)
	lib.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_RoomList, msg))
}