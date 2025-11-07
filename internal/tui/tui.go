package tui

import (
	"cutl/internal/config"
	"cutl/internal/editor"
	"cutl/internal/messages"
	"cutl/internal/tui/commandpanel"
	"cutl/internal/tui/cutable"
	"cutl/internal/tui/styles"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/itchyny/gojq"
)

type viewState int

const (
	tableView viewState = iota
	columnInputView
	filterInputView
	detailView
	editView
)

type Model struct {
	width  int
	height int
	state  viewState

	jsonlPath               string
	table                   cutable.Model
	commandPanel            commandpanel.Model
	detailViewport          viewport.Model
	detailContent           string
	detailLine              int
	confirmationActive      bool
	pendingWriteCmd         tea.Cmd
	statusMessage           string
	clearStatusOnNextAction bool

	// Edit view fields
	editInputs      []textinput.Model
	editSingleMode  bool
	editTargetLines []int

	// Configuration
	config *config.Config
}

func New(jsonlPath string) *Model {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Warnf("Failed to load config: %v", err)
		cfg = &config.Config{Files: make(map[string]config.FileConfig)}
	}

	m := &Model{
		table:        cutable.New(),
		commandPanel: commandpanel.New(),
		jsonlPath:    jsonlPath,
		state:        tableView,
		config:       cfg,
	}

	m.detailViewport = viewport.New(0, 0)

	return m
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		// Try to load saved column configuration for this file
		if fileConfig, exists := m.config.GetFileConfig(m.jsonlPath); exists && len(fileConfig.Columns) > 0 {
			log.Debugf("Loaded saved columns for %s: %v", m.jsonlPath, fileConfig.Columns)
			m.table.SetColumnQueries(fileConfig.Columns)
		}

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
	var cmd tea.Cmd
	cmds := []tea.Cmd{}
	skipTableUpdate := false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if m.clearStatusOnNextAction && !m.confirmationActive {
			m.clearStatusMessage()
		}
		if m.confirmationActive {
			skipTableUpdate = true
			switch key {
			case "y", "Y", "enter":
				m.confirmationActive = false
				if m.pendingWriteCmd != nil {
					m.setStatusMessage("Saving…", false)
					cmds = append(cmds, m.pendingWriteCmd)
					m.pendingWriteCmd = nil
				}
			case "n", "N", "esc":
				m.confirmationActive = false
				m.pendingWriteCmd = nil
				m.setStatusMessage("Save cancelled", true)
			}
			break
		}
		switch m.state {
		case tableView:
			switch key {
			case "c":
				m.state = columnInputView
				m.commandPanel.Activate(m.table.ColumnQueries())
				return m, nil
			case "f":
				m.state = filterInputView
				m.commandPanel.Activate([]string{m.table.FilterQuery()})
				return m, nil
			case "d", "D":
				if entry := m.table.SelectedEntry(); entry != nil {
					m.state = detailView
					m.updateDetailContent(entry, true)
					return m, nil
				}
			case "e", "E":
				skipTableUpdate = true
				m.initializeEditView()
				return m, nil
			case " ":
				skipTableUpdate = true
				m.table.ToggleMarkSelected()
			case "x", "X":
				skipTableUpdate = true
				removed := m.table.DeleteMarkedOrSelected()
				if removed > 0 {
					log.Infof("Deleted %d entries", removed)
				}
			case "w", "W":
				skipTableUpdate = true
				m.requestWriteConfirmation()
			case "esc":
				if m.table.MarkedCount() > 0 {
					skipTableUpdate = true
					m.table.ClearMarks()
				}
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				skipTableUpdate = true
				if columnIndex := int(key[0] - '1'); columnIndex < len(m.table.ColumnQueries()) {
					return m, func() tea.Msg {
						return messages.SortByColumn{ColumnIndex: columnIndex}
					}
				}
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		case columnInputView:
			switch key {
			case "esc":
				m.state = tableView
				m.commandPanel.Deactivate()
			case "enter":
				m.state = tableView
				m.commandPanel.Deactivate()
				rawQueries := m.commandPanel.Value()
				queries := strings.Split(rawQueries, ",")
				for i, q := range queries {
					queries[i] = strings.TrimSpace(q)
				}

				// Save column configuration for this file
				if err := m.config.UpdateColumns(m.jsonlPath, queries); err != nil {
					log.Warnf("Failed to save column configuration: %v", err)
				} else {
					log.Debugf("Saved column configuration for %s: %v", m.jsonlPath, queries)
				}

				return m, func() tea.Msg {
					return messages.ColumnQueryChanged{
						Queries: queries,
					}
				}
			}
		case filterInputView:
			switch key {
			case "esc":
				m.state = tableView
				m.commandPanel.Deactivate()
			case "enter":
				m.state = tableView
				m.commandPanel.Deactivate()
				rawQuery := m.commandPanel.Value()

				// Sanitize input
				sanitizedQuery := strings.ReplaceAll(rawQuery, " ", " ")       // non-breaking space
				sanitizedQuery = strings.ReplaceAll(sanitizedQuery, "'", "\"") // single to double quotes

				return m, func() tea.Msg {
					return messages.FilterQueryChanged{
						Query: sanitizedQuery,
					}
				}
			}
		case detailView:
			skipTableUpdate = true
			switch key {
			case "esc", "d", "D":
				m.state = tableView
				return m, nil
			case "e", "E":
				m.initializeEditView()
				return m, nil
			case " ":
				m.table.ToggleMarkSelected()
			case "x", "X":
				removed := m.table.DeleteMarkedOrSelected()
				if removed > 0 {
					log.Infof("Deleted %d entries", removed)
					if m.table.FilteredRows() == 0 {
						m.state = tableView
						m.detailViewport.SetContent("")
						m.detailContent = ""
						m.detailLine = 0
						return m, nil
					}
					m.updateDetailContent(m.table.SelectedEntry(), true)
				}
			case "w", "W":
				m.requestWriteConfirmation()
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				if columnIndex := int(key[0] - '1'); columnIndex < len(m.table.ColumnQueries()) {
					cmds = append(cmds, func() tea.Msg {
						return messages.SortByColumn{ColumnIndex: columnIndex}
					})
				}
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		case editView:
			skipTableUpdate = true
			switch key {
			case "esc":
				m.state = tableView
				return m, nil
			case "enter":
				skipTableUpdate = true
				log.Debugf("Edit view: Enter pressed, applying edits")
				cmds = append(cmds, m.applyEdits())
				m.state = tableView
			case "tab":
				m.focusNextEditInput()
				return m, nil
			case "shift+tab":
				m.focusPrevEditInput()
				return m, nil
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case messages.InputFileWritten:
		log.Infof("Saved %d entries to %s", msg.Count, msg.Path)
		filename := filepath.Base(msg.Path)
		if filename == "" {
			filename = msg.Path
		}
		m.setStatusMessage(fmt.Sprintf("Saved: %s (%d rows)", filename, msg.Count), true)
	case messages.InputFileWriteError:
		log.Errorf("Failed to write JSONL file %s: %v", m.jsonlPath, msg.Error)
		m.setStatusMessage(fmt.Sprintf("Save failed: %v", msg.Error), true)
	case messages.EditApplied:
		log.Debugf("EditApplied message received")
		if msg.SingleMode {
			m.setStatusMessage("Entry updated", true)
		} else {
			m.setStatusMessage(fmt.Sprintf("Updated %d entries", len(msg.Lines)), true)
		}
	case messages.EditApplyError:
		log.Errorf("Failed to apply edits: %v", msg.Error)
		m.setStatusMessage(fmt.Sprintf("Edit failed: %v", msg.Error), true)
	}

	if !skipTableUpdate && (m.state == tableView || m.state == detailView) {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == editView {
		for i := range m.editInputs {
			m.editInputs[i], cmd = m.editInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.state == detailView {
		var vCmd tea.Cmd
		m.detailViewport, vCmd = m.detailViewport.Update(msg)
		if vCmd != nil {
			cmds = append(cmds, vCmd)
		}

		m.updateDetailContent(m.table.SelectedEntry(), false)
	}
	m.commandPanel, cmd = m.commandPanel.Update(msg)
	cmds = append(cmds, cmd)

	m.commandPanel.SetMeta(
		m.table.TotalRows(),
		m.table.FilteredRows(),
		m.table.SelectedOriginalLine(),
		m.table.MarkedCount(),
		m.table.FilterQuery() != "",
	)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var (
		sections []string
	)

	commandPanelView := m.commandPanel.View()
	commandPanelHeight := lipgloss.Height(commandPanelView)

	// App style has padding 2, so we subtract 4 for top and bottom padding,
	// plus 1 for the blank line.
	tableHeight := m.height - commandPanelHeight - 5
	if tableHeight < 3 {
		tableHeight = 3
	}
	m.table.SetHeight(tableHeight)

	panelStyle := styles.DetailPanel
	frameWidth, frameHeight := panelStyle.GetFrameSize()
	innerWidth := m.width - 8
	viewportWidth := innerWidth - frameWidth
	if viewportWidth < 0 {
		viewportWidth = 0
	}
	m.detailViewport.Width = viewportWidth

	viewportHeight := tableHeight - frameHeight - 2
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	m.detailViewport.Height = viewportHeight

	if m.state == detailView {
		sections = append(sections, m.renderDetailView())
	} else if m.state == editView {
		sections = append(sections, m.renderEditView())
	} else {
		sections = append(sections, m.table.View())
	}
	sections = append(sections, "")
	sections = append(sections, commandPanelView)

	return styles.App.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func (m *Model) renderDetailView() string {
	entry := m.table.SelectedEntry()
	detailStyle := styles.DetailPanel
	innerWidth := m.width - 8
	if innerWidth > 0 {
		detailStyle = detailStyle.Copy().Width(innerWidth)
	} else {
		detailStyle = detailStyle.Copy()
	}

	if entry == nil {
		return detailStyle.Render(styles.Text.Render("No entry selected."))
	}

	info := styles.InfoLabel.Render(fmt.Sprintf("Line %d — press D or ESC to return to the table", entry.Line))
	viewportView := m.detailViewport.View()

	return detailStyle.Render(lipgloss.JoinVertical(lipgloss.Left, info, "", viewportView))
}

func (m *Model) updateDetailContent(entry *editor.Entry, reset bool) {
	var (
		content string
		line    int
	)

	if entry != nil {
		formatted, err := json.MarshalIndent(entry.Data, "", "  ")
		if err != nil {
			content = styles.Text.Copy().Render(fmt.Sprintf("Error formatting entry: %v", err))
		} else {
			content = styles.Text.Copy().Render(string(formatted))
		}
		line = entry.Line
	}

	if content != m.detailContent {
		m.detailViewport.SetContent(content)
		m.detailContent = content
	}

	if reset {
		m.detailViewport.GotoTop()
	}

	m.detailLine = line
}

func (m *Model) setStatusMessage(message string, clearOnNext bool) {
	m.statusMessage = message
	m.clearStatusOnNextAction = clearOnNext
	m.commandPanel.SetStatus(message)
}

func (m *Model) clearStatusMessage() {
	if m.statusMessage == "" && !m.clearStatusOnNextAction {
		return
	}
	m.statusMessage = ""
	m.clearStatusOnNextAction = false
	m.commandPanel.SetStatus("")
}

func (m *Model) requestWriteConfirmation() {
	m.pendingWriteCmd = m.writeTableToFileCmd()
	filename := filepath.Base(m.jsonlPath)
	if filename == "" {
		filename = m.jsonlPath
	}
	prompt := fmt.Sprintf("Write changes to %s? (y/N)", filename)
	m.confirmationActive = true
	m.setStatusMessage(prompt, false)
}

func (m *Model) writeTableToFileCmd() tea.Cmd {
	entries := m.table.Entries()
	return func() tea.Msg {
		if err := editor.WriteJSONL(m.jsonlPath, entries); err != nil {
			return messages.InputFileWriteError{Error: err}
		}
		return messages.InputFileWritten{Path: m.jsonlPath, Count: len(entries)}
	}
}

func (m *Model) initializeEditView() {
	columns := m.table.ColumnQueries()
	if len(columns) == 0 {
		return
	}

	markedCount := m.table.MarkedCount()
	if markedCount > 0 {
		// Multi-line edit mode
		m.editSingleMode = false
		m.editTargetLines = m.getMarkedLines()
	} else {
		// Single line edit mode
		m.editSingleMode = true
		if entry := m.table.SelectedEntry(); entry != nil {
			m.editTargetLines = []int{entry.Line}
		} else {
			return
		}
	}

	// Create text inputs for each column
	m.editInputs = make([]textinput.Model, len(columns))
	for i, col := range columns {
		input := textinput.New()
		input.Placeholder = fmt.Sprintf("Enter value for %s", col)
		input.CharLimit = 500
		input.Width = 50

		// Pre-fill for single line edit
		if m.editSingleMode && len(m.editTargetLines) > 0 {
			if entry := m.table.SelectedEntry(); entry != nil {
				if value := m.extractColumnValue(entry, col); value != "" {
					input.SetValue(value)
				}
			}
		}

		if i == 0 {
			input.Focus()
		}
		m.editInputs[i] = input
	}

	m.state = editView
}

func (m *Model) getMarkedLines() []int {
	return m.table.MarkedLines()
}

func (m *Model) extractColumnValue(entry *editor.Entry, columnQuery string) string {
	// Use jq to extract the value from the entry
	query, err := gojq.Parse(columnQuery)
	if err != nil {
		return ""
	}

	iter := query.Run(entry.Data)
	v, ok := iter.Next()
	if !ok {
		return ""
	}
	if _, isErr := v.(error); isErr {
		return ""
	}

	// For arrays and objects, return JSON representation
	switch val := v.(type) {
	case []interface{}, map[string]interface{}:
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (m *Model) focusNextEditInput() {
	for i, input := range m.editInputs {
		if input.Focused() {
			input.Blur()
			m.editInputs[i] = input

			nextIdx := (i + 1) % len(m.editInputs)
			m.editInputs[nextIdx].Focus()
			break
		}
	}
}

func (m *Model) focusPrevEditInput() {
	for i, input := range m.editInputs {
		if input.Focused() {
			input.Blur()
			m.editInputs[i] = input

			prevIdx := (i - 1 + len(m.editInputs)) % len(m.editInputs)
			m.editInputs[prevIdx].Focus()
			break
		}
	}
}

func (m *Model) renderEditView() string {
	columns := m.table.ColumnQueries()
	if len(columns) == 0 || len(m.editInputs) == 0 {
		return styles.Text.Render("No columns to edit")
	}

	var title string
	if m.editSingleMode {
		title = fmt.Sprintf("Edit Single Entry (Line %d)", m.editTargetLines[0])
	} else {
		title = fmt.Sprintf("Edit Multiple Entries (%d lines)", len(m.editTargetLines))
	}

	sections := []string{
		styles.InfoLabel.Render(title),
		"",
	}

	for i, col := range columns {
		if i < len(m.editInputs) {
			sections = append(sections,
				fmt.Sprintf("%s:", col),
				m.editInputs[i].View(),
				"",
			)
		}
	}

	sections = append(sections,
		styles.InfoLabel.Render("Press Enter to save, ESC to cancel, Tab/Shift+Tab to navigate"),
	)

	return strings.Join(sections, "\n")
}

func (m *Model) applyEdits() tea.Cmd {
	return func() tea.Msg {
		log.Debugf("applyEdits: Starting to apply edits")
		columns := m.table.ColumnQueries()
		log.Debugf("applyEdits: Using %d columns", len(columns))

		// Get the values from inputs
		values := make(map[string]string)
		for i, col := range columns {
			if i < len(m.editInputs) {
				value := strings.TrimSpace(m.editInputs[i].Value())
				if value != "" || m.editSingleMode {
					values[col] = value
					log.Debugf("applyEdits: Adding value %s = %s", col, value)
				}
			}
		}

		log.Debugf("applyEdits: Calling UpdateEntries with %d target lines, %d values, singleMode=%t", 
			len(m.editTargetLines), len(values), m.editSingleMode)

		// Apply changes to the table data
		if err := m.table.UpdateEntries(m.editTargetLines, values, m.editSingleMode); err != nil {
			log.Errorf("applyEdits: UpdateEntries failed: %v", err)
			return messages.EditApplyError{Error: err}
		}

		log.Debugf("applyEdits: Successfully applied edits")
		return messages.EditApplied{
			SingleMode: m.editSingleMode,
			Lines:      m.editTargetLines,
			Values:     values,
		}
	}
}
