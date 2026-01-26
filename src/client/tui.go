package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dracula theme colors per AI.md lines 27377-27390
var (
	Background = lipgloss.Color("#282a36")
	Foreground = lipgloss.Color("#f8f8f2")
	Selection  = lipgloss.Color("#44475a")
	Comment    = lipgloss.Color("#6272a4")
	Cyan       = lipgloss.Color("#8be9fd")
	Green      = lipgloss.Color("#50fa7b")
	Orange     = lipgloss.Color("#ffb86c")
	Pink       = lipgloss.Color("#ff79c6")
	Purple     = lipgloss.Color("#bd93f9")
	Red        = lipgloss.Color("#ff5555")
	Yellow     = lipgloss.Color("#f1fa8c")
)

// TUI styles per AI.md line 27374
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(Comment).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true).
			Padding(1, 2)
)

// tuiModel represents the TUI state per AI.md line 26952
type tuiModel struct {
	config   *CLIConfig
	choices  []string
	cursor   int
	selected map[int]struct{}
	err      error
}

// Init initializes the model
func (m tuiModel) Init() tea.Cmd {
	return nil
}

// Update handles messages per AI.md line 27353
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Quit per AI.md line 27360
		case "q", "ctrl+c":
			return m, tea.Quit

		// Help per AI.md line 27359
		case "?":
			m.showHelp()
			return m, nil

		// Navigation per AI.md line 27357
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// Selection
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

// View renders the TUI per AI.md line 27361
func (m tuiModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Weather CLI - Interactive Mode"))
	b.WriteString("\n\n")

	// Error display per AI.md line 27362
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	// Menu items
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "✓"
		}

		style := lipgloss.NewStyle().Foreground(Foreground)
		if m.cursor == i {
			style = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
		}

		b.WriteString(style.Render(fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)))
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/k: up • ↓/j: down • enter: select • q: quit • ?: help"))

	return b.String()
}

// showHelp displays help screen per AI.md line 27359
func (m *tuiModel) showHelp() {
	// Help display would go here
}

// runTUI launches the interactive TUI mode per AI.md lines 26952-27390
func runTUI(config *CLIConfig) error {
	// Initialize TUI model
	m := tuiModel{
		config: config,
		choices: []string{
			"Current Weather",
			"Forecast",
			"Alerts",
			"Severe Weather",
			"Settings",
			"Exit",
		},
		selected: make(map[int]struct{}),
	}

	// Create program per AI.md line 27368
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

