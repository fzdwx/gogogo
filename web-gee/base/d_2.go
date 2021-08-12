package main

import (
	"fmt"
	"net/http"
)

/*
	根据http.ListenAndServe(":80", nil)的第二个参数可以看到，传入的是一个handler
    我们实现了这个handler，然后针对路由进行处理
*/

type Engine struct {
}

func (engine Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
