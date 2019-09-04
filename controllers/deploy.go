package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os/exec"
)

/**Docker hub repo名称与服务器目录对应的map**/
var repoPathMapping1 = map[string]string{
	"crm": "pss-crm",
	"mjs": "pss-api",
}

const TOKEN1 = "eyJpYXQiOjE1M"

func Deploy(c *gin.Context) {
	token := c.Query("token")
	//项目名称
	repoName := c.Query("repo")

	fmt.Println("request url:", c.Request.URL)
	if token != TOKEN1 {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "msg": "token error"})
		return
	}
	if repoName == "" {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "msg": "repo name is required"})
		return
	}

	var dirName string
	if val, ok := repoPathMapping1[repoName]; ok {
		dirName = val
	} else {
		dirName = repoName
	}
	//服务器docker-compose目录
	path := "/opt/compose/" + dirName

	fmt.Println("prepare to exec: ", dirName)
	//time.Sleep(1 * time.Minute)

	cmd := exec.Command("docker-compose", "pull")
	cmd.Dir = path
	fmt.Println("exec script: docker-compose pull ", dirName)

	//cmd.Stdout = &out
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("docker pull error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": err.Error()})
		return
	}

	cmd = exec.Command("docker-compose", "up", "-d")
	cmd.Dir = path
	fmt.Println("exec script: docker-compose up -d")

	out, err = cmd.Output()
	if err != nil {
		fmt.Println("docker-compose up error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	fmt.Println(string(out))

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
