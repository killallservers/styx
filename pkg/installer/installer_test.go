package installer

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/killallservers/styx/pkg/registry"
)

func TestNewInstaller(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()

	inst, err := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), filepath.Join(tmpDir, "bin"))
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	if inst.registry == nil {
		t.Error("installer registry is nil")
	}
	if inst.downloader == nil {
		t.Error("installer downloader is nil")
	}
}

func TestExtractTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), filepath.Join(tmpDir, "bin"))

	// Create a test tar.gz file
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	f, _ := os.Create(tarPath)

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	// Add a test file to the archive
	testContent := []byte("test binary content")
	header := &tar.Header{
		Name: "bin/testbinary",
		Size: int64(len(testContent)),
		Mode: 0755,
	}
	tw.WriteHeader(header)
	tw.Write(testContent)

	// Close all writers to flush
	tw.Close()
	gz.Close()
	f.Close()

	// Extract it
	extractPath := filepath.Join(tmpDir, "extracted")
	err := inst.extractTarGz(tarPath, extractPath)
	if err != nil {
		t.Fatalf("extractTarGz failed: %v", err)
	}

	// Verify file was extracted
	extractedFile := filepath.Join(extractPath, "bin/testbinary")
	if _, err := os.Stat(extractedFile); err != nil {
		t.Errorf("extracted file not found: %v", err)
	}

	// Verify content
	content, _ := os.ReadFile(extractedFile)
	if string(content) != string(testContent) {
		t.Errorf("extracted content mismatch")
	}
}

func TestExtractTarGzDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), filepath.Join(tmpDir, "bin"))

	// Create a test tar.gz with directory structure
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	f, _ := os.Create(tarPath)

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	// Add a directory
	dirHeader := &tar.Header{
		Name:     "mydir/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}
	tw.WriteHeader(dirHeader)

	// Add a file in the directory
	content := []byte("file content")
	fileHeader := &tar.Header{
		Name: "mydir/file.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	tw.WriteHeader(fileHeader)
	tw.Write(content)

	// Close all writers to flush
	tw.Close()
	gz.Close()
	f.Close()

	// Extract
	extractPath := filepath.Join(tmpDir, "extracted")
	err := inst.extractTarGz(tarPath, extractPath)
	if err != nil {
		t.Fatalf("extractTarGz failed: %v", err)
	}

	// Verify directory
	dirPath := filepath.Join(extractPath, "mydir")
	if info, err := os.Stat(dirPath); err != nil || !info.IsDir() {
		t.Error("directory not extracted correctly")
	}

	// Verify file
	filePath := filepath.Join(extractPath, "mydir/file.txt")
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("file not extracted: %v", err)
	}
}

func TestExtractInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), filepath.Join(tmpDir, "bin"))

	// Try to extract with unsupported format
	_, err := os.Create(filepath.Join(tmpDir, "test.unknown"))
	extractPath := filepath.Join(tmpDir, "extracted")

	err = inst.extract(filepath.Join(tmpDir, "test.unknown"), extractPath, "test.unknown")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestCreateSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	binDir := filepath.Join(tmpDir, "bin")
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), binDir)

	// Create a test executable in storage
	storePath := filepath.Join(tmpDir, "store", "mytool", "1.0.0")
	os.MkdirAll(storePath, 0755)
	execPath := filepath.Join(storePath, "mytool")
	os.WriteFile(execPath, []byte("#!/bin/sh\necho test"), 0755)

	// Create symlink
	err := inst.createSymlink("mytool", storePath, "mytool")
	if err != nil {
		t.Fatalf("createSymlink failed: %v", err)
	}

	// Verify symlink exists
	linkPath := filepath.Join(binDir, "mytool")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Errorf("symlink not created: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("created file is not a symlink")
	}

	// Verify symlink points to correct location
	target, _ := os.Readlink(linkPath)
	if target != execPath {
		t.Errorf("symlink target mismatch: expected %s, got %s", execPath, target)
	}
}

func TestCreateSymlinkReplaceExisting(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	binDir := filepath.Join(tmpDir, "bin")
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), binDir)

	os.MkdirAll(binDir, 0755)

	// Create an existing symlink
	storePath1 := filepath.Join(tmpDir, "store", "mytool", "1.0.0")
	os.MkdirAll(storePath1, 0755)
	execPath1 := filepath.Join(storePath1, "mytool")
	os.WriteFile(execPath1, []byte("v1"), 0755)

	linkPath := filepath.Join(binDir, "mytool")
	os.Symlink(execPath1, linkPath)

	// Create new version and symlink
	storePath2 := filepath.Join(tmpDir, "store", "mytool", "2.0.0")
	os.MkdirAll(storePath2, 0755)
	execPath2 := filepath.Join(storePath2, "mytool")
	os.WriteFile(execPath2, []byte("v2"), 0755)

	err := inst.createSymlink("mytool", storePath2, "mytool")
	if err != nil {
		t.Fatalf("createSymlink failed: %v", err)
	}

	// Verify symlink points to new location
	target, _ := os.Readlink(linkPath)
	if target != execPath2 {
		t.Errorf("symlink not updated: expected %s, got %s", execPath2, target)
	}
}

func TestInstallMissingTool(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), filepath.Join(tmpDir, "bin"))

	// Try to install non-existent tool
	result, _ := inst.Install("nonexistent-tool", "1.0.0")
	if result.Error == nil {
		t.Error("expected error for missing tool")
	}
}

func TestInstallMissingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), filepath.Join(tmpDir, "bin"))

	// Try to install non-existent version
	result, _ := inst.Install("ripgrep", "99.99.99")
	if result.Error == nil {
		t.Error("expected error for missing version")
	}
}

func TestInstallResultStructure(t *testing.T) {
	tmpDir := t.TempDir()
	dl, _ := NewDownloader(tmpDir)
	reg, _ := registry.LoadEmbeddedRegistry()
	inst, _ := NewInstaller(reg, dl, filepath.Join(tmpDir, "store"), filepath.Join(tmpDir, "bin"))

	result, _ := inst.Install("ripgrep", "14.1.0")

	if result.Name != "ripgrep" {
		t.Errorf("expected Name=ripgrep, got %s", result.Name)
	}
	if result.Version != "14.1.0" {
		t.Errorf("expected Version=14.1.0, got %s", result.Version)
	}
	// Error is expected since we can't actually download
	if result.Error == nil {
		t.Error("expected error due to network requirements")
	}
}
