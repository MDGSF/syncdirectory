package main

import (
	"flag"
	"io"
	"net"
	"os"
	"syncdirectory"
	"syncdirectory/client/notifyDir"
	p "syncdirectory/public"

	"github.com/golang/protobuf/proto"
)

var firstInit = flag.Bool("firstInit", false, "first time init directory.")

func main() {
	flag.Parse()

	p.InitLog("client.log")
	p.Log.Println("client start")

	if *firstInit {
		sendInitToServer()
	}

	events := make(chan notifyDir.NotifyEvent)
	notifyDir.StartNotify(events)

	for event := range events {
		p.Log.Println("main", event.EventType, event.Name)
		if event.Changed() {
			notifyChanged(event)
		} else if event.Removed() {
			notifyRemoved(event)
		} else if event.Renamed() {
			notifyRenamed(event)
		} else {
			p.Log.Println("Unknown event")
		}
	}

	done := make(chan bool)
	<-done
}

func notifyChanged(event notifyDir.NotifyEvent) {
	p.Log.Println("notifyChanged")

	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		p.Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	if event.File.IsDir {
		pushNewDirectoryToServer(conn, event)
	} else {
		pushFileToServer(conn, event.File)
	}
}

func notifyRemoved(event notifyDir.NotifyEvent) {
	p.Log.Println("notifyFileDeleted")

	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		p.Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &syncdirectory.MDeleteFile{}
	msg.Root = proto.String(notifyDir.ROOT)
	msg.RelativeFileWithPath = proto.String(event.File.RelativeFileWithPath)
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EDeleteFile), msg)
}

func notifyRenamed(event notifyDir.NotifyEvent) {
	p.Log.Println("notifyRenamed:", event.Name, event.NewName)

	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		p.Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	old, _ := notifyDir.CreateEventFile(event.Name)
	new, _ := notifyDir.CreateEventFile(event.NewName)

	msg := &syncdirectory.MMoveFile{}
	msg.Root = proto.String(notifyDir.ROOT)
	msg.OldFileWithPath = proto.String(old.RelativeFileWithPath)
	msg.NewFileWithPath = proto.String(new.RelativeFileWithPath)
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EMoveFile), msg)
}

func sendInitToServer() {
	p.Log.Println("sendInitToServer")

	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		p.Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &syncdirectory.MInitDirectory{}
	msg.Root = proto.String(notifyDir.ROOT)
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EInitDirectory), msg)

	pushDirectory(conn, notifyDir.DIR_NAME)

	p.Log.Printf("sendInitToServer success\n\n")
}

func pushDirectory(conn net.Conn, path string) {
	p.Log.Println("pushDirectory:", path)

	dir, err := os.Open(path)
	if err != nil {
		p.Log.Println("os.Open failed", err.Error())
		return
	}
	defer dir.Close()

	msg := &syncdirectory.MPushDirectory{}
	msg.Root = proto.String(notifyDir.ROOT)
	msg.Dirname = proto.String(notifyDir.GetRelativePath(path))

	names, err := dir.Readdirnames(-1)
	if err != nil {
		p.Log.Println("dir.Readdirnames failed", err.Error())
		return
	}

	for _, name := range names {
		sub := path + "\\" + name
		if p.IsDir(sub) {
			msg.Subdirname = append(msg.Subdirname, sub)
		} else {
			msg.Subfilename = append(msg.Subfilename, sub)
		}
	}

	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EPushDirectory), msg)

	for _, name := range names {
		sub := path + "\\" + name
		if p.IsDir(sub) {
			pushDirectory(conn, sub)
		} else {
			file, _ := notifyDir.CreateEventFile(sub)
			pushFileToServer(conn, file)
		}
	}
}

func pushNewDirectoryToServer(conn net.Conn, event notifyDir.NotifyEvent) {
	p.Log.Println("pushNewDirectoryToServer", event.File)

	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		p.Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &syncdirectory.MPushDirectory{}
	msg.Root = proto.String(event.File.Root)
	msg.Dirname = proto.String(event.File.RelativeFileWithPath)

	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EPushDirectory), msg)
}

func pushFileToServer(conn net.Conn, file *notifyDir.SEventFile) {
	p.Log.Println("pushFileToServer", file)

	f, err := os.Open(file.AbsoluteFileWithPath)
	if err != nil {
		p.Log.Println("err opening file", file.AbsoluteFileWithPath)
		return
	}
	defer f.Close()

	msg := &syncdirectory.MPushFile{}
	msg.Root = proto.String(file.Root)
	msg.FileName = proto.String(file.FileName)
	msg.FileSize = proto.Int64(file.FileSize)
	msg.FileDir = proto.String(file.RelativePath)
	msg.RelativeFileWithPath = proto.String(file.RelativeFileWithPath)
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EPushFile), msg)

	fileInfo, err := os.Lstat(file.AbsoluteFileWithPath)
	if err != nil {
		p.Log.Println("err Lstat")
		return
	}

	if fileInfo.Size() > 0 {
		written, err := io.CopyN(conn, f, fileInfo.Size())
		if written != fileInfo.Size() || err != nil {
			p.Log.Println("error copy")
			return
		}
	}
}
