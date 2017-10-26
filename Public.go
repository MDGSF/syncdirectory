package syncdirectory

import (
	"os"
)

//Public constant.
const (

	//Server host name.
	ServerHost = "localhost"

	//Server port.
	ServerPort = "10001"

	//Client connect to server use tcp.
	ConnectionType = "tcp"
)

//Client constant.
const (

	//Client store directory location.
	CStoreLocation = "E:"

	//Client root path name.
	CRootName = "fsnotify_demo"

	//Client root absolute path.
	CRootPath = CStoreLocation + string(os.PathSeparator) + CRootName
)

//Server constant.
const (

	//Server store directory location.
	SStoreLocation = "E:\\ServerStore"
)
