package reddit_uploader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func SubmitMedia(username, password, clientID, clientSecret string, file []byte, filename string) (string, error) {
	accessToken, err := GetAccessToken(username, password, clientID, clientSecret)
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return "", err
	}

	link, err := UploadMedia(accessToken, file, filename)
	if err != nil {
		fmt.Println("Error submitting post:", err)
		return "", err
	}

	fmt.Println(link)

	postLink, err := submitLink(accessToken, link)
	if err != nil {
		fmt.Println("Error submitting post:", err)
		return "", err
	}

	return postLink, nil
}

func submitLink(accessToken, link string) (string, error) {
	// Set up the form data
	form := url.Values{}
	form.Add("api_type", "json")
	form.Add("kind", "video")
	form.Add("sr", "test")
	form.Add("video_poster_url", "https://reddit-uploaded-media.s3-accelerate.amazonaws.com/rte_images%2Fb8t0h53gfg1b1")
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
