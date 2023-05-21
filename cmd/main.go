package main

import (
	"flag"
	"fmt"

	"github.com/mariownyou/go-reddit-submit-image/submit_image"
)

func main() {
	username := flag.String("username", "", "Reddit username")
	password := flag.String("password", "", "Reddit password")
	clientID := flag.String("client-id", "", "Reddit client ID")
	clientSecret := flag.String("client-secret", "", "Reddit client secret")
	flag.Parse()

	fmt.Println(*username, *password, *clientID, *clientSecret)

	postLink, _ := submit_image.SubmitMedia(*username, *password, *clientID, *clientSecret, "cmd/image.jpg")
	fmt.Println("Post Link:", postLink)
}
