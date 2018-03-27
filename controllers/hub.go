package controllers

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"os/exec"
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
	fmt.Println(secretKey)

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
	cmd.Dir = "/opt/compose/" + path
	fmt.Println("cd /opt/compose/" + path)
	fmt.Println("docker-compose pull")

	//cmd.Stdout = &out
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("docker pull error:", err.Error())
		c.JSON(http.StatusOK, gin.H{"status": "success"})
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
		c.JSON(http.StatusOK, gin.H{"status": "success"})
		return
	}
	fmt.Println(string(out))

	//cmd.Stdout = &out
	//err = cmd.Run()
	//if err != nil {
	//	fmt.Printf(err.Error())
	//}
	//fmt.Printf("GOGOGO: %q\n", out.String())

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
