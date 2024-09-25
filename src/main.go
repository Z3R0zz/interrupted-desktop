package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"image/png"
	"interrupted-desktop/src/data"
	"interrupted-desktop/src/types"
	"interrupted-desktop/src/uploads"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/sqweek/dialog"
	webview "github.com/webview/webview_go"
	"golang.design/x/clipboard"
)

//go:embed assets/*
var assets embed.FS

func main() {
	err := clipboard.Init()
	if err != nil {
		fmt.Println("Error initializing clipboard:", err)
		return
	}

	apiKey, err := data.ReadApiKey()
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	if apiKey == "" {
		apiKey = showLoginView()
		if apiKey == "" {
			return
		}
	}

	err = data.ClearAppData()
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	showDefaultView(apiKey)
}

func showDefaultView(apiKey string) {
	user := data.GetUserData(apiKey)

	htmlContent, err := assets.ReadFile("assets/index.html")
	if err != nil {
		log.Fatalf("Failed to read HTML file: %v", err)
	}

	cssContent, err := assets.ReadFile("assets/index.css")
	if err != nil {
		log.Fatalf("Failed to read CSS file: %v", err)
	}

	htmlWithUserData := strings.ReplaceAll(string(htmlContent), "{{username}}", user.Username)
	htmlWithUserData = strings.ReplaceAll(htmlWithUserData, "{{avatar}}", user.Avatar)

	htmlWithCSS := htmlWithUserData +
		"<style>" + string(cssContent) + " /* cache-buster: " + fmt.Sprint(time.Now().UnixNano()) + " */" + "</style>"

	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Interrupted.me")
	w.SetSize(1920, 1080, webview.HintNone)

	w.Navigate("data:text/html," + htmlWithCSS)

	w.Bind("logOut", func() {
		data.DeleteApiKey()
		w.Terminate()
	})

	w.Bind("fetchGallery", func() interface{} {

		requestURL := fmt.Sprintf("https://api.intrd.me/api/gallery/%v", apiKey)
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
			os.Exit(1)
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; interrupted/1.0; +https://interrupted.me)")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error reading response body: %s\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s\n", err)
			os.Exit(1)
		}

		if resp.StatusCode != 200 {
			fmt.Printf("API error: %s\n", string(body))
			return "API error"
		}

		var galleryResp types.GalleryResponse
		err = json.Unmarshal(body, &galleryResp)
		if err != nil {
			fmt.Printf("error unmarshalling response body: %s\n", err)
			return "Error parsing gallery response"
		}

		return galleryResp.Data
	})

	w.Bind("selectFile", func() string {
		filePath, err := dialog.File().Title("Select a file").Load()
		if err != nil {
			fmt.Println("Error selecting file:", err)
			return "No file selected"
		}

		fileExtension := filepath.Ext(filePath)
		fileName := fmt.Sprintf("%s.%s", randomString(10), fileExtension)

		url := uploads.UploadFile(filePath, fileName, apiKey)

		if url == "" {
			fmt.Println("Error uploading file")
			return "Error uploading file"
		}

		clipboard.Write(clipboard.FmtText, []byte(url))

		return fmt.Sprintf("Screenshot '%s' uploaded and URL copied to clipboard!", fileName)
	})

	w.Bind("copyToClipboard", func(url string) string {
		if url == "" {
			return "No URL provided"
		}

		clipboard.Write(clipboard.FmtText, []byte(url))

		return "URL copied to clipboard!"
	})

	w.Bind("captureScreenshot", func(index int) string {
		numDisplays := screenshot.NumActiveDisplays()
		if index >= numDisplays || index < 0 {
			return "Invalid display index"
		}

		bounds := screenshot.GetDisplayBounds(index)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return "Error capturing screenshot"
		}

		fileName := fmt.Sprintf("%s.png", randomString(10))

		appDataPath, err := data.GetAppDataPath()
		if err != nil {
			fmt.Println("Error getting app data path:", err)
			return "Error getting app data path"
		}

		savePath := filepath.Join(appDataPath, data.Subdirectory)
		os.MkdirAll(savePath, os.ModePerm)

		file, err := os.Create(filepath.Join(savePath, fileName))
		if err != nil {
			fmt.Println("Error saving screenshot:", err)
			return "Error saving screenshot"
		}
		defer file.Close()

		png.Encode(file, img)

		filePath := filepath.Join(savePath, fileName)
		url := uploads.UploadFile(filePath, fileName, apiKey)

		if url == "" {
			fmt.Println("Error uploading file")
			return "Error uploading file"
		}

		clipboard.Write(clipboard.FmtText, []byte(url))

		return fmt.Sprintf("Screenshot '%s' uploaded and URL copied to clipboard!", fileName)
	})

	w.Bind("fetchMonitors", func() interface{} {
		numDisplays := screenshot.NumActiveDisplays()
		var monitors []map[string]int
		for i := 0; i < numDisplays; i++ {
			bounds := screenshot.GetDisplayBounds(i)
			monitors = append(monitors, map[string]int{
				"width":  bounds.Dx(),
				"height": bounds.Dy(),
			})
		}
		return monitors
	})

	w.Run()
}

func showLoginView() string {
	htmlContent, err := assets.ReadFile("assets/auth/login.html")
	if err != nil {
		log.Fatalf("Failed to read HTML file: %v", err)
	}

	cssContent, err := assets.ReadFile("assets/auth/login.css")
	if err != nil {
		log.Fatalf("Failed to read CSS file: %v", err)
	}

	htmlWithCSS := string(htmlContent) +
		"<style>" + string(cssContent) + " /* cache-buster: " + fmt.Sprint(time.Now().UnixNano()) + " */" + "</style>"

	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Interrupted.me")
	w.SetSize(1080, 800, webview.HintNone)

	var apiKey string

	w.Navigate("data:text/html," + htmlWithCSS)

	w.Bind("logOut", func() {
		w.Terminate()
	})

	w.Bind("login", func(username string, password string) string {
		formData := url.Values{
			"username": {username},
			"password": {password},
		}

		req, err := http.NewRequest("POST", "https://api.intrd.me/api/login", strings.NewReader(formData.Encode()))
		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
			return "Error making http request"
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; interrupted/1.0; +https://interrupted.me)")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error reading response body: %s\n", err)
			return "Error reading response body"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s\n", err)
			os.Exit(1)
		}

		var loginResponse types.LoginResponse
		err = json.Unmarshal(body, &loginResponse)
		if err != nil {
			fmt.Printf("error unmarshalling response body: %s\n", err)
			return "Error unmarshalling response body"
		}

		if resp.StatusCode != 200 {
			if loginResponse.Message != nil {
				fmt.Printf("API error: %s\n", *loginResponse.Message)
				return *loginResponse.Message
			} else {
				fmt.Printf("API error: unknown error\n")
				return "Unknown error"
			}
		}

		if err := data.SaveApiKey(loginResponse.Data.ApiKey); err != nil {
			log.Fatalf("Failed to save API key: %v", err)
			return "Failed to save API key"
		}

		apiKey = loginResponse.Data.ApiKey
		w.Terminate()
		return apiKey
	})

	w.Run()

	return apiKey
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
