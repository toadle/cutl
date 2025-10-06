package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"cutl/internal/editor"
	"cutl/internal/messages"
	"cutl/internal/tui/cutable"
	"cutl/internal/tui/styles"
)

type Model struct {
	width  int
	height int

	jsonlPath string
	table cutable.Model
}

func New(jsonlPath string) *Model {
	m := &Model{
		table:   cutable.New(),
		jsonlPath: jsonlPath,
	}

	return m
}

func (m *Model) Init() tea.Cmd {	
	return func() tea.Msg {
		jsonlContent, err := editor.LoadJSONL(m.jsonlPath)

		if err != nil {
			log.Errorf("Failed to load JSONL file %s: %v", m.jsonlPath, err)
			return messages.InputFileLoadError{
				Error: err,
			}
		} else {
			log.Infof("JSONL file %s loaded successfully.", m.jsonlPath)
			return messages.InputFileLoaded{
				Content: jsonlContent,
			}
		}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	m.table, msg = m.table.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var (
		sections []string
	)
	sections = append(sections, m.table.View())

	return styles.App.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}
