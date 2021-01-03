// Package plugindemo a demo plugin.
package plugindemo

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"time"
)

const (
	metadataServer = "http://169.254.169.254"
)

type iamCredentials struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
}

// Config the plugin configuration.
type Config struct {
	BucketName string
	Endpoint   string
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		BucketName: "",
		Endpoint:   "",
	}
}

// Demo a Demo plugin.
type Demo struct {
	next       http.Handler
	bucketName string
	endpoint   string
	name       string
	template   *template.Template
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.BucketName) == 0 {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}

	if len(config.Endpoint) == 0 {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	return &Demo{
		bucketName: config.BucketName,
		endpoint:   config.Endpoint,
		next:       next,
		name:       name,
		template:   template.New("demo").Delims("[[", "]]"),
	}, nil
}

func getEc2Role() string {
	resp, err := http.Get(metadataServer + "/latest/meta-data/iam/security-credentials/")
	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
	}

	awsRole := string(body)
	return awsRole
}

func getIamCredentials() iamCredentials {
	iamRole := getEc2Role()
	resp, err := http.Get(metadataServer + "/latest/meta-data/iam/security-credentials/" + iamRole)
	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
	}

	var creds iamCredentials
	json.Unmarshal([]byte(body), &creds)

	return creds
}

func (a *Demo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		rw.Write([]byte("No directory listing allowed"))
		return
	}

	newReq := mySign(req, a.bucketName, a.endpoint)

	client := &http.Client{}
	resp, err := client.Do(newReq)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(resp.Status + "\n")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
	}
	bodyString := string(body)
	fmt.Print(bodyString + "\n")

	if resp.Status == "200 OK" {
		rw.Write([]byte(bodyString))
	} else {
		rw.Write([]byte("error"))
	}

	return
}

func mySign(originalReq *http.Request, bucket string, endpoint string) *http.Request {
	urlPath := originalReq.URL.Path

	iam := getIamCredentials()
	file := strings.TrimLeft(urlPath, "/")

	t := time.Now()
	date := t.Format(time.RFC1123Z)

	resource := "/" + bucket + "/" + file
	signatureString := "GET\n\n\n" + date + "\nx-amz-security-token:" + iam.Token + "\n/" + resource
	//hago hmac de skey usando signature y a eso lo hago base64
	signatureStep1 := makeHMac([]byte(iam.SecretAccessKey), []byte(signatureString))
	signatureStep2 := b64.StdEncoding.EncodeToString([]byte(signatureStep1))
	authorization := "AWS " + iam.AccessKeyId + ":" + signatureStep2

	fullEndpoint := endpoint + "/" + resource
	req, err := http.NewRequest(http.MethodGet, fullEndpoint, nil)
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Add("Date", date)
	req.Header.Add("X-AMZ-Security-Token", iam.Token)
	req.Header.Add("Authorization", authorization)

	return req
}

func makeHMac(key []byte, data []byte) []byte {
	hash := hmac.New(sha1.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}
