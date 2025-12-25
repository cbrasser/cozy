package tui

import (
	"github.com/cbrasser/cozy/config"
	"github.com/cbrasser/cozy/ebook"
	tea "github.com/charmbracelet/bubbletea"
)

// View represents different screens in the TUI
type View int

const (
	ViewLibrary View = iota
	ViewReader
)

// Model is the main Bubbletea model
type Model struct {
	config       *config.Config
	currentView  View
	library      *LibraryModel
	reader       *ReaderModel
	width        int
	height       int
	err          error
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config) Model {
	return Model{
		config:      cfg,
		currentView: ViewLibrary,
		library:     NewLibraryModel(cfg),
		reader:      NewReaderModel(cfg),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.library.Init()
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.library.SetSize(msg.Width, msg.Height)
		m.reader.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Save reading progress before quitting
			if m.currentView == ViewReader {
				m.reader.SaveProgress()
			}
			return m, tea.Quit
		}

	case BookSelectedMsg:
		// Switch to reader view when a book is selected
		m.currentView = ViewReader
		m.reader.LoadBook(msg.Book)
		return m, nil

	case BackToLibraryMsg:
		// Return to library view
		m.currentView = ViewLibrary
		return m, nil
	}

	// Route updates to the current view
	var cmd tea.Cmd
	switch m.currentView {
	case ViewLibrary:
		libModel, libCmd := m.library.Update(msg)
		m.library = libModel.(*LibraryModel)
		cmd = libCmd
	case ViewReader:
		readerModel, readerCmd := m.reader.Update(msg)
		m.reader = readerModel.(*ReaderModel)
		cmd = readerCmd
	}

	return m, cmd
}

// View renders the current view
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit."
	}

	switch m.currentView {
	case ViewLibrary:
		return m.library.View()
	case ViewReader:
		return m.reader.View()
	default:
		return "Unknown view"
	}
}

// Messages for inter-view communication
type BookSelectedMsg struct {
	Book *ebook.Book
}

type BackToLibraryMsg struct{}
