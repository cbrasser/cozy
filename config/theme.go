package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Theme represents a complete color and style theme
type Theme struct {
	Name string `toml:"name"`

	// UI Colors
	PrimaryColor     string `toml:"primary_color"`
	SecondaryColor   string `toml:"secondary_color"`
	BackgroundColor  string `toml:"background_color"`

	// Text Colors
	TextColor        string `toml:"text_color"`
	MutedTextColor   string `toml:"muted_text_color"`

	// Element Colors
	HeadingColor     string `toml:"heading_color"`
	LinkColor        string `toml:"link_color"`
	QuoteColor       string `toml:"quote_color"`
	QuoteBorderColor string `toml:"quote_border_color"`
	CodeBgColor      string `toml:"code_bg_color"`
	CodeTextColor    string `toml:"code_text_color"`
	EmphasisColor    string `toml:"emphasis_color"`
	StrongColor      string `toml:"strong_color"`
}

// Built-in themes
var (
	// CozyDark - A warm, purple-tinted dark theme (default)
	CozyDark = Theme{
		Name:             "cozy-dark",
		PrimaryColor:     "#A78BFA",   // Soft purple
		SecondaryColor:   "#C4B5FD",   // Lighter purple
		BackgroundColor:  "#1F2937",   // Dark blue-gray
		TextColor:        "#F3F4F6",   // Off-white
		MutedTextColor:   "#9CA3AF",   // Gray
		HeadingColor:     "#DDD6FE",   // Light purple
		LinkColor:        "#60A5FA",   // Blue
		QuoteColor:       "#D1D5DB",   // Light gray
		QuoteBorderColor: "#7C3AED",   // Purple
		CodeBgColor:      "#374151",   // Darker gray
		CodeTextColor:    "#FCD34D",   // Yellow
		EmphasisColor:    "#FBBF24",   // Amber
		StrongColor:      "#F9A8D4",   // Pink
	}

	// SolarizedDark - Classic Solarized dark theme
	SolarizedDark = Theme{
		Name:             "solarized-dark",
		PrimaryColor:     "#268BD2",   // Blue
		SecondaryColor:   "#2AA198",   // Cyan
		BackgroundColor:  "#002B36",   // Base03
		TextColor:        "#839496",   // Base0
		MutedTextColor:   "#586E75",   // Base01
		HeadingColor:     "#B58900",   // Yellow
		LinkColor:        "#268BD2",   // Blue
		QuoteColor:       "#93A1A1",   // Base1
		QuoteBorderColor: "#2AA198",   // Cyan
		CodeBgColor:      "#073642",   // Base02
		CodeTextColor:    "#859900",   // Green
		EmphasisColor:    "#CB4B16",   // Orange
		StrongColor:      "#DC322F",   // Red
	}

	// Sepia - Warm, book-like theme
	Sepia = Theme{
		Name:             "sepia",
		PrimaryColor:     "#8B4513",   // Saddle brown
		SecondaryColor:   "#A0522D",   // Sienna
		BackgroundColor:  "#F5E6D3",   // Sepia background
		TextColor:        "#3E2723",   // Dark brown
		MutedTextColor:   "#6D4C41",   // Medium brown
		HeadingColor:     "#5D4037",   // Dark brown
		LinkColor:        "#D2691E",   // Chocolate
		QuoteColor:       "#4E342E",   // Dark brown
		QuoteBorderColor: "#8D6E63",   // Brown
		CodeBgColor:      "#EFEBE9",   // Light brown
		CodeTextColor:    "#33691E",   // Dark green
		EmphasisColor:    "#BF360C",   // Deep orange
		StrongColor:      "#6D4C41",   // Medium brown
	}
)

// BuiltInThemes returns all built-in themes
func BuiltInThemes() map[string]Theme {
	return map[string]Theme{
		"cozy-dark":      CozyDark,
		"solarized-dark": SolarizedDark,
		"sepia":          Sepia,
	}
}

// LoadTheme loads a theme by name (built-in or from file)
func LoadTheme(name string) (*Theme, error) {
	// Check built-in themes first
	if theme, ok := BuiltInThemes()[name]; ok {
		return &theme, nil
	}

	// Try to load from theme file
	themePath, err := ThemePath(name)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("theme not found: %s", name)
	}

	var theme Theme
	if _, err := toml.DecodeFile(themePath, &theme); err != nil {
		return nil, fmt.Errorf("failed to load theme file: %w", err)
	}

	return &theme, nil
}

// ThemePath returns the path to a custom theme file
func ThemePath(name string) (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	themesDir := filepath.Join(configDir, "themes")
	return filepath.Join(themesDir, name+".toml"), nil
}

// SaveTheme saves a theme to a file
func SaveTheme(theme *Theme) error {
	themePath, err := ThemePath(theme.Name)
	if err != nil {
		return err
	}

	// Create themes directory if it doesn't exist
	themesDir := filepath.Dir(themePath)
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	file, err := os.Create(themePath)
	if err != nil {
		return fmt.Errorf("failed to create theme file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(theme); err != nil {
		return fmt.Errorf("failed to encode theme: %w", err)
	}

	return nil
}

// ListThemes returns a list of all available theme names
func ListThemes() ([]string, error) {
	themes := []string{}

	// Add built-in themes
	for name := range BuiltInThemes() {
		themes = append(themes, name)
	}

	// Add custom themes from themes directory
	configDir, err := ConfigDir()
	if err != nil {
		return themes, nil // Return built-ins only if config dir fails
	}

	themesDir := filepath.Join(configDir, "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return themes, nil // No custom themes directory
	}

	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return themes, nil // Return built-ins only if read fails
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".toml" {
			name := entry.Name()[:len(entry.Name())-5] // Remove .toml extension
			themes = append(themes, name)
		}
	}

	return themes, nil
}
