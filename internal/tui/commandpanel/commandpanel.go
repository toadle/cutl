package commandpanel

import (
	"cutl/internal/tui/styles"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width         int
	textInput     textinput.Model
	active        bool
	totalRows     int
	filteredRows  int
	currentLine   int
	filterActive  bool
	markedCount   int
	statusMessage string
	isStatusError bool
	isStatusNeutral bool
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

func (m *Model) SetMeta(totalRows, filteredRows, filteredPosition, markedCount int, filterActive bool) {
	m.totalRows = totalRows
	m.filteredRows = filteredRows
	m.currentLine = filteredPosition
	m.markedCount = markedCount
	m.filterActive = filterActive
}

func (m *Model) SetStatus(message string) {
	m.statusMessage = message
	m.isStatusError = false
	m.isStatusNeutral = false
}

func (m *Model) SetStatusError(message string) {
	m.statusMessage = message
	m.isStatusError = true
	m.isStatusNeutral = false
}

func (m *Model) SetStatusNeutral(message string) {
	m.statusMessage = message
	m.isStatusError = false
	m.isStatusNeutral = true
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
	metaContent := m.metaContent()
	selectionInfo := m.selectionInfo()

	if m.active {
		left := m.textInput.View()
		if selectionInfo != "" {
			left = lipgloss.JoinHorizontal(lipgloss.Top, left, selectionInfo)
		}
		return styles.CommandPanel.Width(m.width).Render(m.layoutWithMeta(m.appendStatus(left), metaContent))
	}

	var sections []string
	sections = append(sections, lipgloss.StyleRunes("D Details", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("F Filter", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("C Columns", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("E Edit", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.CommandLabelTrigger.Render("SPACE "),
		styles.CommandLabel.Render("Mark Single"),
	))
	sections = append(sections, lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.CommandLabelTrigger.Render("M "),
		styles.CommandLabel.Render("Filter Marked"),
	))
	sections = append(sections, lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.CommandLabelTrigger.Render("Ctrl+A "),
		styles.CommandLabel.Render("Select all"),
	))
	sections = append(sections, lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.CommandLabelTrigger.Render("X "),
		styles.CommandLabel.Render("Delete"),
	))
	sections = append(sections, lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.CommandLabelTrigger.Render("W "),
		styles.CommandLabel.Render("Write file"),
	))
	sections = append(sections, lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.CommandLabelTrigger.Render("1-9 "),
		styles.CommandLabel.Render("Sort by column"),
	))
	if selectionInfo != "" {
		sections = append(sections, selectionInfo)
	}

	left := lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	return styles.CommandPanel.Width(m.width).Render(m.layoutWithMeta(m.appendStatus(left), metaContent))
}

func (m *Model) metaContent() string {
	current := "–"
	total := m.filteredRows
	if total == 0 {
		total = m.totalRows
	}
	
	if m.currentLine > 0 {
		current = fmt.Sprintf("%d", m.currentLine)
	}

	base := fmt.Sprintf("%s / %d (%d total)", current, total, m.totalRows)

	return base
}

func (m *Model) appendStatus(content string) string {
	if m.statusMessage == "" {
		return content
	}

	var status string
	if m.isStatusError {
		status = styles.CommandStatusError.Render(m.statusMessage)
	} else if m.isStatusNeutral {
		status = styles.CommandStatusNeutral.Render(m.statusMessage)
	} else {
		status = styles.CommandStatus.Render(m.statusMessage)
	}
	return lipgloss.JoinVertical(lipgloss.Left, content, status)
}

func (m *Model) selectionInfo() string {
	if m.markedCount == 0 {
		return ""
	}

	hint := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.CommandSelectionLabelTrigger.Render("ESC "),
		styles.CommandSelectionLabel.Render(fmt.Sprintf("Clear selection of %d lines", m.markedCount)),
	)

	return hint
}

func (m *Model) layoutWithMeta(left, meta string) string {
	leftStyle := lipgloss.NewStyle().Align(lipgloss.Left)
	rightStyle := lipgloss.NewStyle().Align(lipgloss.Right)

	if m.width > 0 {
		// Begrenzen der rechten Anzeige auf maximal 50 Zeichen
		maxRightWidth := 50
		rightWidth := maxRightWidth
		if m.width-8 < maxRightWidth {
			rightWidth = m.width - 8
		}
		leftWidth := (m.width - 8) - rightWidth
		leftStyle = leftStyle.Width(leftWidth)
		rightStyle = rightStyle.Width(rightWidth)
	}

	renderedLeft := leftStyle.Render(left)
	renderedRight := ""
	if meta != "" {
		renderedRight = rightStyle.Render(styles.CommandMeta.Render(meta))
	} else if m.width > 0 {
		renderedRight = rightStyle.Render("")
	}

	if renderedRight != "" || m.width > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Top, renderedLeft, renderedRight)
	}

	return renderedLeft
}
