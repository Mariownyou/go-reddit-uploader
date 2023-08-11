package reddit_uploader

import (
	"bytes"
	"os"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"

	"github.com/google/go-querystring/query"
)

type Uploader struct {
	username, password, clientID, clientSecret, userAgent, token string
}

func New(username, password, clientID, clientSecret, userAgent string) (*Uploader, error) {
	u := &Uploader{
		username:     username,
		password:     password,
		clientID:     clientID,
		clientSecret: clientSecret,
		userAgent:    userAgent,
	}

	token, err := u.GetAccessToken()
	if err != nil {
		return nil, err
	}

	u.token = token
	return u, nil
}

func (u *Uploader) GetAccessToken() (string, error) {
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", u.username)
	form.Add("password", u.password)

	resp := u.Post("https://www.reddit.com/api/v1/access_token", strings.NewReader(form.Encode()), true)
	defer resp.Body.Close()

	type TokenResponse struct {
		Token string `json:"access_token"`
		Error int    `json:"error"`
	}

	var content TokenResponse
	err := json.NewDecoder(resp.Body).Decode(&content)
	if err != nil {
		return "", fmt.Errorf("ERROR: Could not unmarshal response: %s\n", err)
	}

	if content.Error != 0 {
		return "", fmt.Errorf("ERROR: Could not get access token: %d\n", content.Error)
	}

	return content.Token, nil
}

func (u *Uploader) SubmitImage(params Submission, imagePath string) error {
	imageURL, _ := u.UploadMedia(imagePath)

	post := struct {
		Submission
		Kind string `url:"kind,omitempty"`
		URL  string `url:"url,omitempty"`
	}{params, "image", imageURL}

	return u.SubmitMedia(post)
}

func (u *Uploader) SubmitVideo(params Submission, videoPath, previewPath string) error {
	videoURL, _ := u.UploadMedia(videoPath)
	previewURL, _ := u.UploadMedia(previewPath)

	post := struct {
		Submission
		Kind       string `url:"kind,omitempty"`
		URL        string `url:"url,omitempty"`
		PreviewURL string `url:"video_poster_url,omitempty"`
	}{params, "video", videoURL, previewURL}

	return u.SubmitMedia(post)
}

func (u *Uploader) Post(url string, data io.Reader, auth ...bool) *http.Response {
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		panic(fmt.Errorf("ERROR: Could not create a request object: %s\n", err))
	}

	if auth != nil {
		req.SetBasicAuth(u.clientID, u.clientSecret)
	} else {
		req.Header.Set("Authorization", "Bearer "+u.token)
	}
	req.Header.Set("User-Agent", u.userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("ERROR: could not perform a request: %s\n", err))
	}

	return resp
}

func (u *Uploader) UploadMedia(mediaPath string) (string, string) {
	split := strings.Split(mediaPath, ".")
	if len(split) < 2 {
		panic(fmt.Errorf("ERROR: Filepath does not contain any extension\n"))
	}

	ext := "." + split[len(split)-1]
	mimetype := mime.TypeByExtension(ext)
	if mimetype == "" {
		panic(fmt.Errorf("ERROR: Uknown extension\n"))
	}

	form := url.Values{}
	form.Add("filepath", mediaPath)
	form.Add("mimetype", mimetype)

	resp := u.Post("https://oauth.reddit.com/api/media/asset.json", strings.NewReader(form.Encode()))

	defer resp.Body.Close()

	var content UploadMediaResponse
	err := json.NewDecoder(resp.Body).Decode(&content)
	if err != nil {
		panic(fmt.Errorf("ERROR: Could not unmarshal response: %s\n", err))
	}

	uploadLease := content.Args
	uploadURL := "https:" + uploadLease.Action
	uploadData := map[string]string{}

	for _, arg := range uploadLease.Fields {
		uploadData[arg.Name] = arg.Value
	}

	resp = ReadAndPostMedia(mediaPath, uploadURL, uploadData)
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		content, _ := io.ReadAll(resp.Body)
		panic(fmt.Errorf("ERROR: Could not post meida: %s", string(content)))
	}

	mediaURL := fmt.Sprintf("%s/%s", uploadURL, uploadData["key"])
	return mediaURL, content.Asset.WebsocketURL
}

func (u *Uploader) SubmitMedia(post interface{}) error {
	form, err := query.Values(post)
	if err != nil {
		panic(fmt.Errorf("Error parsing query params: %s", err))
	}

	form.Add("api_type", "json")

	resp := u.Post("https://oauth.reddit.com/api/submit", strings.NewReader(form.Encode()))
	defer resp.Body.Close()

	return ParseErrors(resp)
}

type UploadMediaResponse struct {
	Args struct {
		Action string `json:"action"`
		Fields []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"fields"`
	} `json:"args"`
	Asset struct {
		WebsocketURL string `json:"websocket_url"`
	} `json:"asset"`
}

func PostFiles(url string, data map[string]string, files ...FilePart) *http.Response {
	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)

	for key, value := range data {
		w.WriteField(key, value)
	}

	for _, fp := range files {
		part, err := w.CreateFormFile("file", fp.name)
		if err != nil {
			panic(fmt.Errorf("ERROR: Could not create form file: %s\n", err))
		}

		_, err = io.Copy(part, bytes.NewReader(fp.file))
		if err != nil {
			panic(fmt.Errorf("ERROR: Could not write file to form: %s\n", err))
		}
	}

	err := w.Close()
	if err != nil {
		panic(fmt.Errorf("ERROR: Could not close multipart form: %s\n", err))
	}

	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		panic(fmt.Errorf("ERROR: Could not create a request object: %s\n", err))
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("ERROR: could not perform a request: %s\n", err))
	}

	return resp
}

type FilePart struct {
	file []byte
	name string
}

func ReadAndPostMedia(mediaPath, uploadURL string, data map[string]string) *http.Response {
	file, err := os.ReadFile(mediaPath)
	if err != nil {
		panic(fmt.Errorf("ERROR: Could open the file %s\n", err))
	}

	f := FilePart{
		file: file,
		name: "file",
	}

	return PostFiles(uploadURL, data, f)
}

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

type SubmitMediaResponse struct {
	Message string `json:"message"`
	Error   int    `json:"error"`
	JSON    struct {
		Errors [][]string `json:"errors"`
		Data   struct {
			URL               string `json:"url"`
			UserSubmittedPage string `json:"user_submitted_page"`
			WebsocketURL      string `json:"websocket_url"`
		} `json:"data"`
	} `json:"json"`
}

func ParseErrors(r *http.Response) error {
	var content SubmitMediaResponse
	if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
		return fmt.Errorf("ERROR: Could not decode response: %s\n", err)
	}

	if len(content.JSON.Errors) > 0 {
		return fmt.Errorf("ERROR: Could not submit media: %s\n", content.JSON.Errors[0][1])
	}

	if content.Message != "" {
		return fmt.Errorf("ERROR: Could not submit media: %s\n", content.Message)
	}

	fmt.Println("Response Submit Media", content.JSON.Data)
	return nil
}
