package go_concurrent

import "fmt"

func ThefasteriskingDemo() {
	ch := make(chan int, 3)

	go func() {
		ch <- searchByBaidu()
	}()

	go func() {
		ch <- searchByGeogle()
	}()

	go func() {
		ch <- searchByBing()
	}()

	fmt.Println(<-ch)
}

func searchByBing() int {
	return 1
}

func searchByGeogle() int {
	return 2
}

func searchByBaidu() int {
	return 3
}
