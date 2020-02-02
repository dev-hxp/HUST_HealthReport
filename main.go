package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"os/exec"
	"os"
	"log"
	"io"
	"bufio"
	"net/url"
	"strconv"
	"strings"
)

func main(){

	// input your username and password here
	username := "Username"
	password := "Password"

	// target url
	starturl := "https://yqtb.hust.edu.cn/infoplus/form/BKS/start"

	// Get
	client := &http.Client{}
	req, err := http.NewRequest("GET", starturl, nil)
	if err != nil {
		fmt.Println("http get error", err)
		return
	}

	// Set Header
	req.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36`)

	// Request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("request error", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}
	fmt.Println("Response:",string(body))

	// Crawl for lt
	re, _ := regexp.Compile(`value=.*-cas`)	
	lt := re.FindString(string(body))[7:]
	fmt.Println("Got lt:" + lt)

	// Crawl for execution
	re, _ = regexp.Compile(`execution.*"`)	
	execution := re.FindString(string(body))[18:22]
	fmt.Println("Got execution:" + execution)

	// Compute RSA
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("get directory error:", err)
		return
	}
	output, err := exec.Command("/usr/bin/node", path + "/des.js", username, password, lt).Output()
	if err != nil {
		fmt.Println("exec error:", err)
		return
	}
	rsa := string(output)
	fmt.Println("exec succeed, got rsa:", rsa)
	
	// Get the picture of code
	code_url := "http://pass.hust.edu.cn/cas/code"
	response, e := http.Get(code_url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	//open a file for writing
	file, err := os.Create("/tmp/code.gif")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Get code img Succeed!")

	// Show the image
	cmd := exec.Command("/usr/bin/eog", "/tmp/code.gif")
	err = cmd.Start()
	if err != nil {
		fmt.Println("exec error:", err)
		return
	}

	// Ask the user to input code
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter code:")
	code, _ := reader.ReadString('\n')
	fmt.Println("Got code:", code)
	
	// Create a pull request to login
	finalURL := resp.Request.URL.String()
	fmt.Println("finalURL:", finalURL)
	data := url.Values{}
	data.Set("rsa", rsa[:len(rsa)-1])
	data.Set("ul", "11")
	data.Set("pl", "10")
	data.Set("lt", lt)
	data.Set("execution", execution)
	data.Set("_eventId", "submit")
	u, _ := url.ParseRequestURI(finalURL)
	u.Path = ""
	urlStr := u.String() // "https://api.com/user/"

	client = &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	resp, _ = client.Do(r)
	r.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36`)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

