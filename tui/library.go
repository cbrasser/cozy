package tui

import (
	"fmt"

	"github.com/cbrasser/cozy/config"
	"github.com/cbrasser/cozy/ebook"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LibraryModel represents the library view
type LibraryModel struct {
	config    *config.Config
	list      list.Model
	bookPaths []string
	width     int
	height    int
}

type bookItem struct {
	title string
	path  string
}

func (i bookItem) Title() string       { return i.title }
func (i bookItem) Description() string { return i.path }
func (i bookItem) FilterValue() string { return i.title }

// NewLibraryModel creates a new library model
func NewLibraryModel(cfg *config.Config) *LibraryModel {
	items := []list.Item{}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Your Library"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return &LibraryModel{
		config: cfg,
		list:   l,
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

		m.bookPaths = msg.Books
		items := make([]list.Item, len(msg.Books))
		for i, path := range msg.Books {
			// Try to load book metadata
			book, err := ebook.Open(path)
			title := path
			if err == nil && book.Title != "" {
				title = book.Title
				if book.Author != "" {
					title = fmt.Sprintf("%s - %s", book.Title, book.Author)
				}
			}

			items[i] = bookItem{
				title: title,
				path:  path,
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

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SecondaryColor)).
		Padding(1, 0)

	help := helpStyle.Render("↑/↓: navigate • enter: open book • /: search • q: quit")

	return titleStyle.Render("Cozy - E-Book Reader") + "\n" + m.list.View() + "\n" + help
}

// Messages
type BooksLoadedMsg struct {
	Books []string
	Error error
}

type BookLoadErrorMsg struct {
	Error error
}
