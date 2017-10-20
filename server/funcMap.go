package server

import (
	"net"
	"syncdirectory"
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
	Register(int(syncdirectory.ESyncMsgCode_EInitDirectory), ProcessInitDirectory)
	Register(int(syncdirectory.ESyncMsgCode_EPushDirectory), ProcessPushDirectory)
	Register(int(syncdirectory.ESyncMsgCode_EPushFile), ProcessPushFile)
	Register(int(syncdirectory.ESyncMsgCode_EDeleteFile), ProcessDeleteFile)
	Register(int(syncdirectory.ESyncMsgCode_EMoveFile), ProcessMoveFile)
	Register(int(syncdirectory.ESyncMsgCode_EPullDirectoryRequest), ProcessPullDirectoryRequest)
}
