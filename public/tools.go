package public

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
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
		fmt.Println("read header failed:", err, headerLen)
		return nil, 0, errors.New("read header failed")
	}

	bodyLen := BytesToInt(headerBuf)
	fmt.Printf("read head success, body len:%d\n", bodyLen)

	//bodyBuf := make([]byte, bodyLen)
	//readedbodyLen, err := conn.Read(bodyBuf)
	//if err != nil || readedbodyLen != int(bodyLen) {
	//	fmt.Println("read body failed:", err.Error(), readedbodyLen)
	//	return nil, 0, errors.New("read body failed")
	//}
	//fmt.Println("read body success,", readedbodyLen, bodyBuf)

	bodyBuf := make([]byte, bodyLen)
	readedbodyLen, err := io.ReadFull(conn, bodyBuf)
	if err != nil || readedbodyLen != int(bodyLen) {
		fmt.Println("read body failed:", err, readedbodyLen)
		return nil, 0, errors.New("read body failed")
	}
	fmt.Println("read body success,", readedbodyLen, bodyBuf)

	return bodyBuf, bodyLen, nil
}

/*
Write write msg with header 4 byte
*/
func Write(conn net.Conn, msg string) error {

	msgLen := strings.Count(msg, "") - 1

	fmt.Println("Write", msgLen, IntToBytes(msgLen))
	_, err := conn.Write(IntToBytes(msgLen))
	if err != nil {
		fmt.Println("Write header failed.", err)
		return err
	}

	fmt.Println("Write", []byte(msg))

	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Write failed:", err)
		return err
	}

	return nil
}
