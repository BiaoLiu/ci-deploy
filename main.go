package main

import (
	"ci-deploy/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	engine.POST("/dockerhub-deploy", controllers.DockerhubDeploy)
	engine.GET("/deploy", controllers.Deploy)

	engine.Run(":8090")
}
