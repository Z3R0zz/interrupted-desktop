package data

import (
	"encoding/json"
	"fmt"
	"interrupted-desktop/src/types"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

const (
	Subdirectory   = ".interrupted"
	apiKeyFileName = "api_key"
)

func GetAppDataPath() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(user.HomeDir, "AppData", "Roaming"), nil
}

func SaveApiKey(apiKey string) error {
	appDataPath, err := GetAppDataPath()
	if err != nil {
		return err
	}

	apiKeyDir := filepath.Join(appDataPath, Subdirectory)

	if err := os.MkdirAll(apiKeyDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	apiKeyPath := filepath.Join(apiKeyDir, apiKeyFileName)
	return os.WriteFile(apiKeyPath, []byte(apiKey), 0644)
}

func ReadApiKey() (string, error) {
	appDataPath, err := GetAppDataPath()
	if err != nil {
		return "", err
	}

	apiKeyPath := filepath.Join(appDataPath, Subdirectory, apiKeyFileName)

	if _, err := os.Stat(apiKeyPath); os.IsNotExist(err) {
		return "", nil
	}

	data, err := os.ReadFile(apiKeyPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func PromptForApiKey() (string, error) {
	var apiKey string
	fmt.Print("Enter your API key: ")
	fmt.Scanln(&apiKey)
	return apiKey, nil
}

func DeleteApiKey() error {
	appDataPath, err := GetAppDataPath()
	if err != nil {
		return err
	}

	apiKeyPath := filepath.Join(appDataPath, Subdirectory, apiKeyFileName)

	if _, err := os.Stat(apiKeyPath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(apiKeyPath)
}

func GetUserData(apiKey string) types.User {
	requestURL := fmt.Sprintf("http://127.0.0.1:8000/api/whois/%v", apiKey)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("error reading response body: %s\n", err)
		os.Exit(1)
	}

	var apiResponse types.ApiResponse

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("error unmarshalling response body: %s\n", err)
		os.Exit(1)
	}

	if res.StatusCode != 200 {
		if apiResponse.Message != nil {
			fmt.Printf("API error: %s\n", *apiResponse.Message)
		} else {
			fmt.Printf("API error: unknown error\n")
		}
		DeleteApiKey()
		os.Exit(1)
	}

	user := apiResponse.Data
	return user
}
