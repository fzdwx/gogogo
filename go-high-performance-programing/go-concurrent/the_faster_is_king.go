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
		ch <- SearchByBing()
	}()

	fmt.Println(<-ch)
}

func SearchByBing() int {
	return 1
}

func searchByGeogle() int {
	return 2
}

func searchByBaidu() int {
	return 3
}
