package syncdirectory

import (
	"fmt"
	"log"
	"os"
)

/*
Log : global log to file.
*/
var Log *log.Logger

/*
InitLog : Init the global log.
name : the absolute path of the file which used to store the log.
*/
func InitLog(name string) {
	logfile, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}
	//defer logfile.Close()

	arr := []int{2, 3}

	Log = log.New(logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	Log.Print("arr", arr, "\n")
	Log.Printf("arr[0] = %d", arr[0])
	Log.Println("hello")
	Log.Println("oh....")
	Log.Printf("Log test end\n\n")
	//Log.Fatal("test") //Fatal level log will exit the program.
}
