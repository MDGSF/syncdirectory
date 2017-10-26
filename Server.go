package syncdirectory

import (
	"net"
	"os"

	"github.com/golang/protobuf/proto"
)

func StartServer() {

	InitLog("server.log")

	createStoreLocation()

	l, err := net.Listen(ConnectionType, ServerHost+":"+ServerPort)
	if err != nil {
		Log.Println("Listen failed:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	Log.Println("Listen success on " + ServerHost + ":" + ServerPort)

	for {
		conn, err := l.Accept()
		if err != nil {
			Log.Println("Accept failed:", err.Error())
		}

		Log.Printf("\nNew connection:\n")
		go handleRequest(conn)
	}
}

func createStoreLocation() {
	if exists, _ := PathExists(SStoreLocation); !exists {
		if err := os.Mkdir(SStoreLocation, os.ModePerm); err != nil {
			Log.Println("Mkdir failed", SStoreLocation)
			os.Exit(1)
		}
	}
}

func handleRequest(conn net.Conn) {

	defer conn.Close()

	for {
		msg, msgLen, err := Read(conn)
		if err != nil {
			Log.Println("Read failed.", err.Error())
			return
		}

		msgtype, data, err := UnpackJSON(msg)
		if err != nil {
			Log.Println("UnpackJSON failed.", err.Error())
			return
		}
		Log.Printf("UnpackJSON success, msgLen = %d, msgtype = %d\n", msgLen, msgtype)

		v, ok := M[msgtype]
		if !ok {
			Log.Println("Invalid msgtype:", msgtype)
			return
		}

		err = v(conn, data)
		if err != nil {
			Log.Println("Process failed:", err)
			return
		}
	}
}

func ProcessInitDirectory(conn net.Conn, data []byte) error {
	Log.Printf("ProcessInitDirectory\n")

	msg := &MInitDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		Log.Println("Unmarshal MInitDirectory failed")
		return err
	}

	Log.Println(msg)

	newRoot := SStoreLocation + string(os.PathSeparator) + msg.GetRoot()
	Log.Println(newRoot)

	exists, _ := PathExists(newRoot)
	if exists {
		os.RemoveAll(newRoot)
	}

	if err := os.Mkdir(newRoot, os.ModePerm); err != nil {
		Log.Println("Mkdir failed", newRoot)
		return err
	}

	Log.Printf("Mkdir [%s] successfully.\n", newRoot)

	return nil
}

func ProcessPushDirectory(conn net.Conn, data []byte) error {
	Log.Println("ProcessPushDirectory")

	msg := &MPushDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		Log.Println("Unmarshal MPushDirectory failed")
		return err
	}

	Log.Println(msg.GetRoot(), msg.GetDirname(), msg.GetSubdirname(), msg.GetSubfilename())

	path := SStoreLocation + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetDirname()
	if exists, _ := PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			Log.Println("Mkdir failed", path)
			return err
		}
	}

	return nil
}

func ProcessPushFile(conn net.Conn, data []byte) error {
	Log.Println("ProcessPushFile")
	return PushFileRecv(conn, data, SStoreLocation)
}

func ProcessDeleteFile(conn net.Conn, data []byte) error {
	Log.Println("ProcessDeleteFile")

	msg := &MDeleteFile{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		Log.Println("Unmarshal MDeleteFile failed")
		return err
	}

	Log.Println(msg.GetRoot(), msg.GetRelativeFileWithPath())

	fileWithPath := SStoreLocation + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetRelativeFileWithPath()
	Log.Println(fileWithPath)

	err = os.RemoveAll(fileWithPath)
	if err != nil {
		Log.Println("Delete file failed:", fileWithPath)
		return err
	}

	Log.Printf("Delete file %s success.\n", fileWithPath)

	return nil
}

func ProcessMoveFile(conn net.Conn, data []byte) error {
	Log.Println("ProcessMoveFile")

	msg := &MMoveFile{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		Log.Println("Unmarshal MPushFile failed")
		return err
	}

	Log.Println(msg.GetRoot(), msg.GetOldFileWithPath(), msg.GetNewFileWithPath())

	old := SStoreLocation + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetOldFileWithPath()
	new := SStoreLocation + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetNewFileWithPath()

	err = os.Rename(old, new)
	if err != nil {
		Log.Println("os.Rename failed")
		return err
	}

	Log.Printf("Move file from %s to %s success.\n", old, new)

	return nil
}

func ProcessPullDirectoryRequest(conn net.Conn, data []byte) error {
	Log.Println("ProcessPullDirectoryRequest")

	msg := &MPullDirectoryRequest{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		Log.Println("Unmarshal MPushFile failed")
		return err
	}

	Log.Println(msg.GetRoot())

	path := SStoreLocation + string(os.PathSeparator) + msg.GetRoot()
	exists, err := PathExists(path)
	if !exists {
		//TODO send response to client: root is not exists.
		return err
	}

	return nil
}
