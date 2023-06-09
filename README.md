# go-reddit-uploader: Upload media files to Reddit using the platform's native API. Supports images, videos, or GIFs

This package provides only the basic functionality to upload media files to Reddit because at the moment there are no other packages that support native media uploads. If you look for more advanced features, you can use an excellent wrapper [go-reddit](https://github.com/vartanbeno/go-reddit) by vartanbeno.


## Installation

```bash
go get -u github.com/mariownyou/go-reddit-uploader
```

## Usage

```golang
package main

import (
    "fmt"
    "github.com/mariownyou/go-reddit-uploader"
)

func main() {
    // Create a new uploader
    client, _ := reddit_uploader.New("username", "password", "client_id", "client_secret")

    // Read the file
    file, _ := os.ReadFile("path/to/file.jpg")

    // Set up the post
    post := reddit_uploader.Submission{
        Subreddit: "subreddit",
        Title: "title",
    }

    response, err := client.SubmitImage(post, file, "image.jpg")
    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```


## Uplaoding Videos

```golang

video, _ := os.ReadFile("path/to/video.mp4")
preview, _ := os.ReadFile("path/to/preview.jpg")

post := reddit_uploader.Submission{
    Subreddit: "subreddit",
    Title: "title",
}

response, _ := client.SubmitVideo(post, video, nil, "video.mp4") // if preview is nil, default preview will be used
response, _ := client.SubmitVideo(post, video, preview, "video.mp4")
response, _ := client.SubmitVideoLink(post, video, preview, "video.mp4") // Some communities dooesn't allow video uploads, so you can use this method to upload a video link instead, reddit will rednder this link as a regular video

```


## License
[License](LICENSE)
