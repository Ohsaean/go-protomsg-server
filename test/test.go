package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/ohsaean/gogpd/protobuf"
	"log"
)

func main() {
	msg := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqLogin{
			ReqLogin: &gs_protocol.ReqLogin{
				UserID: 1,
			},
		},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	newMsg := &gs_protocol.Message{}
	err = proto.Unmarshal(data, newMsg)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	switch t := newMsg.Payload.(type) {
	case *gs_protocol.Message_ReqLogin:
		fmt.Printf("type : %T\n", t)
		fmt.Println(t.ReqLogin)
		fmt.Println(newMsg.GetReqLogin())
		break
	case *gs_protocol.Message_ReqCreate:
		fmt.Println("create")
	}
}
