package tui

import (
	"fmt"
	"strings"

	"github.com/cbrasser/cozy/config"
	"github.com/cbrasser/cozy/ebook"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LibraryModel represents the library view
type LibraryModel struct {
	config   *config.Config
	list     list.Model
	books    []ebook.BookInfo
	progress *config.ProgressData
	width    int
	height   int
}

type bookItem struct {
	title      string
	author     string
	path       string
	tags       []string
	completion float64
	finished   bool
}

func (i bookItem) Title() string { return i.title }
func (i bookItem) Description() string {
	parts := []string{}

	if len(i.tags) > 0 {
		parts = append(parts, "📁 "+strings.Join(i.tags, " / "))
	}

	if i.author != "" {
		parts = append(parts, i.author)
	}

	// Add completion percentage or finished status
	if i.finished {
		parts = append(parts, "✓ Finished")
	} else if i.completion > 0 {
		parts = append(parts, fmt.Sprintf("%.0f%%", i.completion))
	}

	return strings.Join(parts, " • ")
}
func (i bookItem) FilterValue() string {
	// Allow filtering by title, author, and tags
	filterValue := i.title
	if i.author != "" {
		filterValue += " " + i.author
	}
	if len(i.tags) > 0 {
		filterValue += " " + strings.Join(i.tags, " ")
	}
	return filterValue
}

// NewLibraryModel creates a new library model
func NewLibraryModel(cfg *config.Config) *LibraryModel {
	items := []list.Item{}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Your Library"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	// Add custom help text for the 'f' key
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("f"),
				key.WithHelp("f", "toggle finished"),
			),
		}
	}

	// Load progress data
	progress, err := config.LoadProgress(cfg)
	if err != nil {
		// If loading fails, create empty progress
		progress = &config.ProgressData{
			Books: make(map[string]config.BookProgress),
		}
	}

	return &LibraryModel{
		config:   cfg,
		list:     l,
		progress: progress,
	}
}

// Init initializes the library view
func (m *LibraryModel) Init() tea.Cmd {
	return m.loadBooks()
}

// loadBooks loads books from the library path
func (m *LibraryModel) loadBooks() tea.Cmd {
	return func() tea.Msg {
		bookPaths, err := ebook.ListBooks(m.config.Library.Path)
		if err != nil {
			return BooksLoadedMsg{Error: err}
		}
		return BooksLoadedMsg{Books: bookPaths}
	}
}

// SetSize updates the size of the library view
func (m *LibraryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-4)
}

// Update handles messages for the library view
func (m *LibraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BooksLoadedMsg:
		if msg.Error != nil {
			return m, nil
		}

		m.books = msg.Books
		items := make([]list.Item, len(msg.Books))
		for i, bookInfo := range msg.Books {
			title := bookInfo.Path
			author := ""
			if bookInfo.Title != "" {
				title = bookInfo.Title
			}
			if bookInfo.Author != "" {
				author = bookInfo.Author
			}

			// Get progress data for this book
			completion := 0.0
			finished := false
			if bookProgress, exists := m.progress.GetBookProgress(bookInfo.Path); exists {
				completion = bookProgress.GetCompletionPercentage()
				finished = bookProgress.Finished
			}

			items[i] = bookItem{
				title:      title,
				author:     author,
				path:       bookInfo.Path,
				tags:       bookInfo.Tags,
				completion: completion,
				finished:   finished,
			}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Load the selected book
			if i, ok := m.list.SelectedItem().(bookItem); ok {
				return m, m.openBook(i.path)
			}
		case "f":
			// Toggle finished status for the selected book
			if i, ok := m.list.SelectedItem().(bookItem); ok {
				m.progress.SetBookFinished(i.path, !i.finished)
				config.SaveProgress(m.config, m.progress)
				// Reload the list to reflect changes
				return m, m.loadBooks()
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// openBook opens a book and sends a BookSelectedMsg
func (m *LibraryModel) openBook(path string) tea.Cmd {
	return func() tea.Msg {
		book, err := ebook.Open(path)
		if err != nil {
			return BookLoadErrorMsg{Error: err}
		}
		return BookSelectedMsg{Book: book}
	}
}

// View renders the library view
func (m *LibraryModel) View() string {
	if m.config.ActiveTheme == nil {
		return "Loading..."
	}

	theme := m.config.ActiveTheme

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Padding(1, 0)

	return titleStyle.Render("Cozy - E-Book Reader") + "\n" + m.list.View()
}

// Messages
type BooksLoadedMsg struct {
	Books []ebook.BookInfo
	Error error
}

type BookLoadErrorMsg struct {
	Error error
}
