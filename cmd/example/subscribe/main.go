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

	eventCh, errCh, err := c.SubscribeEvents(bspc.EventTypeNodeRemove, bspc.EventTypeDesktopLayout)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case err := <-errCh:
			panic(err)
		case ev := <-eventCh:
			switch ev.Type {
			case bspc.EventTypeDesktopLayout:
				ev := ev.Payload.(bspc.EventDesktopLayout)
				fmt.Println("Layout Changed: ", ev.DesktopID)
			case bspc.EventTypeNodeRemove:
				ev := ev.Payload.(bspc.EventNodeRemove)
				fmt.Println("Node Removed: ", ev.NodeID)
			}
		}
	}
}
