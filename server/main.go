package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
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
	rooms lib.SharedMap
)

type Message struct {
	userID    int64  // sender
	timestamp int    // send time
	contents  []byte // serialized google protocol-buffer message
}

func NewMessage(userID int64, msg []byte) *Message {
	return &Message{
		userID,
		int(time.Now().Unix()),
		msg,
	}
}

func InitRooms() {
	rooms = lib.NewSMap(lib.RWMutex)
	rand.Seed(time.Now().UTC().UnixNano())
}

func onClientWrite(user *User, c net.Conn) {

	defer user.Leave()

	for {
		select {
		case <-user.exit:
			// when receive signal then finish the program

			lib.Log("Leave user id :" + lib.Itoa64(user.userID))

			return
		case m := <-user.recv:

			lib.Log("Client recv, user id : " + lib.Itoa64(user.userID))

			_, err := c.Write(m.contents) // send data to client
			if err != nil {

				lib.Log(err)

				return
			}
		}
	}
}

func onClientRead(user *User, c net.Conn) {

	data := make([]byte, 4096) // 4096 byte slice (dynamic resize)

	c.SetReadDeadline(time.Now().Add(30 * time.Second))
	defer c.Close() // reserve tcp connection close
	for {
		_, err := c.Read(data)
		if err != nil {

			lib.Log("Fail Stream read, err : ", err)
			break
		}

		message := &gs_protocol.Message{}
		err = proto.Unmarshal(data, message)
		if err != nil {
			lib.CheckError(err)
		}

		messageHandler(user, message)
	}

	// fail read
	user.exit <- struct{}{}
}

func onConnect(c net.Conn) {

	lib.Log("New connection: ", c.RemoteAddr())

	user := NewUser(0, nil) // empty user data

	go onClientRead(user, c)
	go onClientWrite(user, c)
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	ln, err := net.Listen("tcp", ":8000") // using TCP protocol over 8000 port
	defer ln.Close()                      // reserve listen wait close
	if err != nil {
		lib.Log(err)
		return
	}

	InitRooms()

	for {
		conn, err := ln.Accept() // server accept client connection -> return connection
		if err != nil {
			lib.Log("Fail Accept err : ", err)
			continue
		}

		onConnect(conn)
	}
}
