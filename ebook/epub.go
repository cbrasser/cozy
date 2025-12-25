package ebook

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// EPUB metadata structures
type container struct {
	XMLName   xml.Name   `xml:"container"`
	Rootfiles []rootfile `xml:"rootfiles>rootfile"`
}

type rootfile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

type opfPackage struct {
	XMLName  xml.Name     `xml:"package"`
	Metadata opfMetadata  `xml:"metadata"`
	Manifest opfManifest  `xml:"manifest"`
	Spine    opfSpine     `xml:"spine"`
}

type opfMetadata struct {
	Title   []string `xml:"title"`
	Creator []string `xml:"creator"`
	Lang    string   `xml:"language"`
}

type opfManifest struct {
	Items []opfItem `xml:"item"`
}

type opfItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type opfSpine struct {
	Itemrefs []opfItemref `xml:"itemref"`
}

type opfItemref struct {
	IDref string `xml:"idref,attr"`
}

// EPUBReader reads EPUB files
type EPUBReader struct{}

// Read reads an EPUB file
func (r *EPUBReader) Read(path string) (*Book, error) {
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer zipReader.Close()

	book := &Book{
		Metadata: make(map[string]string),
		Title:    filepath.Base(path),
	}

	// Step 1: Read container.xml to find the OPF file
	opfPath, err := findOPFPath(zipReader)
	if err != nil {
		// Fallback: read all HTML files if we can't find OPF
		return r.readFallback(zipReader, book)
	}

	// Step 2: Parse the OPF file
	opf, err := parseOPF(zipReader, opfPath)
	if err != nil {
		return r.readFallback(zipReader, book)
	}

	// Step 3: Extract metadata
	if len(opf.Metadata.Title) > 0 {
		book.Title = opf.Metadata.Title[0]
	}
	if len(opf.Metadata.Creator) > 0 {
		book.Author = opf.Metadata.Creator[0]
	}
	if opf.Metadata.Lang != "" {
		book.Metadata["language"] = opf.Metadata.Lang
	}

	// Step 4: Build manifest map
	manifestMap := make(map[string]opfItem)
	for _, item := range opf.Manifest.Items {
		manifestMap[item.ID] = item
	}

	// Step 5: Read chapters in spine order
	opfDir := filepath.Dir(opfPath)
	for i, itemref := range opf.Spine.Itemrefs {
		if item, ok := manifestMap[itemref.IDref]; ok {
			// Construct the full path relative to OPF
			contentPath := filepath.Join(opfDir, item.Href)
			contentPath = filepath.Clean(contentPath)

			// Read the chapter content
			content, err := readFileFromZip(zipReader, contentPath)
			if err != nil {
				continue
			}

			htmlContent := string(content)

			// Extract chapter title from the HTML or use a default
			chapterTitle := extractTitle(htmlContent)
			if chapterTitle == "" {
				chapterTitle = fmt.Sprintf("Chapter %d", i+1)
			}

			// Store the raw HTML - we'll render it with theme later
			if strings.TrimSpace(htmlContent) != "" {
				book.Chapters = append(book.Chapters, Chapter{
					Title:   chapterTitle,
					Content: htmlContent, // Store raw HTML
					Order:   i,
				})
			}
		}
	}

	if len(book.Chapters) == 0 {
		return nil, fmt.Errorf("no chapters found in EPUB")
	}

	return book, nil
}

// readFallback reads all HTML files when OPF parsing fails
func (r *EPUBReader) readFallback(zipReader *zip.ReadCloser, book *Book) (*Book, error) {
	type fileWithContent struct {
		name    string
		content string
	}

	var files []fileWithContent

	for _, f := range zipReader.File {
		name := strings.ToLower(f.Name)
		if strings.HasSuffix(name, ".html") ||
		   strings.HasSuffix(name, ".xhtml") ||
		   strings.HasSuffix(name, ".htm") {
			fileRC, err := f.Open()
			if err != nil {
				continue
			}

			data, err := io.ReadAll(fileRC)
			fileRC.Close()
			if err != nil {
				continue
			}

			htmlContent := string(data)
			if strings.TrimSpace(htmlContent) != "" && len(htmlContent) > 100 {
				files = append(files, fileWithContent{
					name:    f.Name,
					content: htmlContent, // Store raw HTML
				})
			}
		}
	}

	// Sort files alphabetically
	sort.Slice(files, func(i, j int) bool {
		return files[i].name < files[j].name
	})

	for i, f := range files {
		book.Chapters = append(book.Chapters, Chapter{
			Title:   filepath.Base(f.name),
			Content: f.content,
			Order:   i,
		})
	}

	if book.Title == "" {
		book.Title = strings.TrimSuffix(filepath.Base(book.Path), filepath.Ext(book.Path))
	}

	return book, nil
}

// findOPFPath reads container.xml to find the OPF file path
func findOPFPath(zipReader *zip.ReadCloser) (string, error) {
	data, err := readFileFromZip(zipReader, "META-INF/container.xml")
	if err != nil {
		return "", err
	}

	var cont container
	if err := xml.Unmarshal(data, &cont); err != nil {
		return "", err
	}

	if len(cont.Rootfiles) == 0 {
		return "", fmt.Errorf("no rootfiles found in container.xml")
	}

	return cont.Rootfiles[0].FullPath, nil
}

// parseOPF parses the OPF (Open Packaging Format) file
func parseOPF(zipReader *zip.ReadCloser, opfPath string) (*opfPackage, error) {
	data, err := readFileFromZip(zipReader, opfPath)
	if err != nil {
		return nil, err
	}

	var opf opfPackage
	if err := xml.Unmarshal(data, &opf); err != nil {
		return nil, err
	}

	return &opf, nil
}

// readFileFromZip reads a file from the ZIP archive
func readFileFromZip(zipReader *zip.ReadCloser, path string) ([]byte, error) {
	path = filepath.Clean(path)
	for _, f := range zipReader.File {
		if filepath.Clean(f.Name) == path {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// extractTitle extracts the title from HTML content
func extractTitle(html string) string {
	// Look for <title> or <h1>
	titleStart := strings.Index(html, "<title>")
	if titleStart != -1 {
		titleEnd := strings.Index(html[titleStart:], "</title>")
		if titleEnd != -1 {
			return strings.TrimSpace(html[titleStart+7 : titleStart+titleEnd])
		}
	}

	h1Start := strings.Index(html, "<h1")
	if h1Start != -1 {
		contentStart := strings.Index(html[h1Start:], ">")
		if contentStart != -1 {
			h1End := strings.Index(html[h1Start:], "</h1>")
			if h1End != -1 {
				return stripHTMLTags(html[h1Start+contentStart+1 : h1Start+h1End])
			}
		}
	}

	return ""
}

// htmlToText converts HTML to plain text with some formatting preserved
func htmlToText(html string) string {
	result := html

	// Add line breaks for block elements
	blockElements := []string{"</p>", "</div>", "</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>", "<br>", "<br/>", "</li>"}
	for _, elem := range blockElements {
		result = strings.ReplaceAll(result, elem, elem+"\n")
	}

	// Strip all HTML tags
	result = stripHTMLTags(result)

	// Clean up excessive whitespace
	lines := strings.Split(result, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	return strings.Join(cleanedLines, "\n\n")
}

// stripHTMLTags performs basic HTML tag removal
func stripHTMLTags(html string) string {
	inTag := false
	var result strings.Builder

	for _, char := range html {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	return result.String()
}
