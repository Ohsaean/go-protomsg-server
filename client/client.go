package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/lxn/walk"
	walk_dcl "github.com/lxn/walk/declarative"
	"github.com/ohsaean/gogpd/lib"
	"github.com/ohsaean/gogpd/protobuf"
)

// MsgHandlerFunc 메시지 핸들러
type MsgHandlerFunc func(data *gs_protocol.Message) bool

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

var inputString string

func main() {
	client, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println(err)
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
			messageType := message.Type

			handler, ok := msgHandlerMapping[messageType]
			log.Println("recv message type : ", messageType)
			if ok {
				ret := handler(message) // calling proper handler function
				if !ret {
					log.Println("Fail handler process", handler)
				}
			} else {
				log.Println("Fail no function defined for type", handler)
				break
			}
		}
	}()

	mw.Run()
}

func ReqLogin(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Type: gs_protocol.Type_Login,
		ReqLogin: &gs_protocol.ReqLogin{
			UserID: userUID,
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("ReqLogin client send : %x\n", msg)
}

func ResLogin(data *gs_protocol.Message) bool {
	res := data.GetResLogin()
	if res == nil {
		lib.Log("fail, GetReqLogin()")
		return false
	}

	log.Println("ResLogin server return : user id : " + lib.Itoa64(res.UserID))
	log.Println("ResLogin server return : result code : " + lib.Itoa32(res.Result))
	return true
}

func ReqRoomList(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Type: gs_protocol.Type_RoomList,
		ReqRoomList: &gs_protocol.ReqRoomList{
			UserID: userUID,
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("ReqLogin client send : %x\n", msg)
}

func ResRoomList(data *gs_protocol.Message) bool {

	res := data.GetResRoomList()
	if res == nil {
		lib.Log("fail, GetReqLogin()")
		return false
	}

	log.Println("ResLogin server return : user id : ", res.UserID)

	for _, roomID := range res.RoomIDs {
		log.Println("ResLogin server return : room id : ", roomID)
	}

	return true
}

func ReqCreate(c net.Conn, userUID int64, data []byte) {
	req := &gs_protocol.Message{
		Type: gs_protocol.Type_Create,
		ReqCreate: &gs_protocol.ReqCreate{
			UserID: userUID,
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("ReqLogin client send : %x\n", msg)
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
		Type: gs_protocol.Type_Join,
		ReqJoin: &gs_protocol.ReqJoin{
			UserID: userUID,
			RoomID: roomID,
		},
	}
	msg, err := proto.Marshal(req)
	lib.CheckError(err)
	_, err = c.Write(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("ReqJoin client send : %x\n", msg)
}

func ResJoin(data *gs_protocol.Message) bool {

	res := data.GetResJoin()
	if res == nil {
		lib.Log("fail, GetReqLogin()")
		return false
	}

	log.Println("ResLogin server return : user id : ", res.UserID)
	log.Println("ResLogin server return : room id : ", res.RoomID)

	for _, memberID := range res.Members {
		log.Println("ResLogin server return : room id : ", memberID)
	}

	return true
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

func RunUserIdDialog(owner walk.Form) (int, error) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var inDlg *walk.LineEdit

	return Dialog{
		AssignTo:      &dlg,
		Title:         "input User ID",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{200, 100},
		Layout:        VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "User ID:",
					},
					LineEdit{
						AssignTo: &inDlg,
						Text:     "",
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							inputString = inDlg.Text()
							dlg.Accept()
						},
					},
					PushButton{
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

	return Dialog{
		AssignTo:      &dlg,
		Title:         "input Room ID",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{200, 100},
		Layout:        VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "room id:",
					},
					LineEdit{
						AssignTo: &inDlg,
						Text:     "",
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							inputString = inDlg.Text()
							dlg.Accept()
						},
					},
					PushButton{
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
