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
	width        int
	textInput    textinput.Model
	active       bool
	totalRows    int
	filteredRows int
	currentLine  int
	filterActive bool
	markedCount  int
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

func (m *Model) SetMeta(totalRows, filteredRows, currentLine, markedCount int, filterActive bool) {
	m.totalRows = totalRows
	m.filteredRows = filteredRows
	m.currentLine = currentLine
	m.markedCount = markedCount
	m.filterActive = filterActive
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
		return styles.CommandPanel.Width(m.width).Render(m.layoutWithMeta(left, metaContent))
	}

	var sections []string
	sections = append(sections, styles.CommandTitle.Render("Commands"))

	sections = append(sections, lipgloss.StyleRunes("Details", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Filter", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Columns", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	sections = append(sections, lipgloss.StyleRunes("Edit", []int{0, 0}, styles.CommandLabelTrigger, styles.CommandLabel))
	if selectionInfo != "" {
		sections = append(sections, selectionInfo)
	}

	left := lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	return styles.CommandPanel.Width(m.width).Render(m.layoutWithMeta(left, metaContent))
}

func (m *Model) metaContent() string {
	current := "–"
	if m.currentLine > 0 {
		current = fmt.Sprintf("%d", m.currentLine)
	}

	base := fmt.Sprintf("%s / %d", current, m.totalRows)

	details := []string{}
	if m.filterActive {
		filteredOut := m.totalRows - m.filteredRows
		if filteredOut < 0 {
			filteredOut = 0
		}
		details = append(details, fmt.Sprintf("%d filtered", filteredOut))
	}

	if len(details) == 0 {
		return base
	}

	return fmt.Sprintf("%s (%s)", base, strings.Join(details, ", "))
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
		leftWidth := (m.width - 8) / 2
		rightWidth := (m.width - 8) - leftWidth
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
