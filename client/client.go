package main

import (
	"fmt"
	"syncdirectory/client/notifyDir"
)

func main() {
	fmt.Println("client start")
	events := make(chan notifyDir.NotifyEvent)
	notifyDir.StartNotify(events)

	for event := range events {
		fmt.Println("main", event.EventType, event.Name)
	}

	done := make(chan bool)
	<-done
}
