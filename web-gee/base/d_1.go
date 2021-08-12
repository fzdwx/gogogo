package main

import (
	"fmt"
	"net/http"
)

/*
基础版本：
	1.启动一个服务监听80端口
	2.处理/ 以及 /hello
*/

func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}
