package lock

import (
	"encoding/json"
	"time"
)

type LockFile struct {
	Version          string            `json:"version"`
	GeneratedAt      time.Time         `json:"generated_at"`
	RegistrySnapshot Registry          `json:"registry_snapshot"`
	Tools            []ToolEntry       `json:"tools"`
	Env              map[string]string `json:"env"`
}

type Registry struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

type ToolEntry struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	InstallMethod string `json:"install_method"`
	StoragePath   string `json:"storage_path"`
	BinaryHash    string `json:"binary_hash_sha256"`
	Executable    string `json:"executable"`
	SourceConfig  string `json:"source_config"`
}

func New() *LockFile {
	return &LockFile{
		Version:     "1.0",
		GeneratedAt: time.Now(),
		Tools:       []ToolEntry{},
		Env:         make(map[string]string),
	}
}

func (lf *LockFile) MarshalJSON() ([]byte, error) {
	type Alias LockFile
	return json.MarshalIndent(&struct {
		*Alias
	}{
		Alias: (*Alias)(lf),
	}, "", "  ")
}

func (lf *LockFile) FindTool(name string) *ToolEntry {
	for i := range lf.Tools {
		if lf.Tools[i].Name == name {
			return &lf.Tools[i]
		}
	}
	return nil
}

func (lf *LockFile) AddTool(entry ToolEntry) {
	lf.Tools = append(lf.Tools, entry)
}
