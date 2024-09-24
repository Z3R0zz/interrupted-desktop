package uploads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"interrupted-desktop/src/types"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func UploadFile(filePath string, token string) string {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", file.Name())
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
