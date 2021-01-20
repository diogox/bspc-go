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
	defer func() {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}()

	var st bspc.State
	if err := c.Query("wm --dump-state", bspc.ToStruct(&st)); err != nil {
		panic(err)
	}
	fmt.Println(st)

	var nodes []bspc.ID
	if err := c.Query("query -d focused -N", bspc.ToIDSlice(&nodes)); err != nil {
		panic(err)
	}
	fmt.Println(nodes)
}
