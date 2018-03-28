package main

import "github.com/gin-gonic/gin"
import "dockerhub-deploy/controllers"

func main() {
	engine := gin.Default()

	engine.POST("/deploy",controllers.Deploy)

	engine.Run(":9500")
}
