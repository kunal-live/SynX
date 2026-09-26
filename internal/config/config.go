package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type DeviceConfig struct {
	Name string `json:"name"`
}

type NetworkConfig struct {
	Port          int `json:"port"`
	DiscoveryPort int `json:"discovery_port"`
}

type StorageConfig struct {
	SharedDirectory   string `json:"shared_directory"`
	DownloadDirectory string `json:"download_directory"`
	DataDirectory     string `json:"data_directory"`
	DatabasePath      string `json:"database_path"`
}

type SecurityConfig struct {
	RequirePairing bool   `json:"require_pairing"`
	AutoAccept     bool   `json:"auto_accept"`
	LocalToken     string `json:"local_token"`
}

type TransferConfig struct {
	ChunkSize            int64 `json:"chunk_size"`
	MaxParallelChunks    int   `json:"max_parallel_chunks"`
	MaxParallelTransfers int   `json:"max_parallel_transfers"`
}

type Config struct {
	mu       sync.RWMutex   `json:"-"`
	Device   DeviceConfig   `json:"device"`
	Network  NetworkConfig  `json:"network"`
	Storage  StorageConfig  `json:"storage"`
	Security SecurityConfig `json:"security"`
	Transfer TransferConfig `json:"transfer"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	dataDir := AppDataDir()
	_ = os.MkdirAll(dataDir, 0755)

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "SynX-Node"
	}

	sharedDir := filepath.Join(home, "SynX")
	downloadDir := filepath.Join(home, "Downloads", "SynX")
	_ = os.MkdirAll(sharedDir, 0755)
	_ = os.MkdirAll(downloadDir, 0755)

	return &Config{
		Device: DeviceConfig{
			Name: hostname,
		},
		Network: NetworkConfig{
			Port:          8787,
			DiscoveryPort: 8788,
		},
		Storage: StorageConfig{
			SharedDirectory:   sharedDir,
			DownloadDirectory: downloadDir,
			DataDirectory:     dataDir,
			DatabasePath:      filepath.Join(dataDir, "synx.db"),
		},
		Security: SecurityConfig{
			RequirePairing: true,
			AutoAccept:     false,
		},
		Transfer: TransferConfig{
			ChunkSize:            8 * 1024 * 1024, // 8 MB default
			MaxParallelChunks:    4,
			MaxParallelTransfers: 2,
		},
	}
}

func AppDataDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			return filepath.Join(home, "AppData", "Roaming", "SynX")
		}
		return filepath.Join(appData, "SynX")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "SynX")
	default:
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".config", "synx")
		}
		return filepath.Join(configHome, "synx")
	}
}

func ConfigFilePath() string {
	return filepath.Join(AppDataDir(), "config.json")
}

func Load() (*Config, error) {
	cfg := DefaultConfig()
	path := ConfigFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			_ = cfg.Save()
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, nil
	}

	// Ensure essential folders exist
	_ = os.MkdirAll(cfg.Storage.SharedDirectory, 0755)
	_ = os.MkdirAll(cfg.Storage.DownloadDirectory, 0755)
	_ = os.MkdirAll(cfg.Storage.DataDirectory, 0755)

	return cfg, nil
}

func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := ConfigFilePath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) SetSharedDir(dir string) error {
	c.mu.Lock()
	c.Storage.SharedDirectory = dir
	c.mu.Unlock()
	return c.Save()
}
