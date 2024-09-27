package views

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"interrupted-desktop/src/data"
	"interrupted-desktop/src/types"
	"interrupted-desktop/src/uploads"
	"interrupted-desktop/src/utils"
	"io"
	"log"
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

func ShowDefaultView(apiKey string) {
	user := data.GetUserData(apiKey)
	stats := data.GetUserStats(apiKey)

	replacements := map[string]string{
		"username": user.Username,
		"avatar":   user.Avatar,
		"uploads":  fmt.Sprint(stats.Uploads),
		"pastes":   fmt.Sprint(stats.Pastes),
		"storage":  stats.Storage,
		"uid":      fmt.Sprint(stats.UID),
		"joined":   stats.Joined,
		"invitees": fmt.Sprint(stats.Invitees),
	}

	cssAssets := []string{"assets/index.css"}
	jsAssets := []string{"assets/dist/bundle.js"}

	pageContent, err := loadHTMLTemplate("assets/index.html", cssAssets, jsAssets, replacements)
	if err != nil {
		log.Fatalf("Failed to load template: %v", err)
	}

	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Interrupted.me")
	w.SetSize(1080, 800, webview.HintNone)

	encodedContent := base64.StdEncoding.EncodeToString([]byte(pageContent))
	w.Navigate("data:text/html;base64," + encodedContent)

	w.Bind("logOut", func() {
		data.DeleteApiKey()
		w.Terminate()
	})

	w.Bind("fetchGallery", func() interface{} {
		galleryData, err := utils.FetchGallery(apiKey)
		if err != nil {
			fmt.Printf("Error fetching gallery: %s\n", err)
			return "Error fetching gallery"
		}

		return galleryData
	})

	w.Bind("selectFile", func() string {
		filePath, err := dialog.File().Title("Select a file").Load()
		if err != nil {
			fmt.Println("Error selecting file:", err)
			return "No file selected"
		}

		fileExtension := filepath.Ext(filePath)
		fileName := fmt.Sprintf("%s.%s", utils.RandomString(10), fileExtension)

		msg, err := uploads.UploadAndCopyToClipboard(filePath, fileName, apiKey)
		if err != nil {
			fmt.Println("Error uploading file:", err)
			return "Error uploading file"
		}

		return msg
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

		fileName := fmt.Sprintf("%s.png", utils.RandomString(10))

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
		msg, err := uploads.UploadAndCopyToClipboard(filePath, fileName, apiKey)
		if err != nil {
			fmt.Println("Error uploading file:", err)
			return "Error uploading file"
		}

		return msg
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

func ShowLoginView() string {
	cssAssets := []string{"assets/auth/login.css"}
	jsAssets := []string{"assets/dist/bundle.js"}

	pageContent, err := loadHTMLTemplate("assets/auth/login.html", cssAssets, jsAssets, nil)
	if err != nil {
		log.Fatalf("Failed to load template: %v", err)
	}

	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Interrupted.me")
	w.SetSize(1080, 800, webview.HintNone)

	var apiKey string

	encodedContent := base64.StdEncoding.EncodeToString([]byte(pageContent))
	w.Navigate("data:text/html;base64," + encodedContent)

	w.Bind("logOut", func() {
		w.Terminate()
	})

	w.Bind("login", func(username string, password string) string {
		formData := url.Values{
			"username": {username},
			"password": {password},
		}

		headers := map[string]string{
			"User-Agent":   "Mozilla/5.0 (compatible; interrupted/1.0; +https://interrupted.me)",
			"Content-Type": "application/x-www-form-urlencoded",
		}

		resp, err := utils.SendAPIRequest("POST", "https://api.intrd.me/api/login", strings.NewReader(formData.Encode()), headers)
		if err != nil {
			fmt.Printf("error performing request: %s\n", err)
			return "Error performing request"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s\n", err)
			return "Error reading response body"
		}

		var loginResponse types.LoginResponse
		err = json.Unmarshal(body, &loginResponse)
		if err != nil {
			fmt.Printf("error unmarshalling response body: %s\n", err)
			return "Error unmarshalling response body"
		}

		if resp.StatusCode != http.StatusOK {
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

func loadHTMLTemplate(templatePath string, cssPaths []string, jsPaths []string, replacements map[string]string) (string, error) {
	htmlContent, err := assets.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read HTML file: %s", err)
	}

	var cssContent string
	for _, cssPath := range cssPaths {
		cssData, err := assets.ReadFile(cssPath)
		if err != nil {
			return "", fmt.Errorf("failed to read CSS file %s: %s", cssPath, err)
		}
		cssContent += string(cssData) + "\n"
	}

	var jsContent string
	for _, jsPath := range jsPaths {
		jsData, err := assets.ReadFile(jsPath)
		if err != nil {
			return "", fmt.Errorf("failed to read JS file %s: %s", jsPath, err)
		}
		jsContent += string(jsData) + "\n"
	}

	htmlWithReplacements := string(htmlContent)
	for placeholder, value := range replacements {
		htmlWithReplacements = strings.ReplaceAll(htmlWithReplacements, "{{"+placeholder+"}}", value)
	}

	htmlWithCSSAndJS := fmt.Sprintf(
		"%s<style>%s /* cache-buster: %d */</style><script>%s</script>",
		htmlWithReplacements,
		cssContent,
		time.Now().UnixNano(),
		jsContent,
	)

	return htmlWithCSSAndJS, nil
}
