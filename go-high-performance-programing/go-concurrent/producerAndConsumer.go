package go_concurrent

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Producer 生产者
func Producer(factor int, out chan int) {
	for i := 0; ; i++ {
		out <- i * factor
	}
}

// Consumer 消费者
func Consumer(in chan int) {
	for v := range in {
		fmt.Println(v)
	}
}

func RunProducerAndConsumerDemo() {
	ch := make(chan int, 64)

	go Producer(3, ch)
	go Producer(5, ch)

	go Consumer(ch)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("quit (%v)\n", <-sig)
}
