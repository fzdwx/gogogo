package go_concurrent

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// producer 生产者
func producer(factor int, out chan int) {
	for i := 0; ; i++ {
		out <- i * factor
	}
}

// consumer 消费者
func consumer(in chan int) {
	for v := range in {
		fmt.Println(v)
	}
}

func RunProducerAndConsumerDemo() {
	ch := make(chan int, 64)

	go producer(3, ch)
	go producer(5, ch)

	go consumer(ch)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("quit (%v)\n", <-sig)
}
