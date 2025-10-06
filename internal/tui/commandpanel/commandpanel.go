package commandpanel

import (
	"cutl/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func New() Model {
	m := Model{}
	return m
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return *m, cmd
}

func (m *Model) View() string {
	var sections []string

	sections = append(sections, styles.CommandTitle.Render("Commands"))
	sections = append(sections, lipgloss.StyleRunes("Details", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Filter", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Columns", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Edit", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))

	return styles.CommandPanel.Width(m.width).Render(lipgloss.JoinHorizontal(lipgloss.Top, sections...))
}
