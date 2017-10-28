package syncdirectory

import (
	"flag"
	"fmt"
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

	old, _ := CreateEventFile(event.Name, CRootName)
	new, _ := CreateEventFile(event.NewName, CRootName)

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

	PushDirectory(conn, CRootName, CRootPath)

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

		switch msgtype {
		case int(ESyncMsgCode_EPushDirectory):
			processPushDirectoryFromServer(conn, data)
		case int(ESyncMsgCode_EPushFile):
			PushFileRecv(conn, data, CStoreLocation)
		}
	}
}

func processPushDirectoryFromServer(conn net.Conn, data []byte) error {
	Log.Println("processPushDirectoryFromServer")

	msg := &MPushDirectory{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		Log.Println("Unmarshal MPushDirectory failed")
		return err
	}

	Log.Println(msg.GetRoot(), msg.GetDirname(), msg.GetSubdirname(), msg.GetSubfilename())

	path := CStoreLocation + string(os.PathSeparator) + msg.GetRoot() + string(os.PathSeparator) + msg.GetDirname()
	if exists, _ := PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			Log.Println("Mkdir failed", path)
			return err
		}
	}

	return nil
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

	msg := &MPushFile{}
	msg.Root = proto.String(CRootName)
	msg.FileName = proto.String(file.FileName)
	msg.FileSize = proto.Int64(file.FileSize)
	msg.RelativePath = proto.String(file.RelativePath)

	PushFileSend(conn, file.AbsoluteFileWithPath, msg)
}
