package syncdirectory

import (
	"net"
)

/*
M: function map
key[int]: message type.
value[func(net.Conn, []byte)]: proccess function.
*/
var M map[int]func(net.Conn, []byte) error

func Register(msgtype int, f func(net.Conn, []byte) error) {
	M[msgtype] = f
}

func init() {
	M = make(map[int]func(net.Conn, []byte) error)
	Register(int(ESyncMsgCode_EInitDirectory), ProcessInitDirectory)
	Register(int(ESyncMsgCode_EPushDirectory), ProcessPushDirectory)
	Register(int(ESyncMsgCode_EPushFile), ProcessPushFile)
	Register(int(ESyncMsgCode_EDeleteFile), ProcessDeleteFile)
	Register(int(ESyncMsgCode_EMoveFile), ProcessMoveFile)
	Register(int(ESyncMsgCode_EPullDirectoryRequest), ProcessPullDirectoryRequest)
}
