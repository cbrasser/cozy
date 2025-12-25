package ebook

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TextReader reads plain text files
type TextReader struct{}

const charsPerPage = 2000 // Approximate characters per page

// Read reads a plain text file
func (r *TextReader) Read(path string) (*Book, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open text file: %w", err)
	}
	defer file.Close()

	book := &Book{
		Title:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Metadata: make(map[string]string),
	}

	// Read entire file
	var fullText strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fullText.WriteString(scanner.Text())
		fullText.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	// For plain text, treat the entire file as one chapter
	content := fullText.String()
	book.Chapters = []Chapter{
		{
			Title:   book.Title,
			Content: content,
			Order:   0,
		},
	}

	return book, nil
}

// splitIntoPages splits text into pages of approximately equal size
func splitIntoPages(text string, charsPerPage int) []string {
	if len(text) == 0 {
		return []string{""}
	}

	var pages []string
	lines := strings.Split(text, "\n")

	var currentPage strings.Builder
	currentLength := 0

	for _, line := range lines {
		lineLength := len(line) + 1 // +1 for newline

		// If adding this line would exceed the page limit, start a new page
		if currentLength > 0 && currentLength+lineLength > charsPerPage {
			pages = append(pages, currentPage.String())
			currentPage.Reset()
			currentLength = 0
		}

		currentPage.WriteString(line)
		currentPage.WriteString("\n")
		currentLength += lineLength
	}

	// Add the last page if it has content
	if currentPage.Len() > 0 {
		pages = append(pages, currentPage.String())
	}

	return pages
}
