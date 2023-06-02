package reddit_uploader

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "123"}`))
	}))

	defer s.Close()

	client := newRedditUplaoder(s.URL, s.URL, "username", "password", "clientID", "clientSecret")
	token, err := client.GetAccessToken()

	if err != nil {
		t.Error("error is not nil", err)
	}

	if token != "123" {
		t.Error("token is not correct", token)
	}

	// test bad request
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"error": "invalid_grant"}`))
	}))

	defer s.Close()

	client = newRedditUplaoder(s.URL, s.URL, "username", "password", "clientID", "clientSecret")
	token, err = client.GetAccessToken()

	if err.Error() != "invalid_grant" {
		t.Error("error is not correct", err)
	}

	if token != "" {
		t.Error("token is not empty", token)
	}
}
