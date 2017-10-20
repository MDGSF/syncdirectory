package public

import (
	"net"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "10001"
	CONN_TYPE = "tcp"
	BUF_SIZE  = 4 * 1024
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, err
	}
	return false, err
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		Log.Println("os.Stat failed", err.Error())
		return false
	}
	return fileInfo.IsDir()
}

func FileSize(fileWithPath string) int64 {
	fileInfo, err := os.Stat(fileWithPath)
	if err != nil {
		Log.Println("os.Stat failed", err.Error())
		return 0
	}
	return fileInfo.Size()
}

func GetFilePath(fileWithPath string) string {
	i := strings.LastIndex(fileWithPath, "\\")
	if i == -1 {
		return ""
	}
	return fileWithPath[:i]
}

func GetFileName(fileWithPath string) string {
	i := strings.LastIndex(fileWithPath, "\\")
	if i == -1 {
		return fileWithPath
	}
	return fileWithPath[i+1:]
}

func SendMsg(conn net.Conn, msgCode int, msg proto.Message) {
	protob, err := proto.Marshal(msg)
	if err != nil {
		Log.Println("marshal MInitDirectory failed")
		return
	}

	b, err := PackToJSON(msgCode, []byte(protob))
	if err != nil {
		Log.Println("pack to json MInitDirectory failed")
		return
	}

	Write(conn, string(b))
}
