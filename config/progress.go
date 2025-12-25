package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// BookProgress tracks reading progress for a book
type BookProgress struct {
	BookPath       string `json:"book_path"`
	CurrentChapter int    `json:"current_chapter"`
	ScrollOffset   int    `json:"scroll_offset"` // Viewport Y offset within chapter
}

// ProgressData stores all reading progress
type ProgressData struct {
	Books map[string]BookProgress `json:"books"` // Key is book path
}

// LoadProgress loads reading progress from the data directory
func LoadProgress(cfg *Config) (*ProgressData, error) {
	if err := cfg.EnsureDataDir(); err != nil {
		return nil, err
	}

	progressPath := filepath.Join(cfg.DataDirectory(), "progress.json")

	// If file doesn't exist, return empty progress
	if _, err := os.Stat(progressPath); os.IsNotExist(err) {
		return &ProgressData{
			Books: make(map[string]BookProgress),
		}, nil
	}

	data, err := os.ReadFile(progressPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read progress file: %w", err)
	}

	var progress ProgressData
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("failed to parse progress file: %w", err)
	}

	if progress.Books == nil {
		progress.Books = make(map[string]BookProgress)
	}

	return &progress, nil
}

// SaveProgress saves reading progress to the data directory
func SaveProgress(cfg *Config, progress *ProgressData) error {
	if err := cfg.EnsureDataDir(); err != nil {
		return err
	}

	progressPath := filepath.Join(cfg.DataDirectory(), "progress.json")

	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	if err := os.WriteFile(progressPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write progress file: %w", err)
	}

	return nil
}

// GetBookProgress retrieves progress for a specific book
func (p *ProgressData) GetBookProgress(bookPath string) (BookProgress, bool) {
	progress, exists := p.Books[bookPath]
	return progress, exists
}

// SetBookProgress updates progress for a specific book
func (p *ProgressData) SetBookProgress(bookPath string, chapter, offset int) {
	p.Books[bookPath] = BookProgress{
		BookPath:       bookPath,
		CurrentChapter: chapter,
		ScrollOffset:   offset,
	}
}
