package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	gs "github.com/ohsaean/gogpd/lib"
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
	fmt.Println("=================================================================")
	fmt.Println(" Input user ID (it must be a whole number greater than 0")
	fmt.Println("=================================================================")
	fmt.Print("userID : ")
	fmt.Scanln(&userID)

	ReqLogin(client, userID, data)

	go func() {
		for {
			fmt.Println("=================================================================")
			fmt.Println(" Input command number (1~5)")
			fmt.Println("=================================================================")
			fmt.Println("1. room list")
			fmt.Println("2. room create")
			fmt.Println("3. join")
			fmt.Println("4. action1")
			fmt.Println("5. quit")
			fmt.Print("choose number: ")
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
				gs.Log("Fail Stream read, err : ", err)
				break
			}

			messageType := gs_protocol.Type(gs.ReadInt32(data[0:4]))
			gs.Log("Decoding type : ", messageType)

			rawData := data[4:n]
			handler, ok := msgHandlerMapping[messageType]

			if ok {
				handler(rawData)
			} else {
				gs.Log("Fail no function defined for type", handler)
				break
			}
		}
	}()

	<-exit
}

func ReqLogin(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqLogin)
	req.UserID = proto.Int64(userUID)

	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_Login)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	gs.Logf("client send : %v\n", req)
}

func ResLogin(rawData []byte) {

	res := new(gs_protocol.ResLogin)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)
	fmt.Println("server return : user id", res.GetUserID())
	fmt.Println("server return : result", res.GetResult())
}

func ReqRoomList(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqRoomList)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_RoomList)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	gs.Logf("client send : %v\n", req)
}

func ResRoomList(rawData []byte) {

	res := new(gs_protocol.ResRoomList)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)

	fmt.Printf("server return : room list : %v", res.GetRoomIDs())
}

func ReqCreate(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqCreate)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_Create)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	gs.Logf("client send : %v\n", req)
}

func ResCreate(rawData []byte) {

	res := new(gs_protocol.ResCreate)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)

	fmt.Println("server return : user id", res.GetUserID())
	fmt.Println("server return : room id", res.GetRoomID())
}

func ReqJoin(c net.Conn, userUID int64, data []byte, roomID int64) {
	req := new(gs_protocol.ReqJoin)
	req.UserID = proto.Int64(userUID)
	req.RoomID = proto.Int64(roomID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_Join)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	gs.Logf("client send : %v\n", req)
}

func ResJoin(rawData []byte) {

	res := new(gs_protocol.ResJoin)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)

	fmt.Println("server return : user id", res.GetUserID())
	fmt.Println("server return : room id", res.GetRoomID())
	fmt.Printf("server return : members %v", res.GetMembers())
}

func ReqAction1(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqAction1)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_DefinedAction1)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	gs.Logf("client send : %v\n", req)
}

func ResAction1(rawData []byte) {

	res := new(gs_protocol.ResAction1)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)

	fmt.Println("server return : user id : ", res.GetUserID())
	fmt.Println("server return : result : ", res.GetResult())

}

func ReqQuit(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqQuit)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_Quit)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...)

	_, err = c.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	gs.Logf("client send : %v\n", req)
}

func ResQuit(rawData []byte) {
	res := new(gs_protocol.ResQuit)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)
	fmt.Println("server return : user id : ", res.GetIsSuccess())
}

func NotifyJoinHandler(rawData []byte) {
	res := new(gs_protocol.NotifyJoinMsg)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)
	fmt.Println("server notify return : user id : ", res.GetUserID())
	fmt.Println("server notify return : room id : ", res.GetRoomID())
}

func NotifyAction1Handler(rawData []byte) {
	res := new(gs_protocol.NotifyAction1Msg)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)
	fmt.Println("server notify return : user id : ", res.GetUserID())
}

func NotifyQuitHandler(rawData []byte) {
	res := new(gs_protocol.NotifyQuitMsg)
	err := proto.Unmarshal(rawData, res)
	gs.CheckError(err)
	fmt.Println("server notify return : user id : ", res.GetUserID())
}
