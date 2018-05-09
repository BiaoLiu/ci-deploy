package controllers

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"os/exec"
	"bytes"
	"time"
	"encoding/json"
)

type Webhook struct {
	PushData    PushData   `json:"push_data"`
	CallbackURL string     `json:"callback_url"`
	Repository  Repository `json:"repository"`
}

type PushData struct {
	Tag    string `json:"tag"`
	Pusher string `json:"pusher"`
}

type Repository struct {
	Name     string `json:"name"`
	RepoName string `json:"repo_name"`
}

type Callback struct {
	State       string `json:"state"`
	Context     string `json:"context"`
	Description string `json:"description"`
}

type TemplateData struct {
	Path string
	//RepoName string
	//Name     string
	//Tag      string
	//Params   string
}

/**Docker hub repo名称与服务器路径对应的map**/
var repoPathMapping = map[string]string{
	"ssoserver": "ssoserver",

	"roboshop-api":         "roboshop",
	"roboshop-front":       "roboshop",
	"personalcenter-front": "roboshop",

	"robouser-api":    "robouser-api",
	"robopay-api":     "robopay-api",
	"idgenerator-api": "idgenerator-api",

	"scmoperation-admin": "scm",
	"scmsupplier-admin":  "scm",
	"scmshop-api":        "scm",

	"customerdemand-front": "customerdemand",
	"customerdemand-admin": "customerdemand",
	"customerdemand-api":   "customerdemand",

	"super-admin": "super",
}

const TOKEN = "eyJpYXQiOjE1MjIxNDQ4NjAsInVpZCI6MSwic2lkIjoiOTJhYjlreXFoaWxiNDBscXl3cHAyeGxoeGg4d20yd2wifQ.DZun3A.i6sX5yTSJiJjm0xRCuAj_cw6-l0"

func Deploy(c *gin.Context) {
	token := c.Query("token")
	//对应服务器docker-compose目录名
	projectName := c.Query("repo")

	if token != TOKEN {
		c.JSON(http.StatusForbidden, gin.H{"msg": "token error"})
		return
	}

	var webhook Webhook
	if c.BindJSON(&webhook) != nil {
		fmt.Println("bind error")
	}

	var path string
	if projectName != "" {
		path = projectName
	} else {
		path = repoPathMapping[webhook.Repository.Name]
		if path == "" {
			path = webhook.Repository.Name
		}
	}

	path = "/opt/compose/" + path

	fmt.Println("prepare to exec: ", webhook.Repository.RepoName)
	time.Sleep(1 * time.Minute)

	cmd := exec.Command("docker-compose", "pull")
	cmd.Dir = path
	fmt.Println("exec script: docker-compose pull ", webhook.Repository.RepoName)

	//cmd.Stdout = &out
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("docker pull error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	fmt.Println(string(out))

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

	sendCallback(webhook.CallbackURL, true, "")

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func sendCallback(url string, success bool, description string) {
	body := Callback{
		State:       "failure",
		Context:     "Webhook deploy server",
		Description: description,
	}
	if len(body.Description) > 255 {
		body.Description = body.Description[0:255]
	}
	if success {
		body.State = "success"
	}
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(body)
	res, err := httpClient.Post(url, "application/json; charset=utf-8", buff)
	if err != nil || res.StatusCode != 200 {
		return
	}
}
