package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	myRouter := gin.Default()

	myRouter.GET("/", DefaultRoute)

	myRouter.Run(":8080")
}
