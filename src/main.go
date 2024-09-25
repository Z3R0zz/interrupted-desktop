package main

import (
	"interrupted-desktop/src/data"
	"interrupted-desktop/src/views"
	"log"

	"golang.design/x/clipboard"
)

func main() {
	if err := clipboard.Init(); err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	apiKey, err := data.ReadApiKey()
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	if apiKey == "" {
		apiKey = views.ShowLoginView()
		if apiKey == "" {
			log.Println("Login failed. Exiting application.")
			return
		}
	}

	if err := data.ClearAppData(); err != nil {
		log.Printf("Warning: %v", err)
	}

	views.ShowDefaultView(apiKey)
}
