package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func funcName() {
	m := make(map[string]int)

	data, _ := ioutil.ReadFile("ch1/data.txt")
	for _, line := range strings.Split(string(data), "\n") {
		m[line]++
	}

	for k, v := range m {
		if v > 1 {
			fmt.Printf("%d\t%s\n", v, k)
		}
	}
}
