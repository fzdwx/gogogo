package cgo

import "C"

/*
#include <hello.h>
*/
import "C"
import "fmt"

func F1() {
	C.SayHello(C.CString("hello world\n"))
}

//export SayHello
func SayHello(s *C.char) {
	fmt.Println("123")
	fmt.Print(C.GoString(s))
}
