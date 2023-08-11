# go-reddit-uploader: Upload media files to Reddit using the platform's native API. Supports images, videos, or GIFs

This package provides only the basic functionality to upload media files to Reddit because at the moment there are no other packages that support native media uploads. If you look for more advanced features, you can use an excellent wrapper [go-reddit](https://github.com/vartanbeno/go-reddit) by vartanbeno.


## Installation

```bash
go get -u github.com/mariownyou/go-reddit-uploader/reddit_uploader
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
    client, _ := reddit_uploader.New("username", "password", "client_id", "client_secret", "user_agent")

    // Set up the post
    post := reddit_uploader.Submission{
        Subreddit: "subreddit",
        Title: "title",
    }

    err := client.SubmitImage(post, "image.jpg")
    if err != nil {
        panic(err)
    }
}
```


## Uplaoding Videos

```golang

err := client.SubmitVideo(post, "video.mp4", "preview.jpg")
if err != nil {
    panic(err)
}

```


## License
[License](LICENSE)
