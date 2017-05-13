package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/lxn/walk"
	walk_dcl "github.com/lxn/walk/declarative"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
	"log"
	"net"
	"os"
	"strconv"
)

var inputString string

func main() {
	client, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Print(err)
		return
	}
	defer client.Close()

	data := make([]byte, 4096)
	//exit := make(chan bool, 1)

	var userID int64
	var mw *walk.MainWindow

	err = walk_dcl.MainWindow{
		AssignTo: &mw,
		Title:    "Walk LogView Example",
		MinSize: walk_dcl.Size{
			Width:  320,
			Height: 240,
		},
		Size: walk_dcl.Size{
			Width:  600,
			Height: 400,
		},
		Layout: walk_dcl.VBox{},
		Children: []walk_dcl.Widget{
			walk_dcl.HSplitter{
				Children: []walk_dcl.Widget{
					walk_dcl.PushButton{
						Text: "login",
						OnClicked: func() {
							if cmd, err := RunUserIdDialog(mw); err != nil {
								log.Print(err)
							} else if cmd == walk.DlgCmdOK {
								log.Println("dlg msg : " + inputString)
								num, err := strconv.Atoi(inputString)
								lib.CheckError(err)
								userID = int64(num)
								ReqLogin(client, userID, data)
							}
						},
					},
					walk_dcl.PushButton{
						Text: "room create",
						OnClicked: func() {
							log.Println("req create user id : ", userID)
							ReqCreate(client, userID, data)
						},
					},
					walk_dcl.PushButton{
						Text: "room list",
						OnClicked: func() {
							log.Println("room list user id : ", userID)
							ReqRoomList(client, userID, data)
						},
					},
					walk_dcl.PushButton{
						Text: "join",
						OnClicked: func() {
							if cmd, err := RunRoomJoinDialog(mw); err != nil {
								log.Print(err)
							} else if cmd == walk.DlgCmdOK {
								log.Println("dlg msg : " + inputString)
								num, err := strconv.Atoi(inputString)
								lib.CheckError(err)
								roomID := int64(num)
								ReqJoin(client, userID, data, roomID)
							}
						},
					},
					walk_dcl.PushButton{
						Text: "action1",
						OnClicked: func() {
							ReqAction1(client, userID, data)
						},
					},
					walk_dcl.PushButton{
						Text: "quit",
						OnClicked: func() {
							log.Println("quit user id : ", userID)
							ReqQuit(client, userID, data)
							os.Exit(3)
						},
					},
				},
			},
		},
	}.Create()

	if err != nil {
		log.Fatal(err)
	}

	lv, err := NewLogView(mw)
	if err != nil {
		log.Fatal(err)
	}

	//logFile, err := os.OpenFile("log.txt", os.O_WRONLY, 0666)
	log.SetOutput(lv)

	go func() {
		data := make([]byte, 4096)

		for {
			log.Println("wait for read")
			_, err := client.Read(data)
			if err != nil {
				log.Println("Fail Stream read, err : ", err)
				break
			}

			message := &gs_protocol.Message{}
			err = proto.Unmarshal(data, message)
			if err != nil {
				lib.CheckError(err)
			}
			messageHandler(message)
		}
	}()

	mw.Run()
}

// MessageHandler 여기서 각 proto message 에 대한 적절한 프로시저를 할당함
func messageHandler(msg *gs_protocol.Message) {
	// type switch 말고는 방법이 없나??
	switch msg.Payload.(type) {

	case *gs_protocol.Message_ResLogin:
		ResLogin(msg)

	case *gs_protocol.Message_ResCreate:
		ResCreate(msg)

	case *gs_protocol.Message_ResJoin:
		ResJoin(msg)

	case *gs_protocol.Message_ResAction1:
		ResAction1(msg)

	case *gs_protocol.Message_ResRoomList:
		ResRoomList(msg)

	case *gs_protocol.Message_ResQuit:
		ResQuit(msg)

	case *gs_protocol.Message_NotifyAction1:
		NotifyAction1Handler(msg)

	case *gs_protocol.Message_NotifyJoin:
		NotifyJoinHandler(msg)

	case *gs_protocol.Message_NotifyQuit:
		NotifyQuitHandler(msg)

	default:
		lib.Log("Error, not defined handler")
	}
}

