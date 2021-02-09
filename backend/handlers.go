package main

import (
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
)

func DefaultRoute(c *gin.Context) {
	time.Sleep(10 * time.Second)
	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}
