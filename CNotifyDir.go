package syncdirectory

import (
	"os"
	"strings"
	"time"

	"github.com/fsnotify"
)

type NotifyEvent struct {
	EventType uint32
	Name      string
	Time      time.Time
	NewName   string
	File      *SEventFile
}

func (t NotifyEvent) Changed() bool {
	if fsnotify.Op(t.EventType)&fsnotify.Create == fsnotify.Create {
		return true
	}

	if (fsnotify.Op(t.EventType)&fsnotify.Write == fsnotify.Write) && !IsDir(t.Name) {
		return true
	}

	return false
}

func (t NotifyEvent) Removed() bool {
	if fsnotify.Op(t.EventType)&fsnotify.Remove == fsnotify.Remove {
		return true
	}
	return false
}

func (t NotifyEvent) Renamed() bool {
	if fsnotify.Op(t.EventType)&fsnotify.Rename == fsnotify.Rename {
		return true
	}
	return false
}

func (t NotifyEvent) Equal(u NotifyEvent) bool {
	if t.EventType != u.EventType {
		return false
	}

	if t.Name != u.Name {
		return false
	}

	if !t.TimeEqual(u) {
		return false
	}

	return true
}

func (t NotifyEvent) TimeEqual(u NotifyEvent) bool {
	if t.Time.Year() != u.Time.Year() {
		return false
	}

	if t.Time.Month() != u.Time.Month() {
		return false
	}

	if t.Time.Day() != u.Time.Day() {
		return false
	}

	if t.Time.Hour() != u.Time.Hour() {
		return false
	}

	if t.Time.Minute() != u.Time.Minute() {
		return false
	}

	if t.Time.Second() != u.Time.Second() {
		return false
	}

	return true
}

var watcher *fsnotify.Watcher

/*
StartNotify start notify directory.
*/
func StartNotify(eventChan chan NotifyEvent) {
	Log.Println("Start notify directory:", CRootPath)
	raweventChan := make(chan fsnotify.Event)
	duptimeeventChan := make(chan NotifyEvent)
	dupeventChan := make(chan NotifyEvent)

	go runFsnotify(raweventChan)
	go processRawEvent(raweventChan, duptimeeventChan)
	go processTheSameSecondDupPacket(duptimeeventChan, dupeventChan)
	go processDupEvent(dupeventChan, eventChan)
}

func runFsnotify(rawevent chan fsnotify.Event) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		Log.Println("fsnotify.NewWatcher failed", err.Error())
		return
	}
	defer watcher.Close()

	browserDir(CRootPath)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				//Log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					Log.Println("Create file:", event.Name)
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					Log.Println("Write file:", event.Name)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					Log.Println("Remove file:", event.Name)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					Log.Println("Rename file:", event.Name)
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					Log.Println("Chmod file:", event.Name)
				}
				rawevent <- event
			case err := <-watcher.Errors:
				Log.Println("error:", err)
			}
		}
	}()

	<-done
}

func processRawEvent(raweventChan chan fsnotify.Event, eventChan chan NotifyEvent) {
	for rawevent := range raweventChan {
		event := NotifyEvent{}
		event.EventType = uint32(rawevent.Op)
		event.Name = rawevent.Name
		event.Time = time.Now()

		var err error
		event.File, err = CreateEventFile(rawevent.Name)
		if err != nil {
			continue
		}

		eventChan <- event
	}
}

func processTheSameSecondDupPacket(duptimeEventChan chan NotifyEvent, timeEventChan chan NotifyEvent) {
	preEvent := NotifyEvent{}
	for duptimeevent := range duptimeEventChan {
		if !duptimeevent.Equal(preEvent) {
			preEvent = duptimeevent
			timeEventChan <- duptimeevent
		}
	}
}

func processDupEvent(dupeventChan chan NotifyEvent, eventChan chan NotifyEvent) {
	for {
		dupevent, ok := <-dupeventChan
		if !ok {
			continue
		}

		if fsnotify.Op(dupevent.EventType)&fsnotify.Rename == fsnotify.Rename {

			var next NotifyEvent
			to := time.NewTimer(time.Second)
			select {
			case next, ok = <-dupeventChan:
				if !ok {
					continue
				}
			case <-to.C:
				Log.Println("rename timeout")
				continue
			}

			if !dupevent.TimeEqual(next) {
				continue
			}

			if fsnotify.Op(next.EventType)&fsnotify.Create != fsnotify.Create {
				continue
			}

			dupevent.NewName = next.Name

			if IsDir(dupevent.NewName) {
				watchRenameDir(dupevent.NewName)
			}

			eventChan <- dupevent
			continue
		}
		eventChan <- dupevent
	}
}

/*
fsnotify 有个bug：
监控目录fsnotify_demo
子目录有fsnotify_demo\aaa\ccc.txt
rename aaa -> bbb
rename ccc.txt -> ddd.txt
会出错
因为重命名文件夹的时候，没有把旧的文件夹下的子文件，在内部的状态删除。
*/
func watchRenameDir(path string) {
	Log.Println("watchRenameDir:", path)
	err := watcher.Add(path)
	if err != nil {
		Log.Println("watcher.Add failed", err.Error())
	}

	err = watcher.Remove(path)
	if err != nil {
		Log.Println("watcher.Remove failed", err.Error())
		return
	}

	err = watcher.Add(path)
	if err != nil {
		Log.Println("watcher.Add failed", err.Error())
		return
	}

	if !IsDir(path) {
		return
	}

	dir, err := os.Open(path)
	if err != nil {
		Log.Println("os.Open failed", err.Error())
		return
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		Log.Println("dir.Readdirnames failed", err.Error())
		return
	}

	for _, name := range names {
		sub := path + "\\" + name
		watchRenameDir(sub)
	}
}

func watchDir(path string) {
	Log.Println("watchDir:", path)
	err := watcher.Add(path)
	if err != nil {
		Log.Println("watcher.Add failed", err.Error())
		return
	}
}

func reWatchDir(old string, new string) {
	Log.Println("reWatchDir from", old, "to", new)
	err := watcher.Remove(old)
	if err != nil {
		Log.Println("watcher.Remove failed", err.Error())
		return
	}

	err = watcher.Add(new)
	if err != nil {
		Log.Println("watcher.Add failed", err.Error())
		return
	}
}

func browserDir(path string) {
	Log.Println("browserDir:", path)

	dir, err := os.Open(path)
	if err != nil {
		Log.Println("os.Open failed", err.Error())
		return
	}
	defer dir.Close()

	watchDir(path)

	names, err := dir.Readdirnames(-1)
	if err != nil {
		Log.Println("dir.Readdirnames failed", err.Error())
		return
	}

	for _, name := range names {
		sub := path + "\\" + name
		if !IsDir(sub) {
			continue
		}
		browserDir(sub)
	}
}

func GetRelativePath(absolutePath string) string {
	if strings.HasPrefix(absolutePath, CRootPath) {
		//Log.Println("has prefix")
		if len(absolutePath) > len(CRootPath) {
			return absolutePath[len(CRootPath)+1:]
		} else if len(absolutePath) == len(CRootPath) {
			return ""
		}
	} else {
		//Log.Println("has not prefix")
	}
	return ""
}
