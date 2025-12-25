package ebook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Chapter represents a book chapter
type Chapter struct {
	Title   string
	Content string // Full chapter content
	Order   int    // Position in book
}

// Book represents an e-book
type Book struct {
	Path     string
	Title    string
	Author   string
	Format   Format
	Chapters []Chapter          // Book chapters
	Metadata map[string]string
}

// Format represents the e-book format
type Format string

const (
	FormatEPUB Format = "epub"
	FormatText Format = "txt"
)

// Reader interface for different e-book formats
type Reader interface {
	Read(path string) (*Book, error)
}

// Open opens an e-book file and returns a Book
func Open(path string) (*Book, error) {
	ext := strings.ToLower(filepath.Ext(path))

	var reader Reader
	var format Format

	switch ext {
	case ".epub":
		reader = &EPUBReader{}
		format = FormatEPUB
	case ".txt":
		reader = &TextReader{}
		format = FormatText
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	book, err := reader.Read(path)
	if err != nil {
		return nil, err
	}

	book.Format = format
	book.Path = path

	return book, nil
}

// ListBooks lists all supported e-books in a directory
func ListBooks(dir string) ([]string, error) {
	var books []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".epub" || ext == ".txt" {
			books = append(books, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return books, nil
}

// GetChapter returns a specific chapter
func (b *Book) GetChapter(index int) *Chapter {
	if index < 0 || index >= len(b.Chapters) {
		return nil
	}
	return &b.Chapters[index]
}

// ChapterCount returns the number of chapters
func (b *Book) ChapterCount() int {
	return len(b.Chapters)
}
