package reddit_uploader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-querystring/query"
)

var (
	ErrImagesNotAllowed      = errors.New("community does not allow images")
	ErrSubredditDoesNotExist = errors.New("community does not exist")
	ErrSubredditNotAllowed   = errors.New("this community only allows trusted members to post here")
	ErrMissingVideoURLs      = errors.New("this community requires a video link and a post link")
)

type Submission struct {
	Subreddit string `url:"sr,omitempty"`
	Title     string `url:"title,omitempty"`

	FlairID   string `url:"flair_id,omitempty"`
	FlairText string `url:"flair_text,omitempty"`

	SendReplies *bool `url:"sendreplies,omitempty"`
	Resubmit    bool  `url:"resubmit,omitempty"`
	NSFW        bool  `url:"nsfw,omitempty"`
	Spoiler     bool  `url:"spoiler,omitempty"`
}

type Response struct {
	JSON struct {
		Errors [][]string `json:"errors"`
		Data   struct {
			UserSubmittedPage string `json:"user_submitted_page"`
			WebsocketURL      string `json:"websocket_url"`
		} `json:"data"`
	} `json:"json"`
}

type RedditUplaoder struct {
	authHost     string
	apiHost      string
	username     string
	password     string
	clientID     string
	clientSecret string
	accessToken  string
}

func newRedditUplaoder(authHost, apiHost, username, password, clientID, clientSecret string) *RedditUplaoder {
	return &RedditUplaoder{
		authHost,
		apiHost,
		username,
		password,
		clientID,
		clientSecret,
		"",
	}
}

func New(username, password, clientID, clientSecret string) (*RedditUplaoder, error) {
	c := newRedditUplaoder(
		"https://www.reddit.com",
		"https://oauth.reddit.com",
		username,
		password,
		clientID,
		clientSecret,
	)

	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	c.accessToken = accessToken

	return c, nil
}

func (c *RedditUplaoder) GetAccessToken() (string, error) {
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", c.username)
	form.Add("password", c.password)

	req, err := http.NewRequest("POST", c.authHost+"/api/v1/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("User-Agent", "go-reddit-uploader (by /u/mariownyou)")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(responseBody, &dataMap); err != nil {
		return "", err
	}

	if e, ok := dataMap["error"].(string); ok {
		return "", fmt.Errorf(e)
	}

	accessToken, ok := dataMap["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("error getting access token")
	}

	return accessToken, nil
}

func (c *RedditUplaoder) UploadMedia(file []byte, filename string) (string, error) {
	filetypeSplit := strings.Split(filename, ".")
	filetype := filetypeSplit[len(filetypeSplit)-1]

	filetypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"mp4":  "video/mp4",
		"mov":  "video/quicktime",
	}

	form := url.Values{}
	form.Add("filepath", filename)
	form.Add("mimetype", filetypes[filetype])
	form.Add("api_type", "json")

	req, err := http.NewRequest("POST", c.apiHost+"/api/media/asset.json", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", "go-reddit-uploader (by /u/mariownyou)")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var dataMap map[string]interface{}
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(responseBody, &dataMap); err != nil {
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
		return "", err
	}

	fmt.Println(string(prettyJSON))

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, value := range uploadData {
		writer.WriteField(key, value)
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, bytes.NewReader(file))
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err = http.NewRequest("POST", actionURL, body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 {
		responseBody, err = io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		fmt.Println("Error uploading file:", res.Status, string(responseBody))
	}

	link := fmt.Sprintf("%s/%s", actionURL, uploadData["key"])
	return link, nil
}

func (c *RedditUplaoder) SubmitVideo(params Submission, video []byte, preview []byte, filename string) (string, error) {
	videoLink, err := c.UploadMedia(video, filename)
	if err != nil {
		return "", err
	}

	if preview == nil {
		preview, _ = os.ReadFile("cmd/image.png")
	}

	previewLink, err := c.UploadMedia(preview, "preview.jpg")
	if err != nil {
		return "", err
	}

	form := struct {
		Submission
		Kind       string `url:"kind,omitempty"`
		URL        string `url:"url,omitempty"`
		PreviewURL string `url:"video_poster_url,omitempty"`
	}{params, "video", videoLink, previewLink}

	return c.submit(form)
}

func (c *RedditUplaoder) SubmitVideoLink(params Submission, video, preview, filename string) (string, error) {
	form := struct {
		Submission
		Kind       string `url:"kind,omitempty"`
		URL        string `url:"url,omitempty"`
		PreviewURL string `url:"video_poster_url,omitempty"`
	}{params, "link", video, preview}

	return c.submit(form)
}

func (c *RedditUplaoder) SubmitImage(params Submission, image []byte, filename string) (string, error) {
	link, err := c.UploadMedia(image, filename)
	if err != nil {
		return "", err
	}

	form := struct {
		Submission
		Kind string `url:"kind,omitempty"`
		URL  string `url:"url,omitempty"`
	}{params, "link", link}

	return c.submit(form)
}

func (c *RedditUplaoder) submit(v interface{}) (string, error) {
	form, err := query.Values(v)
	if err != nil {
		fmt.Println("Error parsing query params:", err)
		return "", err
	}
	form.Set("api_type", "json")
	fmt.Println(form.Encode())

	req, err := http.NewRequest("POST", c.apiHost+"/api/submit", strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", "go-reddit-uploader (by /u/mariownyou)")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	return parseResponse(responseBody)
}

func parseResponse(body []byte) (string, error) {
	var response Response

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.JSON.Errors) > 0 {
		switch response.JSON.Errors[0][0] {
		case "IMAGES_NOTALLOWED":
			return "", ErrImagesNotAllowed
		case "SUBREDDIT_NOEXIST":
			return "", ErrSubredditDoesNotExist
		case "SUBREDDIT_NOTALLOWED":
			return "", ErrSubredditNotAllowed
		case "MISSING_VIDEO_URLS":
			return "", ErrMissingVideoURLs
		default:
			return "", errors.New(response.JSON.Errors[0][1])
		}
	}

	jsonString, err := json.Marshal(response.JSON.Data)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
