package syncdirectory

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/golang/protobuf/proto"
)

var firstInit = flag.Bool("firstInit", false, "First time init directory.")
var pullAllFromServer = flag.Bool("pullAllFromServer", false, "Drop local file, and pull all file from server.")

func checkFlag() {
	flag.Parse()
	if *firstInit && *pullAllFromServer {
		fmt.Println("firstInit and pullAllFromServer can't be set at the same time.")
	}
}

func StartClient() {
	checkFlag()

	InitLog("client.log")
	Log.Println("client start")

	if *firstInit {
		sendInitToServer()
	} else if *pullAllFromServer {
		pullDirectoryFromServer()
	}

	events := make(chan NotifyEvent)
	StartNotify(events)

	for event := range events {
		Log.Println("main", event.EventType, event.Name)
		if event.Changed() {
			notifyChanged(event)
		} else if event.Removed() {
			notifyRemoved(event)
		} else if event.Renamed() {
			notifyRenamed(event)
		} else {
			Log.Println("Unknown event")
		}
	}

	done := make(chan bool)
	<-done
}

func notifyChanged(event NotifyEvent) {
	Log.Println("notifyChanged")

	conn, err := net.Dial(ConnectionType, ServerHost+":"+ServerPort)
	if err != nil {
		Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	if event.File.IsDir {
		pushNewDirectoryToServer(conn, event)
	} else {
		pushFileToServer(conn, event.File)
	}
}

func notifyRemoved(event NotifyEvent) {
	Log.Println("notifyFileDeleted")

	conn, err := net.Dial(ConnectionType, ServerHost+":"+ServerPort)
	if err != nil {
		Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &MDeleteFile{}
	msg.Root = proto.String(CRootName)
	msg.RelativeFileWithPath = proto.String(event.File.RelativeFileWithPath)
	SendMsg(conn, int(ESyncMsgCode_EDeleteFile), msg)
}

func notifyRenamed(event NotifyEvent) {
	Log.Println("notifyRenamed:", event.Name, event.NewName)

	conn, err := net.Dial(ConnectionType, ServerHost+":"+ServerPort)
	if err != nil {
		Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	old, _ := CreateEventFile(event.Name)
	new, _ := CreateEventFile(event.NewName)

	msg := &MMoveFile{}
	msg.Root = proto.String(CRootName)
	msg.OldFileWithPath = proto.String(old.RelativeFileWithPath)
	msg.NewFileWithPath = proto.String(new.RelativeFileWithPath)
	SendMsg(conn, int(ESyncMsgCode_EMoveFile), msg)
}

func sendInitToServer() {
	Log.Println("sendInitToServer")

	conn, err := net.Dial(ConnectionType, ServerHost+":"+ServerPort)
	if err != nil {
		Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &MInitDirectory{}
	msg.Root = proto.String(CRootName)
	SendMsg(conn, int(ESyncMsgCode_EInitDirectory), msg)

	pushDirectory(conn, CRootPath)

	Log.Printf("sendInitToServer success\n\n")
}

func pullDirectoryFromServer() {
	Log.Println("pullDirectoryFromServer")

	conn, err := net.Dial(ConnectionType, ServerHost+":"+ServerPort)
	if err != nil {
		Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &MPullDirectoryRequest{}
	msg.Root = proto.String(CRootName)
	SendMsg(conn, int(ESyncMsgCode_EPullDirectoryRequest), msg)
}

func pushDirectory(conn net.Conn, path string) {
	Log.Println("pushDirectory:", path)

	dir, err := os.Open(path)
	if err != nil {
		Log.Println("os.Open failed", err.Error())
		return
	}
	defer dir.Close()

	msg := &MPushDirectory{}
	msg.Root = proto.String(CRootName)
	msg.Dirname = proto.String(GetRelativePath(path))

	names, err := dir.Readdirnames(-1)
	if err != nil {
		Log.Println("dir.Readdirnames failed", err.Error())
		return
	}

	for _, name := range names {
		sub := path + "\\" + name
		if IsDir(sub) {
			msg.Subdirname = append(msg.Subdirname, sub)
		} else {
			msg.Subfilename = append(msg.Subfilename, sub)
		}
	}

	SendMsg(conn, int(ESyncMsgCode_EPushDirectory), msg)

	for _, name := range names {
		sub := path + "\\" + name
		if IsDir(sub) {
			pushDirectory(conn, sub)
		} else {
			file, _ := CreateEventFile(sub)
			pushFileToServer(conn, file)
		}
	}
}

func pushNewDirectoryToServer(conn net.Conn, event NotifyEvent) {
	Log.Println("pushNewDirectoryToServer", event.File)

	conn, err := net.Dial(ConnectionType, ServerHost+":"+ServerPort)
	if err != nil {
		Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &MPushDirectory{}
	msg.Root = proto.String(event.File.Root)
	msg.Dirname = proto.String(event.File.RelativeFileWithPath)

	SendMsg(conn, int(ESyncMsgCode_EPushDirectory), msg)
}

func pushFileToServer(conn net.Conn, file *SEventFile) {
	Log.Println("pushFileToServer", file)

	f, err := os.Open(file.AbsoluteFileWithPath)
	if err != nil {
		Log.Println("err opening file", file.AbsoluteFileWithPath)
		return
	}
	defer f.Close()

	msg := &MPushFile{}
	msg.Root = proto.String(file.Root)
	msg.FileName = proto.String(file.FileName)
	msg.FileSize = proto.Int64(file.FileSize)
	msg.FileDir = proto.String(file.RelativePath)
	msg.RelativeFileWithPath = proto.String(file.RelativeFileWithPath)
	SendMsg(conn, int(ESyncMsgCode_EPushFile), msg)

	fileInfo, err := os.Lstat(file.AbsoluteFileWithPath)
	if err != nil {
		Log.Println("err Lstat")
		return
	}

	if fileInfo.Size() > 0 {
		written, err := io.CopyN(conn, f, fileInfo.Size())
		if written != fileInfo.Size() || err != nil {
			Log.Println("error copy")
			return
		}
	}
}
