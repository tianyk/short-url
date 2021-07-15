package main

import (
    "github.com/gin-gonic/gin"

    "short-url/server"
    _ "short-url/utils"
)

var httpServer *gin.Engine

func main() {
    server.Run(httpServer)
}
