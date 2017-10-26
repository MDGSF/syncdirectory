package syncdirectory

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"

	proto "github.com/golang/protobuf/proto"
)

/*
Msg message struct
*/
type Msg struct {
	Msgtype int
	Data    []byte
}

/*
PackToJSON create json string.
*/
func PackToJSON(msgtype int, data []byte) ([]byte, error) {
	msgBag := &Msg{}
	msgBag.Msgtype = msgtype
	msgBag.Data = data

	b, err := json.Marshal(msgBag)
	if err != nil {
		return nil, err
	}
	return b, nil
}

/*
UnpackJSON unpack json string.
*/
func UnpackJSON(msg []byte) (msgtype int, data []byte, err error) {
	msgBag := &Msg{}
	err = json.Unmarshal(msg, &msgBag)
	if err != nil {
		return 0, nil, err
	}
	return msgBag.Msgtype, msgBag.Data, nil
}

/*
IntToBytes int to bytes
*/
func IntToBytes(n int) []byte {
	temp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, temp)
	return bytesBuffer.Bytes()
}

/*
BytesToInt bytes to int
*/
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var temp int32
	binary.Read(bytesBuffer, binary.BigEndian, &temp)
	return int(temp)
}

/*
Read read msg with header 4 byte.
*/
func Read(conn net.Conn) ([]byte, int, error) {

	//conn.SetReadDeadline(time.Now().Add(1e9))

	headerBuf := make([]byte, 4)
	headerLen, err := conn.Read(headerBuf)
	if err != nil || headerLen != 4 {
		Log.Println("read header failed:", err, headerLen)
		return nil, 0, errors.New("read header failed")
	}

	bodyLen := BytesToInt(headerBuf)
	//Log.Printf("read head success, body len:%d\n", bodyLen)

	//bodyBuf := make([]byte, bodyLen)
	//readedbodyLen, err := conn.Read(bodyBuf)
	//if err != nil || readedbodyLen != int(bodyLen) {
	//	Log.Println("read body failed:", err.Error(), readedbodyLen)
	//	return nil, 0, errors.New("read body failed")
	//}
	//Log.Println("read body success,", readedbodyLen, bodyBuf)

	bodyBuf := make([]byte, bodyLen)
	readedbodyLen, err := io.ReadFull(conn, bodyBuf)
	if err != nil || readedbodyLen != int(bodyLen) {
		Log.Println("read body failed:", err, readedbodyLen)
		return nil, 0, errors.New("read body failed")
	}
	//Log.Println("read body success,", readedbodyLen, bodyBuf)

	return bodyBuf, bodyLen, nil
}

/*
Write write msg with header 4 byte
*/
func Write(conn net.Conn, msg string) error {

	msgLen := strings.Count(msg, "") - 1

	//Log.Println("Write", msgLen, IntToBytes(msgLen))
	_, err := conn.Write(IntToBytes(msgLen))
	if err != nil {
		Log.Println("Write header failed.", err)
		return err
	}

	//Log.Println("Write", []byte(msg))

	_, err = conn.Write([]byte(msg))
	if err != nil {
		Log.Println("Write failed:", err)
		return err
	}

	return nil
}

/*
SendMsg : send msg with msgCode.
*/
func SendMsg(conn net.Conn, msgCode int, msg proto.Message) {
	protob, err := proto.Marshal(msg)
	if err != nil {
		Log.Println("marshal MInitDirectory failed")
		return
	}

	b, err := PackToJSON(msgCode, []byte(protob))
	if err != nil {
		Log.Println("pack to json MInitDirectory failed")
		return
	}

	Write(conn, string(b))
}
