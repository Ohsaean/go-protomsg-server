package main

import (
	proto "github.com/golang/protobuf/proto"
	gs "go-protomsg-server/lib"
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
	gs.Log("room create handler start")
	// request body unmarshaling
	req := new(gs_protocol.ReqCreate)
	err := proto.Unmarshal(data, req)
	gs.CheckError(err)

	if user.userID != req.GetUserID() {
		gs.Log("fail room create, error user_id")
		return
	}

	// room create
	roomID := GetRandomRoomID() // 그냥 random 숫자임
	r := NewRoom(roomID)
	r.users.Set(user.userID, user) // 룸의 유저리스트에 유저 추가
	user.room = r     // 유저 객체에 방 할당
	rooms.Set(roomID, r) // 룸 맵에 등록
	gs.Log("get rand room id : ", gs.Itoa64(roomID))

	// response body marshaling
	res := new(gs_protocol.ResCreate)
	res.RoomID = proto.Int64(roomID)
	res.UserID = proto.Int64(user.userID)

	gs.Log("room create success, room id : ", gs.Itoa64(roomID))

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
		gs.Log("room join fail, does not exist, room id : ", gs.Itoa64(roomID))
		return
	}

	r := value.(*Room) // 캐스팅
	r.users.Set(user.userID, user) // 유저 객체 map 에 삽입
	user.room = r               // 유저객체에도 룸 객체 등록

	// 통지(브로드캐스트) 메시지
	notifyMsg := new(gs_protocol.NotifyJoinMsg)
	notifyMsg.UserID = proto.Int64(user.userID)
	notifyMsg.RoomID = proto.Int64(roomID)
	msg, err := proto.Marshal(notifyMsg)
	gs.CheckError(err)

	// 방에 참가 통지
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

	// TODO Action1 에 대한 로직 같은거 처리?
	gs.Log("action1 userID : ", gs.Itoa64(req.GetUserID()))

	// 통지(브로드캐스트) 메시지
	notifyMsg := new(gs_protocol.NotifyAction1Msg)
	notifyMsg.UserID = proto.Int64(user.userID)
	msg, err := proto.Marshal(notifyMsg)
	gs.CheckError(err)

	// 통지
	user.SendToAll(NewMessage(user.userID, gs_protocol.Type_NotifyAction1, msg))

	// response body marshaling
	res := new(gs_protocol.ResAction1)
	res.UserID = proto.Int64(user.userID)
	res.Result = proto.Int32(1) // success?
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
	res.IsSuccess = proto.Int32(1) // success
	msg, err := proto.Marshal(res)
	gs.CheckError(err)
	user.Push(NewMessage(user.userID, gs_protocol.Type_Quit, msg))

	//	user.Leave()
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
