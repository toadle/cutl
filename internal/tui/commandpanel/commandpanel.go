package commandpanel

import (
	"cutl/internal/tui/styles"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width     int
	textInput textinput.Model
	active    bool
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "id, title, ..."
	ti.CharLimit = 156
	ti.Width = 50

	m := Model{
		textInput: ti,
		active:    false,
	}
	return m
}

func (m *Model) Activate(queries []string) {
	m.active = true
	m.textInput.SetValue(strings.Join(queries, ", "))
	m.textInput.Focus()
}

func (m *Model) Deactivate() {
	m.active = false
	m.textInput.Blur()
}

func (m *Model) Value() string {
	return m.textInput.Value()
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	if m.active {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return *m, cmd
}

func (m *Model) View() string {
	if m.active {
		return styles.CommandPanel.Width(m.width).Render(m.textInput.View())
	}

	var sections []string

	sections = append(sections, styles.CommandTitle.Render("Commands"))
	sections = append(sections, lipgloss.StyleRunes("Details", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Filter", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Columns", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Edit", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))

	return styles.CommandPanel.Width(m.width).Render(lipgloss.JoinHorizontal(lipgloss.Top, sections...))
}
