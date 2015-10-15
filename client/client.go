package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	gs "goserver/lib"
	"goserver/protobuf"
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
	client, err := net.Dial("tcp", "127.0.0.1:8000") // TCP 프로토콜, 127.0.0.1:8000 서버에 연결
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close() // main 함수가 끝나기 직전에 TCP 연결을 닫음

	data := make([]byte, 4096)
	exit := make(chan struct{})

	var userID int64
	var method int
	fmt.Println("=================================================================")
	fmt.Println(" input userID (it must be a whole number greater than 0")
	fmt.Println("=================================================================")
	fmt.Print("userID : ")
	fmt.Scanln(&userID)

	// 로그인은 기본으로..
	ReqLogin(client, userID, data)

	// 송신 고루틴
	go func() {
		for {
			fmt.Println("=================================================================")
			fmt.Println(" input method number (1~5)")
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

	// 수신 고루틴
	go func() {
		data := make([]byte, 4096) // 4096 크기의 바이트 슬라이스 생성 (동적 확장됨)

		for {
			n, err := client.Read(data) // conn 에서 한줄 빼와본다
			if err != nil {
				gs.Log("stream read error : ", err)
				break
			}

			// header - body 형태로 (두개의 패킷이 한 라인에 와야함)
			messageType := gs_protocol.Type(gs.ReadInt32(data[0:4])) // 메시지 타입
			//	bodySize := gs.ReadInt32(data[4:8]) // body (serialized protobuf message) size

			// body 확보
			rawData := data[4:n] // 4~끝까지 읽으면 될려나??; <-- 이거 안되면 바디사이즈 계산해서 참조하도록 해야함

			gs.Log("Decoding type : ", messageType)

			handler, ok := msgHandlerMapping[messageType]
			if ok {
				handler(rawData) // 핸들러 호출
			} else {
				gs.Log("No function defined for type", handler)
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
	data = append(msgTypeBytes, msg...) // 메시지 타입을 붙임

	_, err = c.Write(data) // 서버로 데이터를 보냄
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
	fmt.Println("server return : userid", res.GetUserID()) // 데이터 출력
	fmt.Println("server return : result", res.GetResult()) // 데이터 출력
}

func ReqRoomList(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqRoomList)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_RoomList)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...) // 메시지 타입을 붙임

	_, err = c.Write(data) // 서버로 데이터를 보냄
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

	fmt.Printf("server return : room list : %v", res.GetRoomIDs()) // 데이터 출력
}

func ReqCreate(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqCreate)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_Create)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...) // 메시지 타입을 붙임

	_, err = c.Write(data) // 서버로 데이터를 보냄
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

	fmt.Println("server return : user id", res.GetUserID()) // 데이터 출력
	fmt.Println("server return : room id", res.GetRoomID()) // 데이터 출력
}

func ReqJoin(c net.Conn, userUID int64, data []byte, roomID int64) {
	req := new(gs_protocol.ReqJoin)
	req.UserID = proto.Int64(userUID)
	req.RoomID = proto.Int64(roomID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_Join)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...) // 메시지 타입을 붙임

	_, err = c.Write(data) // 서버로 데이터를 보냄
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

	fmt.Println("server return : user id", res.GetUserID())    // 데이터 출력
	fmt.Println("server return : room id", res.GetRoomID())    // 데이터 출력
	fmt.Printf("server return : members %v", res.GetMembers()) // 데이터 출력
}

func ReqAction1(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqAction1)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_DefinedAction1)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...) // 메시지 타입을 붙임

	_, err = c.Write(data) // 서버로 데이터를 보냄
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

	fmt.Println("server return : user id : ", res.GetUserID()) // 데이터 출력
	fmt.Println("server return : result : ", res.GetResult())  // 데이터 출력

}

func ReqQuit(c net.Conn, userUID int64, data []byte) {
	req := new(gs_protocol.ReqQuit)
	req.UserID = proto.Int64(userUID)
	msgTypeBytes := gs.WriteMsgType(gs_protocol.Type_Quit)
	msg, err := proto.Marshal(req)
	gs.CheckError(err)
	data = append(msgTypeBytes, msg...) // 메시지 타입을 붙임

	_, err = c.Write(data) // 서버로 데이터를 보냄
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
