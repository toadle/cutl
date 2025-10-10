package tui

import (
	"cutl/internal/editor"
	"cutl/internal/messages"
	"cutl/internal/tui/commandpanel"
	"cutl/internal/tui/cutable"
	"cutl/internal/tui/styles"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type viewState int

const (
	tableView viewState = iota
	columnInputView
	filterInputView
	detailView
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
}

func New(jsonlPath string) *Model {
	m := &Model{
		table:        cutable.New(),
		commandPanel: commandpanel.New(),
		jsonlPath:    jsonlPath,
		state:        tableView,
	}

	m.detailViewport = viewport.New(0, 0)

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
					m.setStatusMessage("Speichere…", false)
					cmds = append(cmds, m.pendingWriteCmd)
					m.pendingWriteCmd = nil
				}
			case "n", "N", "esc":
				m.confirmationActive = false
				m.pendingWriteCmd = nil
				m.setStatusMessage("Speichern abgebrochen", true)
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
		m.setStatusMessage(fmt.Sprintf("Gespeichert: %s (%d Zeilen)", filename, msg.Count), true)
	case messages.InputFileWriteError:
		log.Errorf("Failed to write JSONL file %s: %v", m.jsonlPath, msg.Error)
		m.setStatusMessage(fmt.Sprintf("Speichern fehlgeschlagen: %v", msg.Error), true)
	}

	if !skipTableUpdate && (m.state == tableView || m.state == detailView) {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
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
		return detailStyle.Render(styles.Text.Render("Kein Eintrag ausgewählt."))
	}

	info := styles.InfoLabel.Render(fmt.Sprintf("Zeile %d — drücke D oder ESC für die Tabelle", entry.Line))
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
			content = styles.Text.Copy().Render(fmt.Sprintf("Fehler beim Formatieren: %v", err))
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
	prompt := fmt.Sprintf("Änderungen nach %s schreiben? (y/N)", filename)
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
