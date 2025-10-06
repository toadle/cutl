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
					columns = append(columns, table.Column{Title: k, Width: (m.width - 16) / 5})
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