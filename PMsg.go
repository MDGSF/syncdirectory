package syncdirectory

import (
	"io"
	"net"
	"os"

	proto "github.com/golang/protobuf/proto"
)

func PushFileSend(conn net.Conn, AbsoluteFileWithPath string, msg *MPushFile) error {
	f, err := os.Open(AbsoluteFileWithPath)
	if err != nil {
		Log.Println("err opening file", AbsoluteFileWithPath)
		return err
	}
	defer f.Close()

	fileInfo, err := os.Lstat(AbsoluteFileWithPath)
	if err != nil {
		Log.Println("err Lstat")
		return err
	}

	SendMsg(conn, int(ESyncMsgCode_EPushFile), msg)

	if fileInfo.Size() > 0 {
		written, err := io.CopyN(conn, f, fileInfo.Size())
		if written != fileInfo.Size() || err != nil {
			Log.Println("error copy")
			return err
		}
	}

	return nil
}

func PushFileRecv(conn net.Conn, data []byte, StoreLocation string) error {
	push := &MPushFile{}
	err := proto.Unmarshal(data, push)
	if err != nil {
		Log.Println("Unmarshal MPushFile failed")
		return err
	}

	Log.Println(push.GetRoot(), push.GetFileName(), push.GetFileSize(), push.GetRelativePath())

	path := StoreLocation + string(os.PathSeparator) + push.GetRoot()
	if len(push.GetRelativePath()) != 0 {
		path = path + string(os.PathSeparator) + push.GetRelativePath()
	}
	fileWithPath := path + string(os.PathSeparator) + push.GetFileName()
	Log.Println("New file path", fileWithPath)

	if exists, _ := PathExists(path); !exists {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			Log.Println("Mkdir failed", path)
			return err
		}
	}

	f, err := os.OpenFile(fileWithPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		Log.Println(err)
		return err
	}
	defer f.Close()

	io.CopyN(f, conn, push.GetFileSize())

	return nil
}

func PushDirectory(conn net.Conn, RootName string, path string) {
	Log.Println("PushDirectory:", path)

	dir, err := os.Open(path)
	if err != nil {
		Log.Println("os.Open failed", err.Error())
		return
	}
	defer dir.Close()

	msg := &MPushDirectory{}
	msg.Root = proto.String(RootName)
	msg.Dirname = proto.String(GetRelativePath(path, RootName))

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
			PushDirectory(conn, RootName, sub)
		} else {
			file, _ := CreateEventFile(sub, RootName)

			msg := &MPushFile{}
			msg.Root = proto.String(RootName)
			msg.FileName = proto.String(file.FileName)
			msg.FileSize = proto.Int64(file.FileSize)
			msg.RelativePath = proto.String(file.RelativePath)

			PushFileSend(conn, file.AbsoluteFileWithPath, msg)
		}
	}
}
