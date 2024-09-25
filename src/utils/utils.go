package utils

import (
	"encoding/json"
	"fmt"
	"interrupted-desktop/src/types"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func SendAPIRequest(method, urlStr string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return resp, nil
}

func FetchGallery(apiKey string) ([]types.ImageData, error) {
	url := fmt.Sprintf("https://api.intrd.me/api/gallery/%v", apiKey)

	resp, err := SendAPIRequest("GET", url, nil, map[string]string{
		"User-Agent": "Mozilla/5.0 (compatible; interrupted/1.0; +https://interrupted.me)",
	})
	if err != nil {
		return nil, fmt.Errorf("error performing request: %v", err)
	}
	defer resp.Body.Close()

	var galleryResp types.GalleryResponse
	if err := json.NewDecoder(resp.Body).Decode(&galleryResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return galleryResp.Data, nil
}
