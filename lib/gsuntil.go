package gsutil

import (
	"strconv"
	"fmt"
	"os"
	"encoding/binary"
	"math/rand"
	"goserver/protobuf"
)

func Itoa64(i int64) string {
	return strconv.FormatInt(i, 10)
}


func Log(a ...interface{}) {
	fmt.Println(a...);
}

func Logf(format string, a ...interface{}) {
	fmt.Printf(format, a...);
}


func CheckError(err error){
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

func WriteMsgType(n gs_protocol.Type) (buf []byte) {
	buf = WriteInt32(int32(n))
	return
}


func RandInt64(min int64, max int64) int64 {
	return min + rand.Int63n(max-min)
}

func RandInt32(min int32, max int32) int32 {
	return min + rand.Int31n(max-min)
}



