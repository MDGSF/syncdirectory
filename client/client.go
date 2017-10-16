package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syncdirectory"
	"syncdirectory/client/notifyDir"
	p "syncdirectory/public"

	"github.com/golang/protobuf/proto"
)

func main() {
	p.Log.Println("client start")

	sendInitToServer()

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
	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		p.Log.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	path := p.GetFilePath(event.Name)
	name := p.GetFileName(event.Name)
	fmt.Println(path)

	pushFileToServer(conn, path, name)
}

func notifyRemoved(event notifyDir.NotifyEvent) {
	if p.IsDir(event.Name) {
		notifyDirectoryDeleted(event)
	} else {
		notifyFileDeleted(event)
	}
}

func notifyDirectoryDeleted(event notifyDir.NotifyEvent) {
	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		fmt.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &syncdirectory.MDeleteDirectory{}
	msg.Root = proto.String(notifyDir.ROOT)
	msg.FileName = proto.String(p.GetFileName(event.Name))
	msg.FilePath = proto.String(p.GetFilePath(event.Name))
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EDeleteDirectory), msg)
}

func notifyFileDeleted(event notifyDir.NotifyEvent) {
	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		fmt.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &syncdirectory.MDeleteFile{}
	msg.Root = proto.String(notifyDir.ROOT)
	msg.FileName = proto.String(p.GetFileName(event.Name))
	msg.FilePath = proto.String(p.GetFilePath(event.Name))
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EDeleteFile), msg)
}

func notifyRenamed(event notifyDir.NotifyEvent) {
	fmt.Println("notifyRenamed")
}

func sendInitToServer() {
	conn, err := net.Dial(p.CONN_TYPE, p.CONN_HOST+":"+p.CONN_PORT)
	if err != nil {
		fmt.Println("Error dialing", err.Error())
		return
	}
	defer conn.Close()

	msg := &syncdirectory.MInitDirectory{}
	msg.Root = proto.String(notifyDir.ROOT)
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EInitDirectory), msg)

	pushDirectory(conn, notifyDir.DIR_NAME)

	fmt.Println("sendInitToServer end")
}

func pushDirectory(conn net.Conn, path string) {
	fmt.Println("pushDirectory:", path)

	dir, err := os.Open(path)
	if err != nil {
		fmt.Println("os.Open failed", err.Error())
		return
	}
	defer dir.Close()

	msg := &syncdirectory.MPushDirectory{}
	msg.Root = proto.String(notifyDir.ROOT)
	msg.Dirname = proto.String(getRelativePath(notifyDir.DIR_NAME, path))

	names, err := dir.Readdirnames(-1)
	if err != nil {
		fmt.Println("dir.Readdirnames failed", err.Error())
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
			pushFileToServer(conn, path, name)
		}
	}
}

func pushFileToServer(conn net.Conn, path string, filename string) {
	fileWithPath := path + "\\" + filename
	msg := &syncdirectory.MPushFile{}
	msg.Root = proto.String(notifyDir.ROOT)
	msg.FileName = proto.String(filename)
	msg.FileSize = proto.Int64(p.FileSize(fileWithPath))
	msg.FileDir = proto.String(getRelativePath(notifyDir.DIR_NAME, path))
	p.SendMsg(conn, int(syncdirectory.ESyncMsgCode_EPushFile), msg)

	f, err := os.Open(fileWithPath)
	if err != nil {
		fmt.Println("err opening file", fileWithPath)
		return
	}
	defer f.Close()

	fileInfo, err := os.Lstat(fileWithPath)
	if err != nil {
		fmt.Println("err Lstat")
		return
	}
	//fmt.Println(fileInfo.Size())

	written, err := io.CopyN(conn, f, fileInfo.Size())
	if written != fileInfo.Size() || err != nil {
		fmt.Println("error copy")
		return
	}
}

func getRelativePath(root, absolutePath string) string {
	if strings.HasPrefix(absolutePath, root) {
		//fmt.Println("has prefix")
		if len(absolutePath) > len(root) {
			return absolutePath[len(root)+1:]
		} else if len(absolutePath) == len(root) {
			return ""
		}
	} else {
		//fmt.Println("has not prefix")
	}
	return ""
}
