package data

import (
	"encoding/json"
	"fmt"
	"interrupted-desktop/src/types"
	"interrupted-desktop/src/utils"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
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

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(user.HomeDir, "AppData", "Roaming"), nil
	case "darwin":
		return filepath.Join(user.HomeDir, "Library", "Application Support"), nil
	case "linux":
		return filepath.Join(user.HomeDir, ".config"), nil
	default:
		return "", os.ErrNotExist
	}
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
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (compatible; interrupted/1.0; +https://interrupted.me)",
	}

	requestURL := fmt.Sprintf("https://api.intrd.me/api/whois/%v", apiKey)

	resp, err := utils.SendAPIRequest("GET", requestURL, nil, headers)
	if err != nil {
		fmt.Printf("error performing request: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

	if resp.StatusCode != 200 {
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

func GetUserStats(apiKey string) types.Stats {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (compatible; interrupted/1.0; +https://interrupted.me)",
	}

	requestURL := fmt.Sprintf("https://api.intrd.me/api/stats/%v", apiKey)

	resp, err := utils.SendAPIRequest("GET", requestURL, nil, headers)
	if err != nil {
		fmt.Printf("error performing request: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response body: %s\n", err)
		os.Exit(1)
	}

	var statsResponse types.StatsResponse

	err = json.Unmarshal(body, &statsResponse)
	if err != nil {
		fmt.Printf("error unmarshalling response body: %s\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		if statsResponse.Message != nil {
			fmt.Printf("API error: %s\n", *statsResponse.Message)
		} else {
			fmt.Printf("API error: unknown error\n")
		}
		DeleteApiKey()
		os.Exit(1)
	}

	stats := statsResponse.Data
	return stats
}

func ClearAppData() error {
	appDataPath, err := GetAppDataPath()
	if err != nil {
		return err
	}

	appDataDir := filepath.Join(appDataPath, Subdirectory)

	if _, err := os.Stat(appDataDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(appDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() != "api_key" {
			return os.Remove(path)
		}
		return nil
	})
}
