package main

import (
	"fmt"
	"net/http"

	"gee"
)

func main() {

	engine := gee.Default()

	engine.GET("/", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"name": "like",
			"age":  "18",
		})
		fmt.Println("被调用")
	})
	engine.GET("/hello", func(context *gee.Context) {
		context.String(http.StatusOK, "hello")
	})
	engine.Run(":80")

}
