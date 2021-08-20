package go_concurrent

import (
	"fmt"
	"sync"
)

/*
go 常见的并发模式 顺序执行

*/

func F1() {
	var mu sync.Mutex

	mu.Lock()

	go func() {
		fmt.Println("hello World")
		mu.Unlock()
	}()

	mu.Lock()
}

func F2() {
	channel := make(chan int, 1)

	go func() {
		fmt.Println("Hello World")
		channel <- 1
	}()

	<-channel
}

func F3() {
	done := make(chan int, 10) // 带 10 个缓存

	// 开N个后台打印线程
	for i := 0; i < cap(done); i++ {
		go func() {
			fmt.Println("你好, 世界")
			done <- 1
		}()
	}

	// 等待N个后台线程完成
	for i := 0; i < cap(done); i++ {
		<-done
	}
}
