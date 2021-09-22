package main

import (
	"fmt"
)

func main() {

	channel := make(chan int)
	go func() {
		fmt.Println(<-channel)
	}()
	channel <- 2
}
