package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
	"net"
)

type MsgHandlerFunc func(data []byte)

var msgHandlerMapping = map[gs_protocol.Type]MsgHandlerFunc{
	gs_protocol.Type_Login:          ResLogin,
	gs_protocol.Type_Create:         ResCreate,
	gs_protocol.Type_Join:           ResJoin,
	gs_protocol.Type_DefinedAction1: ResAction1,
	gs_protocol.Type_Quit:           ResQuit,
	gs_protocol.Type_NotifyJoin:     NotifyJoinHandler,
	gs_protocol.Type_NotifyAction1:  NotifyAction1Handler,
	gs_protocol.Type_NotifyQuit:     NotifyQuitHandler,
	gs_protocol.Type_RoomList:       ResRoomList,
}

//var buffer bytes.Buffer
//for i := 0; i < b.N; i++ {
//buffer.WriteString(s2)
//}
//s1 := buffer.String()

var user_buffer bytes.Buffer
var recv_buffer bytes.Buffer

func main() {
	client, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	data := make([]byte, 4096)
	exit := make(chan struct{})

	var userID int64
	var method int
	AddUserBuffer("=================================================================\n")
	AddUserBuffer(" Input user ID (it must be a whole number greater than 0\n")
	AddUserBuffer("=================================================================\n")
	AddUserBuffer("userID : ")
	fmt.Scanln(&userID)

	ReqLogin(client, userID, data)

	go func() {
		for {
			AddUserBuffer("=================================================================\n")
			AddUserBuffer(" Input command number (1~5)\n")
			AddUserBuffer("=================================================================\n")
			AddUserBuffer("1. room list\n")
			AddUserBuffer("2. room create\n")
			AddUserBuffer("3. join\n")
			AddUserBuffer("4. action1\n")
			AddUserBuffer("5. quit\n")
			AddUserBuffer("choose number: ")
			fmt.Scanln(&method)

			switch method {
			case 1:
				ReqRoomList(client, userID, data)
			case 2:
				ReqCreate(client, userID, data)
			case 3:
				var roomID int64
				fmt.Print("input room id : ")
				fmt.Scanln(&roomID)
				ReqJoin(client, userID, data, roomID)
			case 4:
				ReqAction1(client, userID, data)
			case 5:
				ReqQuit(client, userID, data)
				fmt.Println("program exit..bye")
				exit <- struct{}{}
				return
			default:
				continue
			}
		}
	}()

	go func() {
		data := make([]byte, 4096)

		for {
			n, err := client.Read(data)
			if err != nil {
				lib.Log("Fail Stream read, err : ", err)
				break
			}

			messageType := gs_protocol.Type(lib.ReadInt32(data[0:4]))
			lib.Log("Decoding type : ", messageType)

			rawData := data[4:n]
			handler, ok := msgHandlerMapping[messageType]

			if ok {
				handler(rawData)
			} else {
				lib.Log("Fail no function defined for type", handler)
				break
			}
		}
	}()

	<-exit
}

func AddUserBuffer(str string) {
	recv_buffer.WriteString(str)
}

func AddUserBufferJson(str string, v interface{}) {
	clientSend, err := json.Marshal(v)
	lib.CheckError(err)
	user_buffer.WriteString(str + string(clientSend))
}

func AddRecvBuffer(str string) {
	recv_buffer.WriteString(str)
}

func ReqLogin(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqLogin)
	req.UserID = proto.Int64(userUID)

	msgTypeBytes := lib.WriteMsgType(gs_protocol.Type_Login)
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	AddUserBufferJson("client send ", req)
}

func ResLogin(rawData []byte) {

	res := new(gs_protocol.ResLogin)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)

	AddRecvBuffer("ResLogin server return : user id " + res.GetUserID())
	AddRecvBuffer("ResLogin server return : result id " + res.GetResult())
}

func ReqRoomList(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqRoomList)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := lib.WriteMsgType(gs_protocol.Type_RoomList)
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	AddUserBufferJson("client send ", req)
}

func ResRoomList(rawData []byte) {

	res := new(gs_protocol.ResRoomList)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)

	AddRecvBuffer("ResRoomList server return : members " + lib.Int64SliceToString(res.GetRoomIDs()))
}

func ReqCreate(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqCreate)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := lib.WriteMsgType(gs_protocol.Type_Create)
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	AddUserBufferJson("client send ", req)
}

func ResCreate(rawData []byte) {

	res := new(gs_protocol.ResCreate)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)

	AddRecvBuffer("ResCreate server return : user id " + res.GetUserID())
	AddRecvBuffer("ResCreate server return : room id " + res.GetRoomID())
}

func ReqJoin(c net.Conn, userUID int64, data []byte, roomID int64) {
	req := new(gs_protocol.ReqJoin)
	req.UserID = proto.Int64(userUID)
	req.RoomID = proto.Int64(roomID)
	msgTypeBytes := lib.WriteMsgType(gs_protocol.Type_Join)
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	AddUserBufferJson("client send ", req)
}

func ResJoin(rawData []byte) {

	res := new(gs_protocol.ResJoin)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)

	AddRecvBuffer("ResJoin server return : user id " + res.GetUserID())
	AddRecvBuffer("ResJoin server return : room id " + res.GetRoomID())
	AddRecvBuffer("ResJoin server return : members " + lib.Int64SliceToString(res.GetMembers()))
}

func ReqAction1(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqAction1)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := lib.WriteMsgType(gs_protocol.Type_DefinedAction1)
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	AddUserBufferJson("client send ", req)
}

func ResAction1(rawData []byte) {

	res := new(gs_protocol.ResAction1)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)

	AddRecvBuffer("ResAction1 server return : user id " + res.GetUserID())
	AddRecvBuffer("ResAction1 server return : result " + res.GetResult())
}

func ReqQuit(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqQuit)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := lib.WriteMsgType(gs_protocol.Type_Quit)
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	AddUserBufferJson("client send ", req)
}

func ResQuit(rawData []byte) {
	res := new(gs_protocol.ResQuit)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)
	AddRecvBuffer("ResQuit server return : is success? " + res.GetIsSuccess())
}

func NotifyJoinHandler(rawData []byte) {
	res := new(gs_protocol.NotifyJoinMsg)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)

	AddRecvBuffer("NotifyJoin server return : user id " + res.GetUserID())
	AddRecvBuffer("NotifyJoin server return : room id " + res.GetRoomID())
}

func NotifyAction1Handler(rawData []byte) {
	res := new(gs_protocol.NotifyAction1Msg)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)
	AddRecvBuffer("NotifyAction1 server return : user id " + res.GetUserID())
}

func NotifyQuitHandler(rawData []byte) {
	res := new(gs_protocol.NotifyQuitMsg)
	err := proto.Unmarshal(rawData, res)
	lib.CheckError(err)
	AddRecvBuffer("NotifyQuit server return : user id " + res.GetUserID())
}
