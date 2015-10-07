package main

import (
	gs "goserver/lib"
	"math"
)

type Room struct {
	users    gs.SharedMap // 방에 있는 유저
	messages chan *Message   // 이벤트를 위한 채널
	roomID   int64           // 룸 아이디
}

func GetRoom(roomID int64) (r *Room) {
	value, ok := rooms.Get(roomID)

	if !ok {
		gs.Log("err, not exist room : ", roomID)
	}
	r = value.(*Room)
	return
}

func GetRandomRoomID() (uuid int64) {
	for {
		//		uuid = gs.RandInt64(1, math.MaxInt64)
		uuid = int64(gs.RandInt32(1, math.MaxInt32))
		if _, ok := rooms.Get(uuid); ok {
			gs.Log("retry rand room id")
			continue // 다시 뽑아라
		}
		break
	}
	return
}

// 생성자
func NewRoom(roomID int64) (r *Room) {
	r = new(Room)
	r.messages = make(chan *Message)
	r.users = gs.NewSMap() // 맵을 한번더 wrapping 한 객체임
	r.roomID = roomID
	go r.RoomMessageLoop() // 방을 생성하면 메시지 채널을 계속 브로드 캐스팅함
	return
}

func (r *Room) RoomMessageLoop() {
	for m := range r.messages { // range 는 이벤트가 들어오는 동안 계속 루프를 돈다. (닫히면 루프종료 ㅎㅎ쩌러)
		for userID, _  := range r.users.Map() {
			if userID != m.userID {
				value, ok := r.users.Get(userID)
				if ok {
					user := value.(*User)
					gs.Log("Push message for broadcast :" + gs.Itoa64(user.userID))
					user.Push(m)
				}
			}
		}
	}
}

func (r *Room) getRoomUsers() (userIDs []int64) {
	userIDs = r.users.GetKeys()
	return userIDs
}

func (r *Room) Leave(userID int64) {
	r.users.Remove(userID) // 룸 유저리스트에서 제거

	if r.users.Count() == 0 { // 빈방이 될 경우 방 폭파
		close(r.messages)       // 브로드캐스트 채널 닫기
		rooms.Remove(r.roomID) // 방 리스트에서 제거
	}
}
