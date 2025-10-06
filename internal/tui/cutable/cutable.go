package cutable

import (
	"cutl/internal/messages"
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Model struct {
	height         int
	width         int
	table         table.Model
	columnQueries []string
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func New() Model {
	m := Model{
		table: table.New(
			table.WithFocused(true),
		),
		columnQueries: []string{}, // Initialisiere leeres Array
	}

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
	m.table.SetStyles(s)

	return m
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateColumnWidths()
	case messages.InputFileLoaded:
		log.Debugf("Received InputFileLoaded message with %d entries.", len(msg.Content))

		var columns []table.Column
		var rows []table.Row

		if len(msg.Content) > 0 {
			if len(m.columnQueries) == 0 {
				if first, ok := msg.Content[0].(map[string]interface{}); ok {
					m.columnQueries = discoverInitialColumnQueries(first)
				}

				i := 0
				for _, k := range m.columnQueries {
					if i >= 5 {
						break
					}
					columns = append(columns, table.Column{Title: k, Width: 10})
					i++
				}

				for _, item := range msg.Content {
					if obj, ok := item.(map[string]interface{}); ok {
						var row []string
						for _, col := range columns {
							val := ""
							if v, ok := obj[col.Title]; ok {
								switch v := v.(type) {
								case float64:
									val = fmt.Sprintf("%.0f", v) // Formatierung als ganze Zahl
								default:
									val = fmt.Sprintf("%v", v)   // Konvertiere Wert zu String
								}
							}
							row = append(row, val)
						}
						rows = append(rows, table.Row(row))
					}
				}

				m.table.SetColumns(columns)
				m.table.SetRows(rows)
				m.updateColumnWidths()
			} else {
				log.Warn("First entry is not a map[string]interface{}")
			}
		}
	}

	m.table, msg = m.table.Update(msg)

	return *m, nil
}

func (m *Model) View() string {
	t := m.table
	t.SetHeight(m.height - 4)

	return t.View()
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
			keys = append(keys, k)
		}
	}

	// Sortiere Keys: id, title, text zuerst, dann alphabetisch
	sort.Slice(keys, func(i, j int) bool {
		a := keys[i]
		b := keys[j]

		aPrio := 2
		bPrio := 2

		if a == "id" {
			aPrio = 0
		} else if a == "title" {
			aPrio = 1
		} else if a == "text" {
			aPrio = 1
		}

		if b == "id" {
			bPrio = 0
		} else if b == "title" {
			bPrio = 1
		} else if b == "text" {
			bPrio = 1
		}

		if aPrio != bPrio {
			return aPrio < bPrio
		}

		return a < b
	})
	return keys
}