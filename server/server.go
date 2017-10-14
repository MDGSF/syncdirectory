package main

import (
	"fmt"
	"jian/tcp/tools"
	"net"
	"os"
	"syncdirectory"

	"github.com/golang/protobuf/proto"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "10001"
	CONN_TYPE = "tcp"
	BUF_SIZE  = 4 * 1024

	STORE_LOCATION = "E:\\ServerStore"
)

func main() {
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("listen failed:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept failed:", err.Error())
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {

	defer conn.Close()

	for {
		msg, msgLen, err := tools.Read(conn)
		if err != nil {
			fmt.Println("Read failed.", err.Error())
			return
		}
		fmt.Println("msg and msgLen", msg, msgLen)

		msgtype, data, err := tools.UnpackJSON(msg)
		if err != nil {
			fmt.Println("UnpackJSON failed.", err.Error())
			return
		}
		fmt.Println("msgtype and data", msgtype, data)

		processMsg(conn, msgtype, data)
	}
}

func processMsg(conn net.Conn, msgtype int, data []byte) error {
	switch msgtype {
	case int(syncdirectory.ESyncMsgCode_EInitDirectory):
		processInitDirectory(conn, data)
	case int(syncdirectory.ESyncMsgCode_EPushFile):
		processPushFile(conn, data)
	default:
	}

	return nil
}

func processInitDirectory(conn net.Conn, data []byte) error {
	fmt.Println("processInitDirectory")
	return nil
}

func processPushFile(conn net.Conn, data []byte) error {
	fmt.Println("processPushFile")

	push := &syncdirectory.MPushFile{}
	err := proto.Unmarshal(data, push)
	if err != nil {
		fmt.Println("Unmarshal MPushFile failed")
		return err
	}

	fmt.Println(push.GetFileName(), push.GetFileSize(), push.GetFileDir())

	return nil
}
