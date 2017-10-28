package syncdirectory

import (
	"errors"
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

type SEventFile struct {
	StoreLocation string

	AbsoluteFileWithPath string // E:\fsnotify_demo\aaa\test.txt
	AbsolutePath         string // E:\fsnotify_demo\aaa
	FileName             string // test.txt
	Root                 string // fsnotify_demo
	RelativeFileWithPath string // aaa\test.txt
	RelativePath         string // aaa

	FileSize int64 // size of file
	IsDir    bool
}

func CreateEventFile(absoluteFileWithPath string, RootName string) (*SEventFile, error) {

	exists, err := PathExists(absoluteFileWithPath)
	if err != nil || !exists {
		return nil, errors.New("not exists")
	}

	s := &SEventFile{}

	s.AbsoluteFileWithPath = absoluteFileWithPath
	s.AbsolutePath = GetFilePath(absoluteFileWithPath)
	s.FileName = GetFileName(absoluteFileWithPath)
	s.Root = RootName
	s.RelativeFileWithPath = GetRelativePath(absoluteFileWithPath, RootName)
	s.RelativePath = GetRelativePath(s.AbsolutePath, RootName)
	s.FileSize = FileSize(absoluteFileWithPath)

	s.IsDir = IsDir(absoluteFileWithPath)

	Log.Println(s)

	return s, nil
}

func GetRelativePath(absolutePath string, RootName string) string {

	i := strings.Index(absolutePath, RootName)
	if i != -1 {
		if len(absolutePath) > len(RootName) && (i+len(RootName)+1 < len(absolutePath)) {
			return absolutePath[i+len(RootName)+1:]
		}
	}
	return ""
}
