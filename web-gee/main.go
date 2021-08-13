package main

import (
	"net/http"

	"gee"
)

func main() {

	engine := gee.New()

	engine.GET("/", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"name": "like",
			"age":  "18",
		})
	})

}
