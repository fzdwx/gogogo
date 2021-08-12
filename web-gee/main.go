package main

import (
	"fmt"
	"net/http"

	"gee"
)

func main() {

	engine := gee.New()

	engine.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello this index.html")
	}).GET("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	}).Run(":80")
}
