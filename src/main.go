package main

import (
	"embed"
	"fmt"
	"image/png"
	"interrupted-desktop/src/data"
	"interrupted-desktop/src/uploads"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
		log.Print("API Key not found.")
		apiKey, err = data.PromptForApiKey()
		if err != nil {
			log.Fatalf("Failed to prompt for API key: %v", err)
		}

		if err := data.SaveApiKey(apiKey); err != nil {
			log.Fatalf("Failed to save API key: %v", err)
		}
	}

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
	w.SetSize(1080, 800, webview.HintNone)

	w.Navigate("data:text/html," + htmlWithCSS)

	w.Bind("logOut", func() {
		data.DeleteApiKey()
		w.Terminate()
	})

	w.Bind("selectFile", func() string {
		fmt.Println("Select file")
		filePath, err := dialog.File().Title("Select a file").Load()
		if err != nil {
			fmt.Println("Error selecting file:", err)
			return ""
		}

		url := uploads.UploadFile(filePath, apiKey)

		if url == "" {
			fmt.Println("Error uploading file")
			return "Error uploading file"
		}

		clipboard.Write(clipboard.FmtText, []byte(url))

		return "Screenshot uploaded and URL copied to clipboard!"
	})

	w.Bind("captureScreenshot", func(param string) string {
		i, err := strconv.Atoi(param)
		if err != nil {
			fmt.Println("Error converting parameter to int:", err)
			return "Error converting parameter to int"
		}

		numDisplays := screenshot.NumActiveDisplays()
		if i >= numDisplays || i < 0 {
			fmt.Println("Invalid display index")
			return "Invalid display index"
		}

		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			fmt.Println("Error capturing screenshot:", err)
			return "Error capturing screenshot"
		}

		fileName := fmt.Sprintf("screenshot-%d.png", time.Now().Unix())
		appDataPath, err := data.GetAppDataPath()
		if err != nil {
			fmt.Println("Error getting app data path:", err)
			return "Error getting app data path"
		}

		savepath := filepath.Join(appDataPath, data.Subdirectory)

		file, err := os.Create(filepath.Join(savepath, fileName))
		if err != nil {
			fmt.Println("Error saving screenshot:", err)
			return "Error saving screenshot"
		}
		defer file.Close()

		png.Encode(file, img)

		filepath := filepath.Join(savepath, fileName)
		url := uploads.UploadFile(filepath, apiKey)

		if url == "" {
			fmt.Println("Error uploading file")
			return "Error uploading file"
		}

		clipboard.Write(clipboard.FmtText, []byte(url))

		return "Screenshot uploaded and URL copied to clipboard!"
	})

	w.Run()
}
