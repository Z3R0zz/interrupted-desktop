package uploads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"interrupted-desktop/src/data"
	"interrupted-desktop/src/types"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kbinani/screenshot"
	"golang.design/x/clipboard"
)

func UploadFile(filePath string, filename string, token string) string {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		fmt.Println("Error creating form file:", err)
		return ""
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error copying file content:", err)
		return ""
	}

	err = writer.Close()
	if err != nil {
		fmt.Println("Error closing writer:", err)
		return ""
	}

	req, err := http.NewRequest("POST", "https://api.interrupted.me/upload", &b)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "ShareX/16.1.0")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ""
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	var fileResponse types.FileResponse
	err = json.Unmarshal(responseBody, &fileResponse)
	if err != nil {
		fmt.Printf("error unmarshalling response body: %s\n", err)
		os.Exit(1)
	}

	if fileResponse.Success {
		return fileResponse.IOS
	} else {
		fmt.Println("Upload failed")
		return ""
	}
}

func CaptureAndSaveScreenshot(displayIndex int, fileName string) (string, error) {
	numDisplays := screenshot.NumActiveDisplays()
	if displayIndex >= numDisplays || displayIndex < 0 {
		return "", fmt.Errorf("invalid display index")
	}

	bounds := screenshot.GetDisplayBounds(displayIndex)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", fmt.Errorf("error capturing screenshot: %v", err)
	}

	appDataPath, err := data.GetAppDataPath()
	if err != nil {
		return "", fmt.Errorf("error getting app data path: %v", err)
	}

	savePath := filepath.Join(appDataPath, data.Subdirectory, fileName)
	os.MkdirAll(filepath.Dir(savePath), os.ModePerm)

	file, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("error creating screenshot file: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("error encoding screenshot to PNG: %v", err)
	}

	return savePath, nil
}

func UploadAndCopyToClipboard(filePath, fileName, apiKey string) (string, error) {
	url := UploadFile(filePath, fileName, apiKey)
	if url == "" {
		return "", fmt.Errorf("error uploading file")
	}
	clipboard.Write(clipboard.FmtText, []byte(url))
	return fmt.Sprintf("File '%s' uploaded and URL copied to clipboard!", fileName), nil
}
