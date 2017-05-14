package lib

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

func GetNow() time.Time {
	t := time.Now()
	return t
}

func GetDateTime() time.Time {
	t := GetNow()
	t.Format("2016-01-01 23:59:59")
	return t
}

func Atoi64(s string) int64 {
	integer, err := strconv.Atoi(s)
	CheckError(err)
	return int64(integer)
}

func Atoi32(s string) int32 {
	integer, err := strconv.Atoi(s)
	CheckError(err)
	return int32(integer)
}

func Atoi(s string) int {
	integer, err := strconv.Atoi(s)
	CheckError(err)
	return integer
}

func Itoa64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Itoa32(i int32) string {
	return strconv.Itoa(int(i))
}

func Itoa(i int) string {
	return strconv.Itoa(i)
}

func Log(a ...interface{}) {
	fmt.Println(a...)
}

func Logf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	}
}

// http://stackoverflow.com/questions/16888357/convert-an-integer-to-a-byte-array
func ReadInt32(data []byte) (ret int32) {
	ret = int32(binary.BigEndian.Uint32(data)) // fastest convert method, do not use "binary.Read"
	return
}

// After benchmarking the "encoding/binary" way, it takes almost 4 times longer than int -> string -> byte
func WriteInt32(n int32) (buf []byte) {
	buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n)) // fastest convert method, do not use "binary.Write"
	return
}

func RandInt64(min int64, max int64) int64 {
	return min + rand.Int63n(max-min)
}

func RandInt32(min int32, max int32) int32 {
	return min + rand.Int31n(max-min)
}

func Int64SliceToString(set []int64) (str string) {
	str += "["
	for _, value := range set {
		str += "," + Itoa64(value)
	}
	str += "]"
	return str
}

func WriteMsg(c net.Conn, contents []byte) bool {
	size := WriteInt32(int32(len(contents)))
	msg := append(size, contents...)
	_, err := c.Write(msg) // send data to client
	if err != nil {
		Log(err)
		return false
	}
	return true
}
