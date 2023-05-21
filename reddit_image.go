package submit_image

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
)

func SubmitMedia(username, password, clientID, clientSecret, filepath string) (string, error) {
	accessToken, err := getAccessToken(username, password, clientID, clientSecret)
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return "", err
	}

	link, err := submitMedia(accessToken, filepath)
	if err != nil {
		fmt.Println("Error submitting post:", err)
		return "", err
	}

	postLink, err := submitLink(accessToken, link)
	if err != nil {
		fmt.Println("Error submitting post:", err)
		return "", err
	}

	return postLink, nil
}

func submitMedia(accessToken, filepath string) (string, error) {
	// Set up the form data
	form := url.Values{}
	form.Add("filepath", filepath)
	form.Add("mimetype", "image/jpeg")
	form.Add("api_type", "json")

	// Set up the HTTP request
	req, err := http.NewRequest("POST", "https://oauth.reddit.com/api/media/asset.json", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// add the access token header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Set the user agent header
	req.Header.Set("User-Agent", "golang:com.example.golang:v0.1 (by /u/mariownyou)")

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

	// Add the file to the multipart form
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	_, err = io.Copy(part, file)

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

	return location, nil
}

func parseResponseBody(body io.ReadCloser) (map[string]interface{}, error) {
	var dataMap map[string]interface{}
	responseBody, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return dataMap, err
	}

	if err := json.Unmarshal(responseBody, &dataMap); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return dataMap, err
	}

	return dataMap, nil
}

func submitPost(accessToken string) (string, error) {
	// Set up the form data
	form := url.Values{}
	form.Add("api_type", "json")
	form.Add("kind", "self")
	form.Add("sr", "test")
	form.Add("title", "Test post from API")
	form.Add("text", "This is a test post from the API")

	// Set up the HTTP request
	req, err := http.NewRequest("POST", "https://oauth.reddit.com/api/submit", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// add the access token header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Set the user agent header
	req.Header.Set("User-Agent", "golang:com.example.golang:v0.1 (by /u/mariownyou)")

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

func submitLink(accessToken, link string) (string, error) {
	// Set up the form data
	form := url.Values{}
	form.Add("api_type", "json")
	form.Add("kind", "link")
	form.Add("sr", "test")
	form.Add("title", "Test post from API")
	form.Add("url", link)

	// Set up the HTTP request
	req, err := http.NewRequest("POST", "https://oauth.reddit.com/api/submit", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// add the access token header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Set the user agent header
	req.Header.Set("User-Agent", "golang:com.example.golang:v0.1 (by /u/mariownyou)")

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

func getAccessToken(username, password, clientID, clientSecret string) (string, error) {
	// Set up the form data
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", username)
	form.Add("password", password)

	// Set up the HTTP request
	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// add basic auth
	req.SetBasicAuth(clientID, clientSecret)

	// Set the user agent header
	req.Header.Set("User-Agent", "golang:com.example.golang:v0.1 (by /u/mariownyou)")

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
