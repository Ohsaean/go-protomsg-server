package main

import (
	gs "go-protomsg-server/lib"
	"go-protomsg-server/protobuf"
	"math"
	"math/rand"
	"net"
	"runtime"
	"time"
)

// 서버 설정값들?
const (
	maxRoom = math.MaxInt32
)

// 전역 변수
var (
	rooms gs.SharedMap
)

// 메시지 객체
type Message struct {
	userID    int64            // 보낸 사람
	msgType   gs_protocol.Type // 메시지 타입
	timestamp int              // 시간 값
	content   []byte           // 각 클라이언트에게 전달할 내용 (serialized protobuf message?)
}

func NewMessage(userID int64, eventType gs_protocol.Type, msg []byte) *Message {
	return &Message{
		userID,
		eventType,
		int(time.Now().Unix()),
		msg,
	}
}

// 방 초기화
func InitRooms() {
	rooms = gs.NewSMap()
	rand.Seed(time.Now().UTC().UnixNano())
}

func ClientSender(user *User, c net.Conn) {

	defer user.Leave()

	for {
		select {
		case <-user.exit:
			// exit 채널에 signal 이 오면 종료
			gs.Log("Leave user :" + gs.Itoa64(user.userID))
			return
		case m := <-user.recv:
			// 유저에게 할당된 recv 채널
			msgTypeBytes := gs.WriteMsgType(m.msgType)
			msg := append(msgTypeBytes, m.content...) // 헤더 + 메시지 타입을 붙임 (slice+slice 시에는'...' 붙여야함!)
			gs.Log("Client recv, user : " + gs.Itoa64(user.userID))
			_, err := c.Write(msg) // 클라이언트로 데이터를 보냄
			if err != nil {
				gs.Log(err)
				return
			}
		}
	}
}

func ClientReader(user *User, c net.Conn) {

	data := make([]byte, 4096) // 4096 크기의 바이트 슬라이스 생성 (동적 확장됨)

	for {
		n, err := c.Read(data) // conn 에서 한줄 빼와본다
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
			handler(user, rawData) // 핸들러 호출
		} else {
			gs.Log("No function defined for type", handler)
			break
		}
	}

	// 읽기 실패이면 종료한다.
	user.exit <- struct{}{}
}

// On Client Connect
func ClientHandler(c net.Conn) {
	gs.Log("New Connection: ", c.RemoteAddr())
	user := NewUser(0, nil) // empty user data
	go ClientReader(user, c)
	go ClientSender(user, c)
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	ln, err := net.Listen("tcp", ":8000") // TCP 프로토콜에 8000 포트로 연결을 받음
	if err != nil {
		gs.Log(err)
		return
	}

	InitRooms()

	defer ln.Close() // main 함수가 끝나기 직전에 연결 대기를 닫음
	for {
		conn, err := ln.Accept() // 클라이언트가 연결되면 TCP 연결을 리턴
		if err != nil {
			gs.Log("Accept Error :", err)
			continue
		}
		defer conn.Close() // main 함수가 끝나기 직전에 TCP 연결을 닫음

		go ClientHandler(conn) // 패킷을 처리할 함수를 고루틴으로 실행
	}
}
