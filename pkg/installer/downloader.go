package installer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Downloader handles binary downloads and verification.
type Downloader struct {
	cacheDir string
	client   *http.Client
}

// NewDownloader creates a new downloader.
func NewDownloader(cacheDir string) (*Downloader, error) {
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &Downloader{
		cacheDir: cacheDir,
		client:   &http.Client{},
	}, nil
}

// DownloadResult holds the result of a download.
type DownloadResult struct {
	Name  string
	URL   string
	Path  string
	Hash  string
	Error error
}

// Download downloads a binary from the given URL and verifies its checksum.
// It stores the file in the cache directory under the hash.
func (d *Downloader) Download(url string, expectedHash string) (path string, actualHash string, err error) {
	resp, err := d.client.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download failed: HTTP %d from %s", resp.StatusCode, url)
	}

	h := sha256.New()
	tempFile := filepath.Join(d.cacheDir, ".tmp", filepath.Base(url))
	if err := os.MkdirAll(filepath.Dir(tempFile), 0700); err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	f, err := os.Create(tempFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(io.MultiWriter(f, h), resp.Body); err != nil {
		os.Remove(tempFile)
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	actualHash = fmt.Sprintf("%x", h.Sum(nil))

	if actualHash != expectedHash {
		os.Remove(tempFile)
		return "", "", fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	finalDir := filepath.Join(d.cacheDir, actualHash[:8])
	if err := os.MkdirAll(finalDir, 0700); err != nil {
		os.Remove(tempFile)
		return "", "", fmt.Errorf("failed to create final directory: %w", err)
	}

	finalPath := filepath.Join(finalDir, filepath.Base(url))
	if err := os.Rename(tempFile, finalPath); err != nil {
		os.Remove(tempFile)
		return "", "", fmt.Errorf("failed to move file to cache: %w", err)
	}

	return finalPath, actualHash, nil
}

// DownloadParallel downloads multiple files in parallel.
func (d *Downloader) DownloadParallel(downloads []struct {
	Name string
	URL  string
	Hash string
}, maxConcurrency int) ([]DownloadResult, error) {
	results := make([]DownloadResult, len(downloads))
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, dl := range downloads {
		wg.Add(1)
		go func(idx int, name, url, hash string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			path, actualHash, err := d.Download(url, hash)
			results[idx] = DownloadResult{
				Name:  name,
				URL:   url,
				Path:  path,
				Hash:  actualHash,
				Error: err,
			}
		}(i, dl.Name, dl.URL, dl.Hash)
	}

	wg.Wait()
	return results, nil
}
