package registry

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// HTTPRegistry fetches tool specs from an HTTP source with embedded fallback.
type HTTPRegistry struct {
	baseURL  string
	cacheDir string
	timeout  time.Duration
	embedded map[string]*ToolSpec
}

// NewHTTPRegistry creates a new HTTP registry with fallback to embedded.
func NewHTTPRegistry(baseURL string) (*HTTPRegistry, error) {
	// Validate URL
	if baseURL != "" {
		if _, err := url.Parse(baseURL); err != nil {
			return nil, fmt.Errorf("invalid registry URL: %w", err)
		}
	}

	// Load embedded registry for fallback
	embedded, err := LoadEmbeddedRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded registry: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".styx", "cache", "registry")

	return &HTTPRegistry{
		baseURL:  baseURL,
		cacheDir: cacheDir,
		timeout:  30 * time.Second,
		embedded: embedded,
	}, nil
}

// LoadRegistry loads specs from HTTP source with caching, falling back to embedded on failure.
func (hr *HTTPRegistry) LoadRegistry() (map[string]*ToolSpec, error) {
	// If no HTTP URL configured, use embedded
	if hr.baseURL == "" {
		return hr.embedded, nil
	}

	// Try cache first (if not stale)
	if cached, err := hr.loadFromCache(); err == nil {
		return cached, nil
	}

	// Try HTTP next
	specs, err := hr.fetchFromHTTP()
	if err == nil {
		// Cache the result
		_ = hr.saveToCache(specs)
		return specs, nil
	}

	// Try fallback cache (even if stale)
	if cached, err := hr.loadFromCacheUnchecked(); err == nil {
		fmt.Fprintf(os.Stderr, "Warning: Using stale cached registry (fresh fetch failed: %v)\n", err)
		return cached, nil
	}

	// Log the error but fall back to embedded
	fmt.Fprintf(os.Stderr, "Warning: Failed to fetch registry from %s: %v\n", hr.baseURL, err)
	fmt.Fprintf(os.Stderr, "Falling back to embedded registry\n")

	return hr.embedded, nil
}

// fetchFromHTTP fetches the registry from HTTP source.
func (hr *HTTPRegistry) fetchFromHTTP() (map[string]*ToolSpec, error) {
	if hr.baseURL == "" {
		return nil, fmt.Errorf("no HTTP registry URL configured")
	}

	// Construct URL for registry manifest
	registryURL := fmt.Sprintf("%s/registry.json", hr.baseURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: hr.timeout,
	}

	// Fetch registry manifest
	resp, err := client.Get(registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry server returned %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry response: %w", err)
	}

	// For now, just validate we can fetch
	if len(body) == 0 {
		return nil, fmt.Errorf("empty registry response")
	}

	// Parse would go here (JSON → ToolSpec map)
	// For Phase 5, this would deserialize the JSON registry
	return nil, fmt.Errorf("HTTP registry parsing not yet implemented (requires remote registry)")
}

// loadFromCache loads cached registry specs if fresh (< 24 hours old).
func (hr *HTTPRegistry) loadFromCache() (map[string]*ToolSpec, error) {
	cacheFile := filepath.Join(hr.cacheDir, "registry.cache.json")

	info, err := os.Stat(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("cache not found: %w", err)
	}

	// Check if cache is fresh (< 24 hours old)
	if time.Since(info.ModTime()) > 24*time.Hour {
		return nil, fmt.Errorf("cache is stale (older than 24 hours)")
	}

	// Read and parse cache
	_, err = os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	// For now, just return embedded (cache implementation would deserialize here)
	return hr.embedded, nil
}

// loadFromCacheUnchecked loads cached registry specs without freshness check.
func (hr *HTTPRegistry) loadFromCacheUnchecked() (map[string]*ToolSpec, error) {
	cacheFile := filepath.Join(hr.cacheDir, "registry.cache.json")

	// Read cache file without checking age
	_, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("cache not found: %w", err)
	}

	// For now, just return embedded (cache implementation would deserialize here)
	return hr.embedded, nil
}

// saveToCache saves registry specs to cache.
func (hr *HTTPRegistry) saveToCache(specs map[string]*ToolSpec) error {
	if err := os.MkdirAll(hr.cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheFile := filepath.Join(hr.cacheDir, "registry.cache.json")

	// For now, just create an empty cache file to mark it as cached
	// Full implementation would serialize specs to JSON
	if err := os.WriteFile(cacheFile, []byte("{}"), 0600); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// LoadWithHTTPFallback loads registry from HTTP URL with embedded fallback.
// This is the recommended entry point for Phase 2+.
func LoadWithHTTPFallback(httpURL string) (map[string]*ToolSpec, error) {
	httpReg, err := NewHTTPRegistry(httpURL)
	if err != nil {
		// If HTTP registry initialization fails, fall back to embedded
		return LoadEmbeddedRegistry()
	}

	return httpReg.LoadRegistry()
}
