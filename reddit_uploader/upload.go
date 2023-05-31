package reddit_uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type RedditUplaoderClient struct {
	authHost     string
	apiHost      string
	username     string
	password     string
	clientID     string
	clientSecret string
	accessToken  string
}

func newRedditUplaoderClient(authHost, apiHost, username, password, clientID, clientSecret string) *RedditUplaoderClient {
	return &RedditUplaoderClient{
		authHost,
		apiHost,
		username,
		password,
		clientID,
		clientSecret,
		"",
	}
}

func NewRedditUplaoderClient(username, password, clientID, clientSecret string) *RedditUplaoderClient {
	c := newRedditUplaoderClient(
		"https://www.reddit.com",
		"https://oauth.reddit.com",
		username,
		password,
		clientID,
		clientSecret,
	)

	accessToken, err := c.GetAccessToken()
	if err != nil {
		panic(err) // TODO: handle error
	}

	c.accessToken = accessToken

	return c
}

func (c *RedditUplaoderClient) GetAccessToken() (string, error) {
	// Set up the form data
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", c.username)
	form.Add("password", c.password)

	// Set up the HTTP request
	req, err := http.NewRequest("POST", c.authHost+"/api/v1/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// add basic auth
	req.SetBasicAuth(c.clientID, c.clientSecret)

	// Set the user agent header
	req.Header.Set("User-Agent", "go-reddit-uploader (by /u/mariownyou)")

	// Set up the HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}

	defer resp.Body.Close()

	// parse the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(responseBody, &dataMap); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", err
	}

	accessToken, ok := dataMap["access_token"].(string)
	if !ok {
		fmt.Println("Error getting access token")
		return "", err
	}

	return accessToken, nil
}

func (c *RedditUplaoderClient) UploadMedia(file []byte, filename string) (string, error) {
	filetypeSplit := strings.Split(filename, ".")
	filetype := filetypeSplit[len(filetypeSplit)-1]

	// filetypes map
	filetypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"mp4":  "video/mp4",
		"mov":  "video/quicktime",
	}

	// Set up the form data
	form := url.Values{}
	form.Add("filepath", filename)
	form.Add("mimetype", filetypes[filetype])
	// form.Add("mimetype", "image/gif")
	form.Add("api_type", "json")

	// Set up the HTTP request
	req, err := http.NewRequest("POST", c.apiHost+"/api/media/asset.json", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// add the access token header
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	// Set the user agent header
	req.Header.Set("User-Agent", "go-reddit-uploader (by /u/mariownyou)")

	// Set up the HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}

	defer resp.Body.Close()

	// parse the response body
	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	var dataMap map[string]interface{}
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	if err := json.Unmarshal(responseBody, &dataMap); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", err
	}

	action := dataMap["args"].(map[string]interface{})["action"].(string)
	actionURL := fmt.Sprintf("https:%s", action)
	uploadFields := dataMap["args"].(map[string]interface{})["fields"].([]interface{})

	uploadData := make(map[string]string)
	for _, field := range uploadFields {
		fieldMap := field.(map[string]interface{})
		uploadData[fieldMap["name"].(string)] = fieldMap["value"].(string)
	}

	// pretty print the JSON response
	prettyJSON, err := json.MarshalIndent(uploadData, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return "", err
	}

	fmt.Println(string(prettyJSON))

	// Create a new multipart buffer
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, value := range uploadData {
		writer.WriteField(key, value)
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// Write the bytes to the part
	_, err = io.Copy(part, bytes.NewReader(file))
	if err != nil {
		fmt.Println("Error copying file to part:", err)
		return "", err
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// Create a new HTTP request
	req, err = http.NewRequest("POST", actionURL, body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client = &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	// parse SOAP response
	responseBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	regexp := regexp.MustCompile("<Location>(.*)</Location>")
	location := regexp.FindStringSubmatch(string(responseBody))[1]
	location = fmt.Sprintf("%s/%s", actionURL, uploadData["key"])
	// wait for 1 ses
	time.Sleep(1 * time.Second)
	return location, nil
}

func (c *RedditUplaoderClient) SubmitMediaAsLink(file []byte, filename string) (string, error) {
	link, err := c.UploadMedia(file, filename)
	if err != nil {
		fmt.Println("Error uploading post to reddit server:", err)
		return "", err
	}

	fmt.Println(link)

	kinds := map[string]string{
		"jpg":  "link",
		"jpeg": "link",
		"png":  "link",
		"gif":  "video",
		"mp4":  "video",
		"mov":  "video",
	}

	filenameSplit := strings.Split(filename, ".")
	kind := kinds[filenameSplit[len(filenameSplit)-1]]

	postLink, err := c.SubmitLink(link, kind)
	if err != nil {
		fmt.Println("Error submitting post:", err)
		return "", err
	}

	return postLink, nil
}

func (c *RedditUplaoderClient) SubmitLink(link, kind string) (string, error) {
	// Set up the form data
	form := url.Values{}
	form.Add("api_type", "json")
	form.Add("kind", kind)
	form.Add("sr", "test")
	if kind == "video" {
		file, _ := os.ReadFile("cmd/image.jpg")
		poster, _ := c.UploadMedia(file, "image.jpg")
		fmt.Println("poster:", poster)
		form.Add("video_poster_url", poster)
	}
	form.Add("title", "Test post from API")
	form.Add("url", link)

	// Set up the HTTP request
	req, err := http.NewRequest("POST", "https://oauth.reddit.com/api/submit", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// add the access token header
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	// Set the user agent header
	req.Header.Set("User-Agent", "go-reddit-uploader (by /u/mariownyou)")

	// Set up the HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}

	defer resp.Body.Close()

	// parse the response body
	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	return string(responseBody), nil
}
