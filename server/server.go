package main

import (
	"fmt"
	"io"
	"jian/tcp/tools"
	"net"
	"os"
	"syncdirectory"
	"syncdirectory/public"

	"github.com/golang/protobuf/proto"
)

const (
	STORE_LOCATION = "E:\\ServerStore"
)

func main() {
	createStoreLocation()

	l, err := net.Listen(public.CONN_TYPE, public.CONN_HOST+":"+public.CONN_PORT)
	if err != nil {
		fmt.Println("listen failed:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on " + public.CONN_HOST + ":" + public.CONN_PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept failed:", err.Error())
		}

		fmt.Printf("\n")
		go handleRequest(conn)
	}
}

func createStoreLocation() {
	if exists, _ := public.PathExists(STORE_LOCATION); !exists {
		if err := os.Mkdir(STORE_LOCATION, os.ModePerm); err != nil {
			fmt.Println("mkdir failed", STORE_LOCATION)
			return
		}
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
		//fmt.Println("msg and msgLen", msg, msgLen)

		msgtype, data, err := tools.UnpackJSON(msg)
		if err != nil {
			fmt.Println("UnpackJSON failed.", err.Error())
			return
		}
		//fmt.Println("msgtype and data", msgtype, data)
		fmt.Printf("msgLen = %d, msgtype = %d\n", msgLen, msgtype)

		processMsg(conn, msgtype, data)
	}
}

func processMsg(conn net.Conn, msgtype int, data []byte) error {
	switch msgtype {
	case int(syncdirectory.ESyncMsgCode_EInitDirectory):
		processInitDirectory(conn, data)
	case int(syncdirectory.ESyncMsgCode_EPushDirectory):
		processPushDirectory(conn, data)
	case int(syncdirectory.ESyncMsgCode_EDeleteDirectory):
		processDeleteDirectory(conn, data)
	case int(syncdirectory.ESyncMsgCode_EPushFile):
		processPushFile(conn, data)
	case int(syncdirectory.ESyncMsgCode_EDeleteFile):
		processDeleteFile(conn, data)
	case int(syncdirectory.ESyncMsgCode_EMoveFile):
		processMoveFile(conn, data)
	default:
		fmt.Println("Unknown msgtype", msgtype)
	}

	return nil
}

func processInitDirectory(conn net.Conn, data []byte) error {
	fmt.Printf("processInitDirectory\n")

	msg := &syncdirectory.MInitDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Unmarshal MInitDirectory failed")
		return err
	}

	fmt.Println(msg)

	newRoot := STORE_LOCATION + "\\" + msg.GetRoot()
	fmt.Println(newRoot)

	exists, _ := public.PathExists(newRoot)
	if exists {
		os.RemoveAll(newRoot)
	}

	if err := os.Mkdir(newRoot, os.ModePerm); err != nil {
		fmt.Println("mkdir failed", newRoot)
		return err
	}

	fmt.Printf("create %s successfully.\n", newRoot)

	return nil
}

func processPushDirectory(conn net.Conn, data []byte) error {
	fmt.Println("processPushDirectory")

	msg := &syncdirectory.MPushDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Unmarshal MPushDirectory failed")
		return err
	}

	fmt.Println(msg.GetRoot(), msg.GetDirname(), msg.GetSubdirname(), msg.GetSubfilename())

	path := STORE_LOCATION + "\\" + msg.GetRoot() + string(os.PathSeparator) + msg.GetDirname()
	if exists, _ := public.PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			fmt.Println("mkdir failed", path)
			return err
		}
	}

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

	fmt.Println(push.GetRoot(), push.GetFileName(), push.GetFileSize(), push.GetFileDir())

	path := STORE_LOCATION + "\\" + push.GetRoot()
	if len(push.GetFileDir()) != 0 {
		path = path + "\\" + push.GetFileDir()
	}
	fileWithPath := path + "\\" + push.GetFileName()
	fmt.Println("new file path", fileWithPath)

	if exists, _ := public.PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			fmt.Println("mkdir failed", path)
			return err
		}
	}

	f, err := os.OpenFile(fileWithPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	io.CopyN(f, conn, push.GetFileSize())

	return nil
}

func processDeleteDirectory(conn net.Conn, data []byte) error {
	fmt.Println("processDeleteDirectory")
	return nil
}

func processDeleteFile(conn net.Conn, data []byte) error {
	fmt.Println("processDeleteFile")
	return nil
}

func processMoveFile(conn net.Conn, data []byte) error {
	fmt.Println("processMoveFile")

	msg := &syncdirectory.MMoveFile{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Unmarshal MPushFile failed")
		return err
	}

	fmt.Println(msg.GetRoot(), msg.GetOldFileWithPath(), msg.GetNewFileWithPath())

	old := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + "\\" + msg.GetOldFileWithPath()
	new := STORE_LOCATION + "\\" + msg.GetRoot() + "\\" + msg.GetNewFileWithPath()

	err = os.Rename(old, new)
	if err != nil {
		fmt.Println("osRename failed")
		return err
	}

	return nil
}
