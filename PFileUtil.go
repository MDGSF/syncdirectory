package syncdirectory

import (
	"os"
	"strings"
)

/*
PathExists : check the path exists or not.
*/
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

/*
IsDir : check the path is dir or not.
*/
func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		Log.Println("os.Stat failed", err.Error())
		return false
	}
	return fileInfo.IsDir()
}

/*
FileSize : get the file size.
*/
func FileSize(fileWithPath string) int64 {
	fileInfo, err := os.Stat(fileWithPath)
	if err != nil {
		Log.Println("os.Stat failed", err.Error())
		return 0
	}
	return fileInfo.Size()
}

/*
GetFilePath : get the absolute path.
*/
func GetFilePath(fileWithPath string) string {
	i := strings.LastIndex(fileWithPath, "\\")
	if i == -1 {
		return ""
	}
	return fileWithPath[:i]
}

/*
GetFileName : get the file name.
*/
func GetFileName(fileWithPath string) string {
	i := strings.LastIndex(fileWithPath, "\\")
	if i == -1 {
		return fileWithPath
	}
	return fileWithPath[i+1:]
}
