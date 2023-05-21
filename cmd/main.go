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

	fmt.Println(*username, *password, *clientID, *clientSecret)

	file, err := os.ReadFile("cmd/image.jpg")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	postLink, _ := reddit_uploader.SubmitMedia(*username, *password, *clientID, *clientSecret, file, "image.jpg")
	fmt.Println("Post Link:", postLink)
}
