package server

import (
	"io"
	"net"
	"os"
	"syncdirectory"
	p "syncdirectory/public"

	"github.com/golang/protobuf/proto"
)

func StartServer() {

	p.InitLog("server.log")

	createStoreLocation()

	l, err := net.Listen(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		p.Log.Println("Listen failed:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	p.Log.Println("Listen success on " + p.CONN_HOST + ":" + p.CONN_PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			p.Log.Println("Accept failed:", err.Error())
		}

		p.Log.Printf("\nNew connection:\n")
		go handleRequest(conn)
	}
}

func createStoreLocation() {
	if exists, _ := p.PathExists(STORE_LOCATION); !exists {
		if err := os.Mkdir(STORE_LOCATION, os.ModePerm); err != nil {
			p.Log.Println("Mkdir failed", STORE_LOCATION)
			return
		}
	}
}

func handleRequest(conn net.Conn) {

	defer conn.Close()

	for {
		msg, msgLen, err := p.Read(conn)
		if err != nil {
			p.Log.Println("Read failed.", err.Error())
			return
		}

		msgtype, data, err := p.UnpackJSON(msg)
		if err != nil {
			p.Log.Println("UnpackJSON failed.", err.Error())
			return
		}
		p.Log.Printf("UnpackJSON success, msgLen = %d, msgtype = %d\n", msgLen, msgtype)

		v, ok := M[msgtype]
		if !ok {
			p.Log.Println("Invalid msgtype:", msgtype)
			return
		}

		err = v(conn, data)
		if err != nil {
			p.Log.Println("Process failed:", err)
			return
		}
	}
}

func ProcessInitDirectory(conn net.Conn, data []byte) error {
	p.Log.Printf("ProcessInitDirectory\n")

	msg := &syncdirectory.MInitDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		p.Log.Println("Unmarshal MInitDirectory failed")
		return err
	}

	p.Log.Println(msg)

	newRoot := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot()
	p.Log.Println(newRoot)

	exists, _ := p.PathExists(newRoot)
	if exists {
		os.RemoveAll(newRoot)
	}

	if err := os.Mkdir(newRoot, os.ModePerm); err != nil {
		p.Log.Println("Mkdir failed", newRoot)
		return err
	}

	p.Log.Printf("Mkdir [%s] successfully.\n", newRoot)

	return nil
}

func ProcessPushDirectory(conn net.Conn, data []byte) error {
	p.Log.Println("ProcessPushDirectory")

	msg := &syncdirectory.MPushDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		p.Log.Println("Unmarshal MPushDirectory failed")
		return err
	}

	p.Log.Println(msg.GetRoot(), msg.GetDirname(), msg.GetSubdirname(), msg.GetSubfilename())

	path := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetDirname()
	if exists, _ := p.PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			p.Log.Println("Mkdir failed", path)
			return err
		}
	}

	return nil
}

func ProcessPushFile(conn net.Conn, data []byte) error {
	p.Log.Println("ProcessPushFile")

	push := &syncdirectory.MPushFile{}
	err := proto.Unmarshal(data, push)
	if err != nil {
		p.Log.Println("Unmarshal MPushFile failed")
		return err
	}

	p.Log.Println(push.GetRoot(), push.GetFileName(), push.GetFileSize(), push.GetFileDir())

	path := STORE_LOCATION + string(os.PathSeparator) + push.GetRoot()
	if len(push.GetFileDir()) != 0 {
		path = path + string(os.PathSeparator) + push.GetFileDir()
	}
	fileWithPath := path + string(os.PathSeparator) + push.GetFileName()
	p.Log.Println("New file path", fileWithPath)

	if exists, _ := p.PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			p.Log.Println("Mkdir failed", path)
			return err
		}
	}

	f, err := os.OpenFile(fileWithPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		p.Log.Println(err)
		return err
	}
	defer f.Close()

	io.CopyN(f, conn, push.GetFileSize())

	return nil
}

func ProcessDeleteFile(conn net.Conn, data []byte) error {
	p.Log.Println("ProcessDeleteFile")

	msg := &syncdirectory.MDeleteFile{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		p.Log.Println("Unmarshal MDeleteFile failed")
		return err
	}

	p.Log.Println(msg.GetRoot(), msg.GetRelativeFileWithPath())

	fileWithPath := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetRelativeFileWithPath()
	p.Log.Println(fileWithPath)

	err = os.RemoveAll(fileWithPath)
	if err != nil {
		p.Log.Println("Delete file failed:", fileWithPath)
		return err
	}

	p.Log.Printf("Delete file %s success.\n", fileWithPath)

	return nil
}

func ProcessMoveFile(conn net.Conn, data []byte) error {
	p.Log.Println("ProcessMoveFile")

	msg := &syncdirectory.MMoveFile{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		p.Log.Println("Unmarshal MPushFile failed")
		return err
	}

	p.Log.Println(msg.GetRoot(), msg.GetOldFileWithPath(), msg.GetNewFileWithPath())

	old := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetOldFileWithPath()
	new := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetNewFileWithPath()

	err = os.Rename(old, new)
	if err != nil {
		p.Log.Println("os.Rename failed")
		return err
	}

	p.Log.Printf("Move file from %s to %s success.\n", old, new)

	return nil
}

func ProcessPullDirectoryRequest(conn net.Conn, data []byte) error {
	p.Log.Println("ProcessPullDirectoryRequest")

	msg := &syncdirectory.MPullDirectoryRequest{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		p.Log.Println("Unmarshal MPushFile failed")
		return err
	}

	p.Log.Println(msg.GetRoot())

	path := STORE_LOCATION + string(os.PathSeparator) + msg.GetRoot()
	exists, err := p.PathExists(path)
	if !exists {
		//TODO send response to client: root is not exists.
		return err
	}

	return nil
}
