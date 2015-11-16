package main

import (
	proto "github.com/golang/protobuf/proto"
	gs "github.com/Ohsaean/go-protomsg-server/lib"
	"go-protomsg-server/protobuf"
)

type MsgHandlerFunc func(user *User, data []byte)

var msgHandlerMapping = map[gs_protocol.Type]MsgHandlerFunc{
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
	gs.CheckError(err)
	user.userID = req.GetUserID()

	// TODO validation logic here

	// response body marshaling
	res := new(gs_protocol.ResLogin)
	res.Result = proto.Int32(1)
	res.UserID = proto.Int64(user.userID)

	msg, err := proto.Marshal(res)
	gs.CheckError(err)
	user.recv <- NewMessage(user.userID, gs_protocol.Type_Login, msg)
}

func CreateHandler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqCreate)
	err := proto.Unmarshal(data, req)
	gs.CheckError(err)

	if user.userID != req.GetUserID() {
		gs.Log("Fail room create, user id missmatch")
		return
	}

	// room create
	roomID := GetRandomRoomID()
	r := NewRoom(roomID)
	r.users.Set(user.userID, user) // insert user
	user.room = r                  // set room
	rooms.Set(roomID, r)           // set room into global shared map
	gs.Log("Get rand room id : ", gs.Itoa64(roomID))

	// response body marshaling
	res := new(gs_protocol.ResCreate)
	res.RoomID = proto.Int64(roomID)
	res.UserID = proto.Int64(user.userID)

	gs.Log("Room create, room id : ", gs.Itoa64(roomID))

	msg, err := proto.Marshal(res)
	gs.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_Create, msg))
}

func JoinHandler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqJoin)
	err := proto.Unmarshal(data, req)
	gs.CheckError(err)

	roomID := req.GetRoomID()

	value, ok := rooms.Get(roomID)

	if !ok {
		gs.Log("Fail room join, room does not exist, room id : ", gs.Itoa64(roomID))
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
	gs.CheckError(err)

	user.SendToAll(NewMessage(user.userID, gs_protocol.Type_NotifyJoin, msg))

	// response body marshaling
	res := new(gs_protocol.ResJoin)
	res.UserID = proto.Int64(user.userID)
	res.RoomID = proto.Int64(roomID)
	res.Members = r.getRoomUsers()

	msg, err = proto.Marshal(res)
	gs.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_Join, msg))
}

func Action1Handler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqAction1)
	err := proto.Unmarshal(data, req)
	gs.CheckError(err)

	// TODO create business logic for Action1 Type
	gs.Log("Action1 userID : ", gs.Itoa64(req.GetUserID()))

	// broadcast message
	notifyMsg := new(gs_protocol.NotifyAction1Msg)
	notifyMsg.UserID = proto.Int64(user.userID)
	msg, err := proto.Marshal(notifyMsg)
	gs.CheckError(err)

	user.SendToAll(NewMessage(user.userID, gs_protocol.Type_NotifyAction1, msg))

	// response body marshaling
	res := new(gs_protocol.ResAction1)
	res.UserID = proto.Int64(user.userID)
	res.Result = proto.Int32(1) // is success?
	msg, err = proto.Marshal(res)
	gs.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_DefinedAction1, msg))
}

func QuitHandler(user *User, data []byte) {

	// request body unmarshaling
	req := new(gs_protocol.ReqQuit)
	err := proto.Unmarshal(data, req)
	gs.CheckError(err)

	res := new(gs_protocol.ResQuit)
	res.IsSuccess = proto.Int32(1) // is success?
	msg, err := proto.Marshal(res)
	gs.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_Quit, msg))

	// same act user.Leave()
	user.exit <- struct{}{}
}

func RoomListHandler(user *User, data []byte) {
	// request body unmarshaling
	req := new(gs_protocol.ReqRoomList)
	err := proto.Unmarshal(data, req)
	gs.CheckError(err)

	res := new(gs_protocol.ResRoomList)
	res.RoomIDs = rooms.GetKeys()
	msg, err := proto.Marshal(res)
	gs.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_RoomList, msg))
}
