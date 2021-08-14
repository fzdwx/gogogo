package main

type FormatFunc func(string, ...interface{})

func test(formatFunc FormatFunc) {
	formatFunc("123", 123)
}
