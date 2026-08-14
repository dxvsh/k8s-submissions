package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	cacheDuration = 10 * time.Minute
	imagePath     = "cache/cached-picsum-image.jpg"
	picsumURL     = "https://picsum.photos/500"
)

var (
	imageCacheMu sync.Mutex
	httpClient   = &http.Client{Timeout: 15 * time.Second}
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/image", handleImage)

	log.Printf("Server started in port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func handleImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := refreshImageCacheIfNeeded(); err != nil {
		log.Printf("Failed to refresh image cache: %v", err)

		if _, statErr := os.Stat(imagePath); statErr != nil {
			http.Error(w, "Image unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, imagePath)
}

func refreshImageCacheIfNeeded() error {
	imageCacheMu.Lock()
	defer imageCacheMu.Unlock()

	fileInfo, err := os.Stat(imagePath)
	if err == nil && time.Since(fileInfo.ModTime()) < cacheDuration {
		return nil
	}

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("checking cached image: %w", err)
	}

	return downloadImage()
}

func downloadImage() error {
	response, err := httpClient.Get(picsumURL)
	if err != nil {
		return fmt.Errorf("downloading image: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading image: unexpected status %s", response.Status)
	}

	temporaryPath := imagePath + ".tmp"
	file, err := os.Create(temporaryPath)
	if err != nil {
		return fmt.Errorf("creating temporary image file: %w", err)
	}

	_, copyErr := io.Copy(file, response.Body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("saving image: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("closing temporary image file: %w", closeErr)
	}

	if err := os.Rename(temporaryPath, imagePath); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("replacing cached image: %w", err)
	}

	return nil
}
