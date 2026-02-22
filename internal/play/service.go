package play

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	cachedToken string
	tokenExpiry time.Time
	mu          sync.Mutex
)

func getAccessToken() (string, error) {
	mu.Lock()
	defer mu.Unlock()

	if cachedToken != "" && time.Now().Before(tokenExpiry) {
		return cachedToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", os.Getenv("SPOTIFY_REFRESH_TOKEN"))
	data.Set("client_id", os.Getenv("SPOTIFY_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("SPOTIFY_CLIENT_SECRET"))

	resp, err := http.Post(
		"https://accounts.spotify.com/api/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != 200 || result.AccessToken == "" {
		return "", fmt.Errorf("spotify token refresh failed (status %d): %s - %s", resp.StatusCode, result.Error, result.ErrorDesc)
	}

	cachedToken = result.AccessToken
	tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return cachedToken, nil
}

func spotifyGet(url string) (*http.Response, error) {
	token, err := getAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	return http.DefaultClient.Do(req)
}
