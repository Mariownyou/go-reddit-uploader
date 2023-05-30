package reddit_uploader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Test struct {
	name          string
	server        *httptest.Server
	response      string
	expectedError error
}

func TestUploadMedia(t *testing.T) {
	tests := []Test{
		{
			name: "TestUploadMedia",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"args":{"action": "post", "fields": [{"name": "key", "value": "value"}]}}`))
			})),
			response:      "Hello, client",
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			resp, err := UploadMedia(test.server.URL, []byte("Hello, server"), "test.txt")
			fmt.Println("response", resp, err)
		})
	}
}
