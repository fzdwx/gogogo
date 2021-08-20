package main

import (
	"encoding/json"
	"fmt"
)

func main() {

	channel := make(chan int)

	go func() {
		fmt.Println("Hello World")
		channel <- 1
	}()

	i := <-channel
	fmt.Println(i)
}

func Jsonify(raw interface{}) string {
	bytes, _ := json.Marshal(raw)
	return string(bytes)
}
