package tui

import (
	"cutl/internal/editor"
	"cutl/internal/messages"
	"cutl/internal/tui/commandpanel"
	"cutl/internal/tui/cutable"
	"cutl/internal/tui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type viewState int

const (
	tableView viewState = iota
	columnInputView
	filterInputView
)

type Model struct {
	width  int
	height int
	state  viewState

	jsonlPath    string
	table        cutable.Model
	commandPanel commandpanel.Model
}

func New(jsonlPath string) *Model {
	m := &Model{
		table:        cutable.New(),
		commandPanel: commandpanel.New(),
		jsonlPath:    jsonlPath,
		state:        tableView,
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
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case tableView:
			switch msg.String() {
			case "c":
				m.state = columnInputView
				m.commandPanel.Activate(m.table.ColumnQueries())
				return m, nil
			case "f":
				m.state = filterInputView
				m.commandPanel.Activate([]string{m.table.FilterQuery()})
				return m, nil
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		case columnInputView:
			switch msg.String() {
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
			switch msg.String() {
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
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.commandPanel, cmd = m.commandPanel.Update(msg)
	cmds = append(cmds, cmd)

	m.commandPanel.SetMeta(
		m.table.TotalRows(),
		m.table.FilteredRows(),
		m.table.SelectedOriginalLine(),
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
	m.table.SetHeight(tableHeight)

	sections = append(sections, m.table.View())
	sections = append(sections, "")
	sections = append(sections, commandPanelView)

	return styles.App.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}
