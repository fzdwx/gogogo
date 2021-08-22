package go_concurrent

import "fmt"

// generateNatural 生成最初的2, 3, 4, ...自然数序列（不包含开头的0、1）：
func generateNatural() chan int {
	ch := make(chan int)

	go func() {
		for i := 2; ; i++ {
			ch <- i
		}
	}()

	return ch
}

// primeFilter 管道过滤器: 删除能被素数整除的数
func primeFilter(in chan int, prime int) chan int {
	out := make(chan int)

	go func() {
		for {
			if i := <-in; i%prime != 0 {
				out <- i
			}
		}
	}()
	return out
}

func RunPrimesieve() {
	ch := generateNatural() // 自然数序列: 2, 3, 4, ...
	for i := 0; i < 100; i++ {
		prime := <-ch // 新出现的素数
		fmt.Printf("%v: %v\n", i+1, prime)
		ch = primeFilter(ch, prime) // 基于新素数构造的过滤器
	}
}
