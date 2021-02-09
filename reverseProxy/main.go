package main

import (
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	myURL, err := url.Parse("http://localhost:8080")
	if err != nil {
		panic(err)
	}

	myProxy := ReverseProxy(*myURL, "")

	myRouter := gin.Default()

	myRouter.GET("/", myProxy)

	myRouter.Run(":8081")
}
