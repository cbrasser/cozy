package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Library          LibraryConfig `toml:"library"`
	ThemeName        string        `toml:"theme_name"` // Name of theme to load
	Reading          ReadingConfig `toml:"reading"`
	Display          DisplayConfig `toml:"display"`
	DataDir          string        `toml:"data_dir"`           // Directory for app data (bookmarks, progress, etc.)
	UseLibraryForData bool          `toml:"use_library_for_data"` // If true, store data in library path

	// Active theme (loaded at runtime, not saved to file)
	ActiveTheme *Theme `toml:"-"`
}

type LibraryConfig struct {
	Path string `toml:"path"`
}

type ReadingConfig struct {
	CurrentBook string `toml:"current_book"`
	Position    int    `toml:"position"`
}

type DisplayConfig struct {
	FontSize    int `toml:"font_size"`
	LineSpacing int `toml:"line_spacing"`
	MarginLeft  int `toml:"margin_left"`
	MarginRight int `toml:"margin_right"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	defaultTheme := CozyDark
	configDir := filepath.Join(homeDir, ".config", "cozy")

	return Config{
		Library: LibraryConfig{
			Path: filepath.Join(homeDir, "Documents", "Books"),
		},
		ThemeName:        "cozy-dark",
		DataDir:          filepath.Join(configDir, "data"),
		UseLibraryForData: false,
		Reading: ReadingConfig{
			CurrentBook: "",
			Position:    0,
		},
		Display: DisplayConfig{
			FontSize:    14,
			LineSpacing: 2,
			MarginLeft:  4,
			MarginRight: 4,
		},
		ActiveTheme: &defaultTheme,
	}
}

// ConfigDir returns the path to the config directory
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "cozy")
	return configDir, nil
}

// ConfigPath returns the full path to the config file
func ConfigPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.toml"), nil
}

// DataDir returns the path to the data directory based on config
func (c *Config) DataDirectory() string {
	if c.UseLibraryForData {
		// Use hidden folder in library path
		return filepath.Join(c.Library.Path, ".cozy")
	}
	// Use configured data directory (defaults to ~/.config/cozy/data)
	return c.DataDir
}

// EnsureDataDir creates the data directory if it doesn't exist
func (c *Config) EnsureDataDir() error {
	dataDir := c.DataDirectory()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	return nil
}

// Load loads the config from the config file, creating a default one if it doesn't exist
func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create config directory
		configDir, err := ConfigDir()
		if err != nil {
			return nil, err
		}

		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		// Create default config
		config := DefaultConfig()
		if err := Save(&config); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

		return &config, nil
	}

	// Load existing config
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Load the theme
	if config.ThemeName == "" {
		config.ThemeName = "cozy-dark"
	}

	theme, err := LoadTheme(config.ThemeName)
	if err != nil {
		// Fall back to default theme if loading fails
		defaultTheme := CozyDark
		theme = &defaultTheme
	}
	config.ActiveTheme = theme

	return &config, nil
}

// Save saves the config to the config file
func Save(config *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
