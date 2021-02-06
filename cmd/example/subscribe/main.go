package main

import (
	"fmt"

	"github.com/diogox/bspc-go"
)

type logger struct{}

func (l logger) Info(msg string) {
	fmt.Println(msg)
}

func (l logger) Warn(msg string) {
	fmt.Println(msg)
}

func main() {
	c, err := bspc.New(logger{})
	if err != nil {
		panic(err)
	}

	eventCh, errCh, err := c.SubscribeEvents(bspc.EventTypeNodeAdd, bspc.EventTypeNodeRemove, bspc.EventTypePointerAction, bspc.EventTypeDesktopFocus)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case err := <-errCh:
			panic(err)
		case ev := <-eventCh:
			switch ev.Type {
			case bspc.EventTypeNodeAdd:
				ev := ev.Payload.(bspc.EventNodeAdd)
				fmt.Println("Node Added: ", ev.NodeID)
			case bspc.EventTypeNodeRemove:
				ev := ev.Payload.(bspc.EventNodeRemove)
				fmt.Println("Node Removed: ", ev.NodeID)
			case bspc.EventTypePointerAction:
				ev := ev.Payload.(bspc.EventPointerAction)
				fmt.Println("Pointer Action: ", ev.PointerActionState)
			case bspc.EventTypeDesktopFocus:
				ev := ev.Payload.(bspc.EventDesktopFocus)
				fmt.Println("Desktop Focused: ", ev.DesktopID)
			}
		}
	}
}
