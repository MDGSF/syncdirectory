package syncdirectory

import (
	"io"
	"net"
	"os"

	proto "github.com/golang/protobuf/proto"
)

func PushFileSend(conn net.Conn, file *SEventFile) error {
	f, err := os.Open(file.AbsoluteFileWithPath)
	if err != nil {
		Log.Println("err opening file", file.AbsoluteFileWithPath)
		return err
	}
	defer f.Close()

	fileInfo, err := os.Lstat(file.AbsoluteFileWithPath)
	if err != nil {
		Log.Println("err Lstat")
		return err
	}

	msg := &MPushFile{}
	msg.Root = proto.String(file.Root)
	msg.FileName = proto.String(file.FileName)
	msg.FileSize = proto.Int64(file.FileSize)
	msg.RelativePath = proto.String(file.RelativePath)
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
