package cutable

import (
	"cutl/internal/messages"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/itchyny/gojq"
)

type Model struct {
	height        int
	width         int
	table         table.Model
	columnQueries []string
	rawData       []any
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func New() Model {
	t := table.New(
		table.WithFocused(true),
	)
	t.SetStyles(defaultStyles())

	m := Model{
		table:         t,
		columnQueries: []string{}, // Initialisiere leeres Array
	}

	return m
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateColumnWidths()
	case messages.ColumnQueryChanged:
		m.columnQueries = msg.Queries
		m.rebuildTable()
	case messages.InputFileLoaded:
		log.Debugf("Received InputFileLoaded message with %d entries.", len(msg.Content))
		m.rawData = msg.Content

		if len(m.columnQueries) == 0 {
			if first, ok := m.rawData[0].(map[string]interface{}); ok {
				m.columnQueries = discoverInitialColumnQueries(first)
			}
		}
		m.rebuildTable()
	}

	m.table, msg = m.table.Update(msg)

	return *m, nil
}

func (m *Model) rebuildTable() {
	var columns []table.Column
	for _, q := range m.columnQueries {
		columns = append(columns, table.Column{Title: q, Width: 10})
	}

	var rows []table.Row
	for _, item := range m.rawData {
		var row []string
		for _, col := range columns {
			query, err := gojq.Parse(col.Title)
			if err != nil {
				log.Errorf("Error parsing jq query '%s': %v", col.Title, err)
				row = append(row, "ERR:PARSE")
				continue
			}

			iter := query.Run(item) // or query.RunWithContext(context.Background(), item)
			val := ""
			v, ok := iter.Next()
			if !ok {
				// No value found
				val = ""
			} else if err, ok := v.(error); ok {
				log.Errorf("Error executing jq query '%s': %v", col.Title, err)
				val = "ERR:EXEC"
			} else {
				switch v := v.(type) {
				case float64:
					val = fmt.Sprintf("%.0f", v)
				default:
					val = fmt.Sprintf("%v", v)
				}
			}
			row = append(row, val)
		}
		rows = append(rows, table.Row(row))
	}

	h := m.table.Height()

	// Create a new table model
	newTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(h),
	)

	newTable.SetStyles(defaultStyles())
	m.table = newTable
	m.updateColumnWidths()
}

func defaultStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	return s
}

func (m *Model) View() string {
	return m.table.View()
}

func (m *Model) ColumnQueries() []string {
	return m.columnQueries
}

func (m *Model) SetHeight(height int) {
	m.table.SetHeight(height)
}

func (m *Model) updateColumnWidths() {
	if m.width == 0 || len(m.table.Rows()) == 0 {
		return
	}

	columns := m.table.Columns()
	rows := m.table.Rows()
	numColumns := len(columns)
	if numColumns == 0 {
		return
	}

	// Calculate ideal widths
	idealWidths := make([]int, numColumns)
	for i, col := range columns {
		idealWidths[i] = len(col.Title)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < numColumns {
				if len(cell) > idealWidths[i] {
					idealWidths[i] = len(cell)
				}
			}
		}
	}

	totalIdealWidth := 0
	for _, w := range idealWidths {
		totalIdealWidth += w
	}

	// availableWidth := m.width - (numColumns * 3) - 2 // Adjust for padding/borders
	availableWidth := m.width - 16 // Adjust for padding/borders

	if totalIdealWidth >= availableWidth {
		// Shrink columns proportionally
		for i := range columns {
			columns[i].Width = (idealWidths[i] * availableWidth) / totalIdealWidth
		}
	} else {
		// Grow columns, distribute extra space
		extraSpace := availableWidth - totalIdealWidth
		growableColumns := 0
		for _, w := range idealWidths {
			if w > 0 { // Only consider columns with content
				growableColumns++
			}
		}

		if growableColumns > 0 {
			extraPerColumn := extraSpace / growableColumns
			for i := range columns {
				columns[i].Width = idealWidths[i]
				if idealWidths[i] > 0 {
					columns[i].Width += extraPerColumn
				}
			}
			// Distribute remainder
			remainder := extraSpace % growableColumns
			for i := 0; i < remainder; i++ {
				if i < len(columns) {
					columns[i].Width++
				}
			}
		} else {
			// Fallback if no growable columns
			for i := range columns {
				columns[i].Width = availableWidth / numColumns
			}
		}
	}

	m.table.SetColumns(columns)
}

func discoverInitialColumnQueries(first map[string]interface{}) []string {
	var keys []string
	for k := range first {
		if len(k) > 0 && k[0] != '_' {
			keys = append(keys, "."+k)
		}
	}

	// Sortiere Keys: id, title, text zuerst, dann alphabetisch
	sort.Slice(keys, func(i, j int) bool {
		a := keys[i]
		b := keys[j]

		// Extract last part of the query for priority sorting
		aName := a[strings.LastIndex(a, ".")+1:]
		bName := b[strings.LastIndex(b, ".")+1:]

		aPrio := 2
		bPrio := 2

		if aName == "id" {
			aPrio = 0
		} else if aName == "title" || aName == "text" {
			aPrio = 1
		}

		if bName == "id" {
			bPrio = 0
		} else if bName == "title" || bName == "text" {
			bPrio = 1
		}

		if aPrio != bPrio {
			return aPrio < bPrio
		}

		return a < b
	})

	if len(keys) > 5 {
		return keys[:5]
	}
	return keys
}