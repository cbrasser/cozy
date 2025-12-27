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
	Tags     []string           // Folder names as tags (relative to library root)
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

// BookInfo holds basic information about a book for library display
type BookInfo struct {
	Path   string
	Title  string
	Author string
	Tags   []string
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
func ListBooks(dir string) ([]BookInfo, error) {
	var books []BookInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".epub" || ext == ".txt" {
			// Extract tags from folder path relative to library root
			tags := extractTags(path, dir)

			// Try to get book metadata
			bookInfo := BookInfo{
				Path: path,
				Tags: tags,
			}

			// Attempt to load title and author
			if book, err := Open(path); err == nil {
				bookInfo.Title = book.Title
				bookInfo.Author = book.Author
			}

			books = append(books, bookInfo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return books, nil
}

// extractTags extracts folder names as tags from the book path
func extractTags(bookPath, libraryRoot string) []string {
	// Get relative path from library root
	relPath, err := filepath.Rel(libraryRoot, bookPath)
	if err != nil {
		return []string{}
	}

	// Split path into components
	dir := filepath.Dir(relPath)

	// If file is directly in library root, no tags
	if dir == "." {
		return []string{}
	}

	// Split directory path into folder names
	parts := strings.Split(dir, string(filepath.Separator))

	// Filter out empty parts
	tags := []string{}
	for _, part := range parts {
		if part != "" && part != "." {
			tags = append(tags, part)
		}
	}

	return tags
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
