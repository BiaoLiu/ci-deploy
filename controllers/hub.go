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

var repoPathMapping = map[string]string{
	"ssoserver": "ssoserver",

	"roboshop-api":   "roboshop",
	"roboshop-front": "roboshop",

	"robouser-api":    "robouser-api",
	"robopay-api":     "robopay-api",
	"idgenerator-api": "idgenerator-api",

	"scmoperation-admin": "scm",
	"scmsupplier-admin":  "scm",
	"scmshop-api":        "scm",

	"customerdemand-front": "customerdemand",
	"customerdemand-admin": "customerdemand",
	"customerdemand-api":   "customerdemand",
}

func Deploy(c *gin.Context) {
	secretKey := c.DefaultQuery("secretkey", "")

	if secretKey != "test" {
		c.JSON(http.StatusForbidden, gin.H{"msg": "secretkey错误"})
		return
	}

	var webhook Webhook
	if c.BindJSON(&webhook) != nil {
		fmt.Println("bind error")
	}

	path := repoPathMapping[webhook.Repository.Name]
	if path == "" {
		path = webhook.Repository.Name
	}

	path = "/opt/compose/" + path

	//var out bytes.Buffer

	cmd := exec.Command("docker-compose", "pull")
	cmd.Dir = path
	fmt.Println("docker-compose pull ", path)

	//cmd.Stdout = &out
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("docker pull error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	fmt.Println(string(out))

	//err = cmd.Run()
	//if err != nil {
	//	fmt.Printf(err.Error())
	//}
	//fmt.Printf("GOGOGO: %q\n", out.String())

	cmd = exec.Command("docker-compose", "up", "-d")
	cmd.Dir = path
	fmt.Println("docker-compose up -d")

	out, err = cmd.Output()
	if err != nil {
		fmt.Println("docker-compose up error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	fmt.Println(string(out))

	//cmd.Stdout = &out
	//err = cmd.Run()
	//if err != nil {
	//	fmt.Printf(err.Error())
	//}
	//fmt.Printf("GOGOGO: %q\n", out.String())

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
