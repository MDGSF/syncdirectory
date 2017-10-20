package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syncdirectory"
	"syncdirectory/public"

	"github.com/golang/protobuf/proto"
)

func StartServer() {
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

		fmt.Printf("\nnew Connection:\n")
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
		msg, msgLen, err := public.Read(conn)
		if err != nil {
			fmt.Println("Read failed.", err.Error())
			return
		}
		//fmt.Println("msg and msgLen", msg, msgLen)

		msgtype, data, err := public.UnpackJSON(msg)
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

	if v, ok := M[msgtype]; ok {
		v(conn, data)
	} else {
		fmt.Println("Invalid msgtype:", msgtype)
		return errors.New("Invalid msgtype")
	}

	return nil

	//switch msgtype {
	//case int(syncdirectory.ESyncMsgCode_EInitDirectory):
	//	ProcessInitDirectory(conn, data)
	//case int(syncdirectory.ESyncMsgCode_EPushDirectory):
	//	ProcessPushDirectory(conn, data)
	//case int(syncdirectory.ESyncMsgCode_EPushFile):
	//	ProcessPushFile(conn, data)
	//case int(syncdirectory.ESyncMsgCode_EDeleteFile):
	//	ProcessDeleteFile(conn, data)
	//case int(syncdirectory.ESyncMsgCode_EMoveFile):
	//	ProcessMoveFile(conn, data)
	//default:
	//	fmt.Println("Unknown msgtype", msgtype)
	//}

	//return nil
}

func ProcessInitDirectory(conn net.Conn, data []byte) error {
	fmt.Printf("ProcessInitDirectory\n")

	msg := &syncdirectory.MInitDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Unmarshal MInitDirectory failed")
		return err
	}

	fmt.Println(msg)

	newRoot := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot()
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

func ProcessPushDirectory(conn net.Conn, data []byte) error {
	fmt.Println("ProcessPushDirectory")

	msg := &syncdirectory.MPushDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Unmarshal MPushDirectory failed")
		return err
	}

	fmt.Println(msg.GetRoot(), msg.GetDirname(), msg.GetSubdirname(), msg.GetSubfilename())

	path := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetDirname()
	if exists, _ := public.PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			fmt.Println("mkdir failed", path)
			return err
		}
	}

	return nil
}

func ProcessPushFile(conn net.Conn, data []byte) error {
	fmt.Println("ProcessPushFile")

	push := &syncdirectory.MPushFile{}
	err := proto.Unmarshal(data, push)
	if err != nil {
		fmt.Println("Unmarshal MPushFile failed")
		return err
	}

	fmt.Println(push.GetRoot(), push.GetFileName(), push.GetFileSize(), push.GetFileDir())

	path := STORE_LOCATION + string(os.PathSeparator) + push.GetRoot()
	if len(push.GetFileDir()) != 0 {
		path = path + string(os.PathSeparator) + push.GetFileDir()
	}
	fileWithPath := path + string(os.PathSeparator) + push.GetFileName()
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

func ProcessDeleteFile(conn net.Conn, data []byte) error {
	fmt.Println("ProcessDeleteFile")

	msg := &syncdirectory.MDeleteFile{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Unmarshal MDeleteFile failed")
		return err
	}

	fmt.Println(msg.GetRoot(), msg.GetRelativeFileWithPath())

	fileWithPath := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetRelativeFileWithPath()
	fmt.Println(fileWithPath)

	err = os.RemoveAll(fileWithPath)
	if err != nil {
		fmt.Println("Delete file failed:", fileWithPath)
		return err
	}

	fmt.Printf("Delete file %s success.\n", fileWithPath)

	return nil
}

func ProcessMoveFile(conn net.Conn, data []byte) error {
	fmt.Println("ProcessMoveFile")

	msg := &syncdirectory.MMoveFile{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Unmarshal MPushFile failed")
		return err
	}

	fmt.Println(msg.GetRoot(), msg.GetOldFileWithPath(), msg.GetNewFileWithPath())

	old := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetOldFileWithPath()
	new := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetNewFileWithPath()

	err = os.Rename(old, new)
	if err != nil {
		fmt.Println("osRename failed")
		return err
	}

	return nil
}