func ReqLogin(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqLogin{
			ReqLogin: &gs_protocol.ReqLogin{
				UserID: userUID,
			},
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("ReqLogin client send :", req)
}

func ResLogin(data *gs_protocol.Message) bool {
	res := data.GetResLogin()
	if res == nil {
		lib.Log("fail, GetReqLogin()")
		return false
	} else {
		lib.Log("GetReqLogin() : ", res)
	}

	log.Println("ResLogin server return : user id : " + lib.Itoa64(res.UserID))
	log.Println("ResLogin server return : result code : " + lib.Itoa32(res.Result))
	return true
}

func ReqRoomList(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqRoomList{
			ReqRoomList: &gs_protocol.ReqRoomList{
				UserID: userUID,
			},
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("ReqRoomList client send :", req)
}

func ResRoomList(data *gs_protocol.Message) bool {

	res := data.GetResRoomList()
	if res == nil {
		lib.Log("fail, GetResRoomList()")
		return false
	}

	log.Println("GetResRoomList server return : user id : ", res.UserID)

	for _, roomID := range res.RoomIDs {
		log.Println("GetResRoomList server return : room id : ", roomID)
	}

	return true
}

func ReqCreate(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqCreate{
			ReqCreate: &gs_protocol.ReqCreate{
				UserID: userUID,
			},
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("ReqCreate client send :", req)
}

func ResCreate(data *gs_protocol.Message) bool {

	res := data.GetResCreate()
	if res == nil {
		lib.Log("fail, GetReqLogin()")
		return false
	}

	log.Println("ResCreate server return : user id : ", res.UserID)
	log.Println("ResCreate server return : room id : ", res.RoomID)

	return true
}

func ReqJoin(c net.Conn, userUID int64, data []byte, roomID int64) {
	req := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqJoin{
			ReqJoin: &gs_protocol.ReqJoin{
				UserID: userUID,
				RoomID: roomID,
			},
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("ReqJoin client send :", req)
}

func ResJoin(data *gs_protocol.Message) bool {

	res := data.GetResJoin()
	if res == nil {
		lib.Log("fail, GetReqLogin()")
		return false
	}

	log.Println("GetResJoin server return : user id : ", res.UserID)
	log.Println("GetResJoin server return : room id : ", res.RoomID)

	for _, memberID := range res.Members {
		log.Println("GetResJoin server return : member id : ", memberID)
	}

	return true
}

func ReqAction1(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqAction1{
			ReqAction1: &gs_protocol.ReqAction1{
				UserID: userUID,
			},
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)

	_, err = c.Write(msg)
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("ReqAction1 client send :", req)
}

func ResAction1(data *gs_protocol.Message) bool {

	res := data.GetResAction1()
	if res == nil {
		lib.Log("fail, GetResAction1()")
		return false
	}

	log.Println("GetResAction1 server return : user id : ", res.UserID)
	log.Println("GetResAction1 server return : result : ", res.Result)

	return true
}

func ReqQuit(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Payload: &gs_protocol.Message_ReqQuit{
			ReqQuit: &gs_protocol.ReqQuit{
				UserID: userUID,
			},
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)

	_, err = c.Write(msg)
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("ReqQuit client send :", req)
}

func ResQuit(data *gs_protocol.Message) bool {
	res := data.GetResQuit()
	if res == nil {
		lib.Log("fail, GetResAction1()")
		return false
	}

	log.Println("GetResQuit server return : user id : ", res.IsSuccess)
	return true
}

func NotifyJoinHandler(data *gs_protocol.Message) bool {
	res := data.GetNotifyJoin()
	if res == nil {
		lib.Log("fail, GetResAction1()")
		return false
	}

	log.Println("GetNotifyJoin server return : user id : ", res.UserID)
	log.Println("GetNotifyJoin server return : room id : ", res.RoomID)
	return true
}

func NotifyAction1Handler(data *gs_protocol.Message) bool {
	res := data.GetNotifyAction1()
	if res == nil {
		lib.Log("fail, GetResAction1()")
		return false
	}

	log.Println("GetNotifyAction1 server return : user id : ", res.UserID)
	return true
}

func NotifyQuitHandler(data *gs_protocol.Message) bool {
	res := data.GetNotifyQuit()
	if res == nil {
		lib.Log("fail, GetResAction1()")
		return false
	}

	log.Println("GetNotifyQuit server return : user id : ", res.UserID)
	log.Println("GetNotifyQuit server return : room id : ", res.RoomID)
	return true
}

func RunUserIdDialog(owner walk.Form) (int, error) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var inDlg *walk.LineEdit

	return walk_dcl.Dialog{
		AssignTo:      &dlg,
		Title:         "input User ID",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize: walk_dcl.Size{
			Width: 200, Height: 100},
		Layout: walk_dcl.VBox{},
		Children: []walk_dcl.Widget{
			walk_dcl.Composite{
				Layout: walk_dcl.Grid{Columns: 2},
				Children: []walk_dcl.Widget{
					walk_dcl.Label{
						Text: "User ID:",
					},
					walk_dcl.LineEdit{
						AssignTo: &inDlg,
						Text:     "",
					},
				},
			},
			walk_dcl.Composite{
				Layout: walk_dcl.HBox{},
				Children: []walk_dcl.Widget{
					walk_dcl.HSpacer{},
					walk_dcl.PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							inputString = inDlg.Text()
							dlg.Accept()
						},
					},
					walk_dcl.PushButton{
						AssignTo: &cancelPB,
						Text:     "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}

func RunRoomJoinDialog(owner walk.Form) (int, error) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var inDlg *walk.LineEdit

	return walk_dcl.Dialog{
		AssignTo:      &dlg,
		Title:         "input Room ID",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       walk_dcl.Size{Width: 200, Height: 100},
		Layout:        walk_dcl.VBox{},
		Children: []walk_dcl.Widget{
			walk_dcl.Composite{
				Layout: walk_dcl.Grid{Columns: 2},
				Children: []walk_dcl.Widget{
					walk_dcl.Label{
						Text: "room id:",
					},
					walk_dcl.LineEdit{
						AssignTo: &inDlg,
						Text:     "",
					},
				},
			},
			walk_dcl.Composite{
				Layout: walk_dcl.HBox{},
				Children: []walk_dcl.Widget{
					walk_dcl.HSpacer{},
					walk_dcl.PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							inputString = inDlg.Text()
							dlg.Accept()
						},
					},
					walk_dcl.PushButton{
						AssignTo: &cancelPB,
						Text:     "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}
