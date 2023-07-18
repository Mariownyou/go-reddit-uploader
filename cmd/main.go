package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mariownyou/go-reddit-uploader/reddit_uploader"
)

func main() {
	username := flag.String("username", "", "Reddit username")
	password := flag.String("password", "", "Reddit password")
	clientID := flag.String("client-id", "", "Reddit client ID")
	clientSecret := flag.String("client-secret", "", "Reddit client secret")
	flag.Parse()

	client, _ := reddit_uploader.New(*username, *password, *clientID, *clientSecret)

	file, _ := os.ReadFile("cmd/image.png")
	// post := reddit_uploader.Submission{
	// 	Subreddit: "test",
	// 	Title:     "Test post from API",
	// }
	// postLink, _ := client.SubmitImage(post, file, "image.png")
	// fmt.Println("Post Link:", postLink)

	video, _ := os.ReadFile("cmd/vid.mp4")
	params := reddit_uploader.Submission{
		Subreddit: "VerifiedFeet",
		Title:     "Test post",
		FlairID:   "19f59cb2-7c9f-11e8-bd3a-0edbbe2223ea",
	}
	// videoPost, _ := client.SubmitVideo(params, video, nil, "video.mp4")
	// fmt.Println("Post Link:", videoPost)
	preview, _ := client.UploadMedia(file, "image.jpg")
	link, _ := client.UploadMedia(video, "video.mp4")

	u := func() {
		res, err := client.SubmitVideoLink(params, link, preview)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res)
	}

	u()
	time.Sleep(1 * time.Second)
	u()
}
