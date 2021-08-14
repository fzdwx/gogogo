package main

import (
	"fmt"
)

func main() {

	channel := make(chan string, 4)

	for i := 0; i < 10; i++ {
		i := i
		go func() {
			channel <- fmt.Sprint("qwe - ", i)
		}()
	}

	for {
		x, ok := <-channel
		if !ok {
			close(channel)
			return
		}
		fmt.Println(x)
	}
}
