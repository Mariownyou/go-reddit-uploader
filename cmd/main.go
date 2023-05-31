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

	redditUploader := reddit_uploader.NewRedditUplaoderClient(*username, *password, *clientID, *clientSecret)

	// file, _ := os.ReadFile("cmd/image.jpg")
	// postLink, _ := redditUploader.SubmitMediaAsLink(file, "image.jpg")
	// fmt.Println("Post Link:", postLink)

	video, _ := os.ReadFile("cmd/vid.mp4")
	videoPost, _ := redditUploader.SubmitVideo(video, nil, "video.mp4")
	fmt.Println("Post Link:", videoPost)
}
