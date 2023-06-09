package main

import (
	"flag"
	"fmt"
	"os"

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
	post := reddit_uploader.Submission{
		Subreddit: "test",
		Title:     "Test post from API",
	}
	postLink, _ := client.SubmitImage(post, file, "image.png")
	fmt.Println("Post Link:", postLink)

	// video, _ := os.ReadFile("cmd/vid.mp4")
	// params := reddit_uploader.Submission{
	// 	Subreddit: "test",
	// 	Title:     "Test post",
	// }
	// videoPost, _ := client.SubmitVideo(params, video, nil, "video.mp4")
	// fmt.Println("Post Link:", videoPost)
}
