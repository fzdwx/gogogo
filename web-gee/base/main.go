package main

import (
	"log"
	"net/http"
)

func main() {
	d_2()
}

func d_2() {
	e := new(Engine)
	log.Fatal(http.ListenAndServe(":80", e))
}

func d_1() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}
