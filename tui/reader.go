package tui

import (
	"fmt"
	"strings"

	"github.com/cbrasser/cozy/config"
	"github.com/cbrasser/cozy/ebook"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// readerKeyMap defines key bindings for the reader
type readerKeyMap struct {
	NextChapter     key.Binding
	PrevChapter     key.Binding
	NextHeading     key.Binding
	PrevHeading     key.Binding
	FirstChapter    key.Binding
	LastChapter     key.Binding
	ScrollUp        key.Binding
	ScrollDown      key.Binding
	HalfPageUp      key.Binding
	HalfPageDown    key.Binding
	Back            key.Binding
	Quit            key.Binding
	ToggleHelp      key.Binding
}

func (k readerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextChapter, k.PrevChapter, k.NextHeading, k.PrevHeading, k.ScrollUp, k.ScrollDown, k.Back, k.Quit}
}

func (k readerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextChapter, k.PrevChapter, k.NextHeading, k.PrevHeading, k.FirstChapter, k.LastChapter},
		{k.ScrollUp, k.ScrollDown, k.HalfPageUp, k.HalfPageDown, k.Back, k.Quit},
		{k.ToggleHelp},
	}
}

var readerKeys = readerKeyMap{
	NextChapter: key.NewBinding(
		key.WithKeys("l", "right", "n", "pgdown"),
		key.WithHelp("l/→/n", "next chapter"),
	),
	PrevChapter: key.NewBinding(
		key.WithKeys("h", "left", "p", "pgup"),
		key.WithHelp("h/←/p", "previous chapter"),
	),
	NextHeading: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "next section"),
	),
	PrevHeading: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "previous section"),
	),
	FirstChapter: key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "first chapter"),
	),
	LastChapter: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "last chapter"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("down", "j", " "),
		key.WithHelp("↓/j/space", "scroll down"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "half page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "half page down"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back to library"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	ToggleHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

// ReaderModel represents the book reader view
type ReaderModel struct {
	config           *config.Config
	book             *ebook.Book
	viewport         viewport.Model
	help             help.Model
	keys             readerKeyMap
	currentChapter   int
	headingPositions []int // Line numbers of H2/H3 headings in current chapter
	progress         *config.ProgressData
	width            int
	height           int
}

// NewReaderModel creates a new reader model
func NewReaderModel(cfg *config.Config) *ReaderModel {
	vp := viewport.New(0, 0)
	h := help.New()

	// Load reading progress
	progress, err := config.LoadProgress(cfg)
	if err != nil {
		// If loading fails, create empty progress
		progress = &config.ProgressData{
			Books: make(map[string]config.BookProgress),
		}
	}

	return &ReaderModel{
		config:   cfg,
		viewport: vp,
		help:     h,
		keys:     readerKeys,
		progress: progress,
	}
}

// Init initializes the reader model
func (m *ReaderModel) Init() tea.Cmd {
	return nil
}

// SaveProgress saves the current reading position
func (m *ReaderModel) SaveProgress() {
	if m.book != nil {
		m.progress.SetBookProgress(m.book.Path, m.currentChapter, m.viewport.YOffset)
		config.SaveProgress(m.config, m.progress)
	}
}

// LoadBook loads a book into the reader
func (m *ReaderModel) LoadBook(book *ebook.Book) {
	m.book = book

	// Try to restore saved progress for this book
	if savedProgress, exists := m.progress.GetBookProgress(book.Path); exists {
		m.currentChapter = savedProgress.CurrentChapter
		// Ensure chapter is valid
		if m.currentChapter >= book.ChapterCount() {
			m.currentChapter = 0
		}
		m.updateViewport()
		// Restore scroll position
		m.viewport.SetYOffset(savedProgress.ScrollOffset)
	} else {
		// No saved progress, start from beginning
		m.currentChapter = 0
		m.updateViewport()
	}
}

// SetSize updates the size of the reader view
func (m *ReaderModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.help.Width = width
	m.viewport.Width = width - m.config.Display.MarginLeft - m.config.Display.MarginRight
	m.viewport.Height = height - 6 // Account for header and footer
	m.updateViewport()
}

// updateViewport updates the viewport with the current chapter content
func (m *ReaderModel) updateViewport() {
	if m.book == nil || m.config.ActiveTheme == nil {
		return
	}

	chapter := m.book.GetChapter(m.currentChapter)
	if chapter == nil {
		return
	}

	// Use viewport width for rendering
	renderWidth := m.viewport.Width
	if renderWidth <= 0 {
		renderWidth = 80 // Default width
	}

	// Render HTML to styled text based on book format
	var renderedContent string
	if m.book.Format == ebook.FormatEPUB {
		// EPUB: render HTML with rich formatting and track heading positions
		renderResult := ebook.RenderToStyledTextWithHeadings(chapter.Content, m.config.ActiveTheme, renderWidth)
		renderedContent = renderResult.Text
		m.headingPositions = renderResult.HeadingPositions
	} else {
		// Plain text: just wrap it
		renderedContent = wordwrap.String(chapter.Content, renderWidth)
		m.headingPositions = []int{}
	}

	m.viewport.SetContent(renderedContent)
	m.viewport.GotoTop()
}

// Update handles messages for the reader view
func (m *ReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.book == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Back):
			// Save reading progress
			m.SaveProgress()
			return m, func() tea.Msg { return BackToLibraryMsg{} }

		case key.Matches(msg, m.keys.NextChapter):
			// Next chapter
			if m.currentChapter < m.book.ChapterCount()-1 {
				m.currentChapter++
				m.updateViewport()
			}
			return m, nil

		case key.Matches(msg, m.keys.NextHeading):
			// Jump to next heading (H2/H3) within the current chapter
			currentLine := m.viewport.YOffset

			// Find the next heading after the current position
			nextHeadingLine := -1
			for _, headingLine := range m.headingPositions {
				if headingLine > currentLine {
					nextHeadingLine = headingLine
					break
				}
			}

			if nextHeadingLine >= 0 {
				// Jump to the heading within the current chapter
				m.viewport.SetYOffset(nextHeadingLine)
			} else {
				// No more headings in this chapter, go to next chapter
				if m.currentChapter < m.book.ChapterCount()-1 {
					m.currentChapter++
					m.updateViewport()
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.PrevHeading):
			// Jump to previous heading (H2/H3) within the current chapter
			currentLine := m.viewport.YOffset

			// Find the previous heading before the current position
			prevHeadingLine := -1
			for i := len(m.headingPositions) - 1; i >= 0; i-- {
				headingLine := m.headingPositions[i]
				if headingLine < currentLine {
					prevHeadingLine = headingLine
					break
				}
			}

			if prevHeadingLine >= 0 {
				// Jump to the heading within the current chapter
				m.viewport.SetYOffset(prevHeadingLine)
			} else {
				// No more headings before this in the chapter, go to previous chapter
				if m.currentChapter > 0 {
					m.currentChapter--
					m.updateViewport()
					// Go to the last heading in the previous chapter
					if len(m.headingPositions) > 0 {
						m.viewport.SetYOffset(m.headingPositions[len(m.headingPositions)-1])
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.HalfPageDown):
			// Scroll down half a viewport
			m.viewport.HalfViewDown()
			return m, nil

		case key.Matches(msg, m.keys.HalfPageUp):
			// Scroll up half a viewport
			m.viewport.HalfViewUp()
			return m, nil

		case key.Matches(msg, m.keys.PrevChapter):
			// Previous chapter
			if m.currentChapter > 0 {
				m.currentChapter--
				m.updateViewport()
			}
			return m, nil

		case key.Matches(msg, m.keys.FirstChapter):
			// First chapter
			m.currentChapter = 0
			m.updateViewport()
			return m, nil

		case key.Matches(msg, m.keys.LastChapter):
			// Last chapter
			m.currentChapter = m.book.ChapterCount() - 1
			m.updateViewport()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the reader view
func (m *ReaderModel) View() string {
	if m.book == nil || m.config.ActiveTheme == nil {
		return "No book loaded"
	}

	theme := m.config.ActiveTheme

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Padding(0, 1)

	chapterTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SecondaryColor)).
		Italic(true).
		Padding(0, 1)

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SecondaryColor)).
		Padding(0, 1)

	// Header with book title
	title := m.book.Title
	if m.book.Author != "" {
		title = fmt.Sprintf("%s - %s", m.book.Title, m.book.Author)
	}
	header := headerStyle.Render(title)

	// Chapter title
	chapter := m.book.GetChapter(m.currentChapter)
	chapterTitle := ""
	if chapter != nil {
		chapterTitle = chapterTitleStyle.Render(fmt.Sprintf("Chapter %d/%d: %s",
			m.currentChapter+1,
			m.book.ChapterCount(),
			chapter.Title))
	}

	// Progress indicator
	progress := fmt.Sprintf("Chapter %d/%d • Scroll: %.0f%%",
		m.currentChapter+1,
		m.book.ChapterCount(),
		m.viewport.ScrollPercent()*100,
	)

	// Help view
	helpView := m.help.View(m.keys)

	// Combine header, viewport, and footer
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		chapterTitle,
		strings.Repeat("─", m.width),
		m.viewport.View(),
		strings.Repeat("─", m.width),
		progressStyle.Render(progress),
		helpView,
	)
}
