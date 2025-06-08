package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	UI       UIConfig       `json:"ui"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	PoolSize int    `json:"pool_size"`
}

type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

type UIConfig struct {
	ThemeDark      string            `json:"theme_dark"`
	ThemeLight     string            `json:"theme_light"`
	CurrentTheme   string            `json:"current_theme"`
	Colors         map[string]string `json:"colors"`
	AutoSave       bool              `json:"auto_save"`
	AutoSaveInterval int             `json:"auto_save_interval"`
	CacheTimeout   int               `json:"cache_timeout"`
	WindowWidth    int               `json:"window_width"`
	WindowHeight   int               `json:"window_height"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return getDefaultConfig(), nil // Return default config if file doesn't exist
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}

func getDefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "jobscanner",
			Username: "root",
			Password: "",
			PoolSize: 5,
		},
		Server: ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		UI: UIConfig{
			ThemeDark:        "darkly",
			ThemeLight:       "flatly",
			CurrentTheme:     "dark",
			AutoSave:         true,
			AutoSaveInterval: 300,
			CacheTimeout:     300,
			WindowWidth:      1400,
			WindowHeight:     800,
			Colors: map[string]string{
				"primary":     "#007bff",
				"background":  "#ffffff",
				"text":        "#000000",
				"selection":   "#e9ecef",
				"success":     "#28a745",
				"error":       "#dc3545",
				"warning":     "#ffc107",
				"dark_bg":     "#2b2b2b",
				"dark_text":   "#ffffff",
			},
		},
	}
}