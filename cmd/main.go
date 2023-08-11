package main

import (
	"flag"
	"fmt"

	"github.com/mariownyou/go-reddit-uploader/reddit_uploader"
)

func main() {
	username := flag.String("username", "", "Reddit username")
	password := flag.String("password", "", "Reddit password")
	clientID := flag.String("client-id", "", "Reddit client ID")
	clientSecret := flag.String("client-secret", "", "Reddit client secret")
	flag.Parse()

	params := reddit_uploader.Submission{
		Title: "Test",
		Subreddit: "test",
	}

	uploader, err := reddit_uploader.New(*username, *password, *clientID, *clientSecret, "go-reddit-uploader (by /u/mariownyou)")
	if err != nil {
		panic(err)
	}

    err = uploader.SubmitVideo(params, "cmd/video.mp4", "cmd/image.png")
    err = uploader.SubmitImage(params, "cmd/image.png")

    if err != nil {
    	panic(err)
    }
}
