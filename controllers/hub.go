package controllers

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"bytes"
	"os/exec"
	"text/template"
	"log"
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

var httpClient = &http.Client{Timeout: 10 * time.Second}
var deployScript = template.Must(template.ParseFiles("deploy.sh"))

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

	info := TemplateData{
		Path: "robo2025/" + path,
	}

	buff := new(bytes.Buffer)
	buff.Reset()
	deployScript.Execute(buff, info)

	fmt.Println(buff.String())

	cmd := exec.Command("bash", "-c", buff.String())
	stdouterr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("script error: %s", err.Error())
		fmt.Println(stdouterr)
		//sendCallback(w, payload.CallbackURL, false, "script error: "+err.Error())
		//return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})

}

func sendCallback(w http.ResponseWriter, url string, success bool, description string) {
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
		log.Printf("invalid callback: %d", res.StatusCode)
		//writeError(w, "invalid callback", http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, "{\"ok\":true}")
}
