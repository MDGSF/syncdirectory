package notifyDir

import (
	"fmt"
	"log"
	"os"
	"syncdirectory/public"
	"time"

	"github.com/fsnotify"
)

type NotifyEvent struct {
	EventType uint32
	Name      string
	Time      time.Time
	NewName   string
}

func (t NotifyEvent) Changed() bool {
	if fsnotify.Op(t.EventType)&fsnotify.Create == fsnotify.Create {
		return true
	}

	if fsnotify.Op(t.EventType)&fsnotify.Write == fsnotify.Write {
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

const (
	ROOT     = "fsnotify_demo"
	DIR_NAME = "E:\\fsnotify_demo"
)

var watcher *fsnotify.Watcher

/*
StartNotify start notify directory.
*/
func StartNotify(eventChan chan NotifyEvent) {
	raweventChan := make(chan fsnotify.Event)
	dupeventChan := make(chan NotifyEvent)
	go runFsnotify(raweventChan)
	go processRawEvent(raweventChan, dupeventChan)
	go processDupEvent(dupeventChan, eventChan)
}

func runFsnotify(rawevent chan fsnotify.Event) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("fsnotify.NewWatcher failed", err.Error())
		return
	}
	defer watcher.Close()

	browserDir(DIR_NAME)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				//log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Create file:", event.Name)
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Write file:", event.Name)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("Remove file:", event.Name)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("Rename file:", event.Name)
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					log.Println("Chmod file:", event.Name)
				}
				rawevent <- event
			case err := <-watcher.Errors:
				log.Println("error:", err)
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
		eventChan <- event
	}
}

func processDupEvent(dupeventChan chan NotifyEvent, eventChan chan NotifyEvent) {
	//preEvent := NotifyEvent{}
	//for dupevent := range dupeventChan {
	//	if !dupevent.Equal(preEvent) {
	//		fmt.Println("pre", preEvent)
	//		fmt.Println("dup", dupevent)
	//		preEvent = dupevent
	//		eventChan <- dupevent
	//	}
	//}

	preEvent := NotifyEvent{}
	for {
		dupevent, ok := <-dupeventChan
		if !ok {
			continue
		}

		if fsnotify.Op(dupevent.EventType)&fsnotify.Rename == fsnotify.Rename {
			next, ok := <-dupeventChan
			if !ok {
				continue
			}

			if !dupevent.TimeEqual(next) {
				continue
			}

			if fsnotify.Op(next.EventType)&fsnotify.Create != fsnotify.Create {
				continue
			}

			dupevent.NewName = next.Name

			next, ok = <-dupeventChan
			if !ok {
				continue
			}

			if !dupevent.TimeEqual(next) {
				continue
			}

			if fsnotify.Op(next.EventType)&fsnotify.Write != fsnotify.Write {
				continue
			}

			preEvent = dupevent
			eventChan <- dupevent
		} else {
			if !dupevent.Equal(preEvent) {
				fmt.Println("pre", preEvent)
				fmt.Println("dup", dupevent)
				preEvent = dupevent
				eventChan <- dupevent
			}
		}
	}
}

func watchDir(path string) {
	fmt.Println("watchDir:", path)
	err := watcher.Add(path)
	if err != nil {
		fmt.Println("watcher.Add failed", err.Error())
		return
	}
}

func browserDir(path string) {
	fmt.Println("browserDir:", path)

	dir, err := os.Open(path)
	if err != nil {
		fmt.Println("os.Open failed", err.Error())
		return
	}
	defer dir.Close()

	watchDir(path)

	names, err := dir.Readdirnames(-1)
	if err != nil {
		fmt.Println("dir.Readdirnames failed", err.Error())
		return
	}

	for _, name := range names {
		sub := path + "\\" + name
		if !public.IsDir(sub) {
			continue
		}
		browserDir(sub)
	}
}
