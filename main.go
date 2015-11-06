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

// server config
const (
	maxRoom = math.MaxInt32
)

// global variable
var (
	rooms gs.SharedMap
)

type Message struct {
	userID    int64 // sender
	msgType   gs_protocol.Type
	timestamp int    // send time
	content   []byte // serialized google protocol-buffer message
}

func NewMessage(userID int64, eventType gs_protocol.Type, msg []byte) *Message {
	return &Message{
		userID,
		eventType,
		int(time.Now().Unix()),
		msg,
	}
}

func InitRooms() {
	rooms = gs.NewSMap()
	rand.Seed(time.Now().UTC().UnixNano())
}

func ClientSender(user *User, c net.Conn) {

	defer user.Leave()

	for {
		select {
		case <-user.exit:
			// when receive signal then finish the program
			gs.Log("Leave user id :" + gs.Itoa64(user.userID))
			return
		case m := <-user.recv:
			// on receive message
			msgTypeBytes := gs.WriteMsgType(m.msgType)

			// msg header + msg type
			msg := append(msgTypeBytes, m.content...) // '...' need when concat between slice+slice
			gs.Log("Client recv, user id : " + gs.Itoa64(user.userID))
			_, err := c.Write(msg) // send data to client
			if err != nil {
				gs.Log(err)
				return
			}
		}
	}
}

func ClientReader(user *User, c net.Conn) {

	data := make([]byte, 4096) // 4096 byte slice (dynamic resize)

	for {
		n, err := c.Read(data)
		if err != nil {
			gs.Log("Fail Stream read, err : ", err)
			break
		}

		// header - body format (header + body in single line)
		messageType := gs_protocol.Type(gs.ReadInt32(data[0:4]))
		gs.Log("Decoding type : ", messageType)

		rawData := data[4:n] // 4~ end of line <--if fail read rawData, need calculated body size data (field)
		handler, ok := msgHandlerMapping[messageType]

		if ok {
			handler(user, rawData) // calling proper handler function
		} else {
			gs.Log("Fail no function defined for type", handler)
			break
		}
	}

	// fail read
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
	ln, err := net.Listen("tcp", ":8000") // using TCP protocol over 8000 port
	if err != nil {
		gs.Log(err)
		return
	}

	InitRooms()

	defer ln.Close() // reserve listen wait close
	for {
		conn, err := ln.Accept() // server accept client connection -> return connection
		if err != nil {
			gs.Log("Fail Accept err : ", err)
			continue
		}
		defer conn.Close() // reserve tcp connection close

		go ClientHandler(conn)
	}
}
