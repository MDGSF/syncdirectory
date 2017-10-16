package public

import (
	"fmt"
	"log"
	"os"
)

var Log *log.Logger

func init() {
	logfile, err := os.OpenFile("test.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		//os.Exit(-1)
	}
	//defer logfile.Close()

	arr := []int{2, 3}

	Log = log.New(logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Log.Print("arr", arr, "\n")
	Log.Printf("arr[0] = %d", arr[0])
	Log.Println("hello")
	Log.Println("oh....")
	//Log.Fatal("test") //这个日志会直接让程序退出。
}
