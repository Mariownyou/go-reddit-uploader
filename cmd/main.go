package main

import (
	"fmt"

	"github.com/mariownyou/go-reddit-submit-image/submit_image"
)

func main() {
	postLink, _ := submit_image.SubmitMedia("", "", "", "", "image.jpg")
	fmt.Println(postLink)
}
