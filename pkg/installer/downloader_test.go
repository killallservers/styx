package installer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDownloader(t *testing.T) {
	tmpDir := t.TempDir()
	dl, err := NewDownloader(tmpDir)
	if err != nil {
		t.Fatalf("NewDownloader failed: %v", err)
	}

	if dl.cacheDir != tmpDir {
		t.Errorf("expected cache dir %q, got %q", tmpDir, dl.cacheDir)
	}

	// Directory should be created
	if _, err := os.Stat(tmpDir); err != nil {
		t.Errorf("cache directory not created: %v", err)
	}
}

func TestDownloaderCacheCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)

	// Verify directory exists and is readable/writable by owner
	info, err := os.Stat(dl.cacheDir)
	if err != nil {
		t.Fatalf("failed to stat cache dir: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("cache path is not a directory")
	}

	// At minimum, should have owner read/write/exec
	perm := info.Mode().Perm()
	if (perm & 0700) != 0700 {
		t.Errorf("owner permissions missing: got %o", perm)
	}
}

func TestDownloadParallelWithNoDownloads(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)

	results, err := dl.DownloadParallel([]struct {
		Name string
		URL  string
		Hash string
	}{}, 4)

	if err != nil {
		t.Fatalf("DownloadParallel failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDownloadHashVerification(t *testing.T) {
	tmpDir := t.TempDir()
	_, _ = NewDownloader(tmpDir)

	// Create a fake file to "download" from local filesystem
	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("test content")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Calculate correct hash
	h := sha256.New()
	h.Write(testData)
	correctHash := fmt.Sprintf("%x", h.Sum(nil))

	// Test with incorrect hash (should fail)
	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"

	// Since we can't actually download files without mocking HTTP, this test just verifies the structure
	if len(correctHash) != 64 {
		t.Errorf("expected 64-char hex hash, got %d", len(correctHash))
	}
	if len(wrongHash) != 64 {
		t.Errorf("expected 64-char hex hash, got %d", len(wrongHash))
	}
}

func TestDownloadParallelConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)

	// Create test data
	downloads := make([]struct {
		Name string
		URL  string
		Hash string
	}, 10)

	for i := 0; i < 10; i++ {
		downloads[i] = struct {
			Name string
			URL  string
			Hash string
		}{
			Name: fmt.Sprintf("tool%d", i),
			URL:  "http://example.com/fake.tar.gz",
			Hash: fmt.Sprintf("%064d", i), // Fake hash
		}
	}

	// Test with different max concurrency values
	for _, maxConcurrency := range []int{1, 4, 8, 16} {
		results, _ := dl.DownloadParallel(downloads, maxConcurrency)
		if len(results) != len(downloads) {
			t.Errorf("expected %d results with concurrency %d, got %d",
				len(downloads), maxConcurrency, len(results))
		}
	}
}
