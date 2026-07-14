package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/killallservers/styx/pkg/lock"
	"github.com/killallservers/styx/pkg/platform"
	"github.com/killallservers/styx/pkg/registry"
)

// Installer orchestrates tool installation: resolve → download → verify → extract → symlink.
type Installer struct {
	registry   map[string]*registry.ToolSpec
	downloader *Downloader
	storeDir   string
	binDir     string
	platform   string
}

// NewInstaller creates a new installer.
func NewInstaller(reg map[string]*registry.ToolSpec, downloader *Downloader, storeDir, binDir string) (*Installer, error) {
	p := platform.Detect()

	return &Installer{
		registry:   reg,
		downloader: downloader,
		storeDir:   storeDir,
		binDir:     binDir,
		platform:   p.String(),
	}, nil
}

// InstallResult holds the result of a tool installation.
type InstallResult struct {
	Name       string
	Version    string
	Method     string
	StorePath  string
	BinaryHash string
	Executable string
	Error      error
}

// Install installs a single tool.
func (inst *Installer) Install(toolName, version string) (*InstallResult, error) {
	toolSpec, ok := inst.registry[toolName]
	if !ok {
		return &InstallResult{
			Name:  toolName,
			Error: fmt.Errorf("tool %q not found in registry", toolName),
		}, nil
	}

	versionSpec := toolSpec.GetVersion(version)
	if versionSpec == nil {
		return &InstallResult{
			Name:    toolName,
			Version: version,
			Error:   fmt.Errorf("version %q not found for tool %q", version, toolName),
		}, nil
	}

	method := versionSpec.GetMethodForPlatform(inst.platform, "binary")
	if method == nil {
		return &InstallResult{
			Name:    toolName,
			Version: version,
			Error:   fmt.Errorf("no binary method available for %s on %s", toolName, inst.platform),
		}, nil
	}

	filename, ok := method.Platforms[inst.platform]
	if !ok {
		return &InstallResult{
			Name:    toolName,
			Version: version,
			Error:   fmt.Errorf("no binary available for %s on %s", toolName, inst.platform),
		}, nil
	}

	expectedHash, ok := method.Checksums[inst.platform]
	if !ok {
		return &InstallResult{
			Name:    toolName,
			Version: version,
			Error:   fmt.Errorf("no checksum available for %s on %s", toolName, inst.platform),
		}, nil
	}

	// For MVP, use a placeholder URL that would come from registry in Phase 2
	url := fmt.Sprintf("https://releases.example.com/%s/%s", toolName, filename)

	// Download
	path, hash, err := inst.downloader.Download(url, expectedHash)
	if err != nil {
		return &InstallResult{
			Name:    toolName,
			Version: version,
			Error:   fmt.Errorf("download failed: %w", err),
		}, nil
	}

	// Extract and get executable
	executable := method.Executable
	storePath := filepath.Join(inst.storeDir, toolName, version)

	if err := inst.extract(path, storePath, filename); err != nil {
		return &InstallResult{
			Name:    toolName,
			Version: version,
			Error:   fmt.Errorf("extraction failed: %w", err),
		}, nil
	}

	// Create symlink in bin directory
	if err := inst.createSymlink(toolName, storePath, executable); err != nil {
		return &InstallResult{
			Name:    toolName,
			Version: version,
			Error:   fmt.Errorf("symlink creation failed: %w", err),
		}, nil
	}

	return &InstallResult{
		Name:       toolName,
		Version:    version,
		Method:     "binary",
		StorePath:  storePath,
		BinaryHash: hash,
		Executable: executable,
	}, nil
}

// InstallAll installs multiple tools.
func (inst *Installer) InstallAll(tools map[string]string) ([]InstallResult, error) {
	results := make([]InstallResult, 0)

	for toolName, version := range tools {
		result, _ := inst.Install(toolName, version)
		if result != nil {
			results = append(results, *result)
		}
	}

	return results, nil
}

// extract extracts a tar.gz or zip file to the storage path.
func (inst *Installer) extract(filePath string, storePath string, filename string) error {
	if err := os.MkdirAll(storePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	if strings.HasSuffix(filename, ".tar.gz") {
		return inst.extractTarGz(filePath, storePath)
	} else if strings.HasSuffix(filename, ".tar.xz") {
		return fmt.Errorf("tar.xz extraction not yet implemented")
	} else if strings.HasSuffix(filename, ".zip") {
		return inst.extractZip(filePath, storePath)
	}

	return fmt.Errorf("unsupported archive format: %s", filename)
}

// extractTarGz extracts a tar.gz file.
func (inst *Installer) extractTarGz(filePath, storePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		path := filepath.Join(storePath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}
			f.Close()
		}
	}

	return nil
}

// extractZip extracts a zip file.
func (inst *Installer) extractZip(filePath, storePath string) error {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(storePath, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		rc, err := file.Open()
		if err != nil {
			f.Close()
			return fmt.Errorf("failed to open zip entry: %w", err)
		}

		if _, err := io.Copy(f, rc); err != nil {
			f.Close()
			rc.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}
		f.Close()
		rc.Close()
	}

	return nil
}

// createSymlink creates a symlink from bin directory to the executable in storage.
func (inst *Installer) createSymlink(toolName string, storePath string, executable string) error {
	if err := os.MkdirAll(inst.binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Find the executable in the storage path
	var execPath string
	err := filepath.Walk(storePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, executable) {
			execPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to find executable: %w", err)
	}

	if execPath == "" {
		return fmt.Errorf("executable %q not found in %s", executable, storePath)
	}

	linkPath := filepath.Join(inst.binDir, toolName)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing symlink: %w", err)
	}

	if err := os.Symlink(execPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// InstallFromLock installs a tool using data from a lock file entry.
// Used by styx sync to reproduce exact environment from lock file.
func (inst *Installer) InstallFromLock(toolEntry lock.ToolEntry, lockFile *lock.LockFile) (*InstallResult, error) {
	// Download from URL in lock file
	path, hash, err := inst.downloader.Download(toolEntry.StoragePath, toolEntry.BinaryHash)
	if err != nil {
		return &InstallResult{
			Name:    toolEntry.Name,
			Version: toolEntry.Version,
			Error:   fmt.Errorf("download failed: %w", err),
		}, nil
	}

	// Verify checksum matches lock
	if hash != toolEntry.BinaryHash {
		return &InstallResult{
			Name:    toolEntry.Name,
			Version: toolEntry.Version,
			Error:   fmt.Errorf("checksum mismatch: expected %s, got %s", toolEntry.BinaryHash, hash),
		}, nil
	}

	// Extract to storage path
	storePath := filepath.Join(inst.storeDir, toolEntry.StoragePath)
	if err := inst.extract(path, storePath, toolEntry.StoragePath); err != nil {
		return &InstallResult{
			Name:    toolEntry.Name,
			Version: toolEntry.Version,
			Error:   fmt.Errorf("extraction failed: %w", err),
		}, nil
	}

	// Create symlink in bin directory
	if err := inst.createSymlink(toolEntry.Name, storePath, toolEntry.Executable); err != nil {
		return &InstallResult{
			Name:    toolEntry.Name,
			Version: toolEntry.Version,
			Error:   fmt.Errorf("symlink creation failed: %w", err),
		}, nil
	}

	return &InstallResult{
		Name:       toolEntry.Name,
		Version:    toolEntry.Version,
		Method:     toolEntry.InstallMethod,
		StorePath:  storePath,
		BinaryHash: hash,
		Executable: toolEntry.Executable,
	}, nil
}
