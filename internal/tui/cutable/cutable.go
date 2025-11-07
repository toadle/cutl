package cutable

import (
	"cutl/internal/editor"
	"cutl/internal/messages"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/itchyny/gojq"
)

type Model struct {
	height          int
	width           int
	table           table.Model
	columnQueries   []string
	filterQuery     string
	rawEntries      []editor.Entry
	filteredEntries []editor.Entry
	marked          map[int]struct{}
	sortColumn      int
	sortAscending   bool
}

const (
	paddingWidthOffset = 18
)

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
		marked:        make(map[int]struct{}),
		sortColumn:    -1,
		sortAscending: true,
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
	case messages.FilterQueryChanged:
		m.filterQuery = msg.Query
		m.rebuildTable()
	case messages.SortByColumn:
		if msg.ColumnIndex == m.sortColumn {
			m.sortAscending = !m.sortAscending
		} else {
			m.sortColumn = msg.ColumnIndex
			m.sortAscending = true
		}
		m.rebuildTableWithSort()
	case messages.InputFileLoaded:
		log.Debugf("Received InputFileLoaded message with %d entries.", len(msg.Content))
		m.rawEntries = msg.Content

		// Only discover columns if none are set (they might be loaded from config)
		if len(m.columnQueries) == 0 && len(m.rawEntries) > 0 {
			if first, ok := m.rawEntries[0].Data.(map[string]interface{}); ok {
				m.columnQueries = discoverInitialColumnQueries(first)
				log.Debugf("Auto-discovered columns: %v", m.columnQueries)
			}
		} else {
			log.Debugf("Using pre-configured columns: %v", m.columnQueries)
		}
		m.rebuildTable()
	}

	m.table, msg = m.table.Update(msg)

	return *m, nil
}

func (m *Model) rebuildTable() {
	m.rebuildTableInternal(true)
}

func (m *Model) rebuildTableWithSort() {
	m.rebuildTableInternal(false)
}

func (m *Model) rebuildTableInternal(preserveSelection bool) {
	selectedLine := -1
	if preserveSelection && len(m.filteredEntries) > 0 {
		if line := m.SelectedOriginalLine(); line > 0 {
			selectedLine = line
		}
	}

	showMarker := len(m.marked) > 0

	columns := make([]table.Column, 0, len(m.columnQueries)+1)
	if showMarker {
		columns = append(columns, table.Column{Title: "●", Width: 2})
	}
	for i, q := range m.columnQueries {
		title := q
		if m.sortColumn == i {
			if m.sortAscending {
				title = q + " ↑"
			} else {
				title = q + " ↓"
			}
		}
		columns = append(columns, table.Column{Title: title, Width: 10})
	}

	m.filteredEntries = m.rawEntries
	if m.filterQuery != "" {
		filterStr := fmt.Sprintf("select(%s)", m.filterQuery)
		query, err := gojq.Parse(filterStr)
		if err != nil {
			log.Errorf("Error parsing filter query '%s': %v", filterStr, err)
			m.filteredEntries = m.rawEntries
		} else {
			var (
				filtered []editor.Entry
				hadError bool
			)
			for _, entry := range m.rawEntries {
				iter := query.Run(entry.Data)
				v, ok := iter.Next()
				if !ok {
					continue
				}
				if execErr, isErr := v.(error); isErr {
					log.Errorf("Error executing filter query '%s': %v", filterStr, execErr)
					hadError = true
					break
				}
				filtered = append(filtered, entry)
			}
			if hadError {
				m.filteredEntries = m.rawEntries
			} else {
				m.filteredEntries = filtered
			}
		}
	}

	// Sort entries if a sort column is specified
	if m.sortColumn >= 0 && m.sortColumn < len(m.columnQueries) {
		m.sortEntries()
	}

	var (
		rows      []table.Row
		cursorPos = -1
	)

	for idx, entry := range m.filteredEntries {
		row := make([]string, 0, len(columns))
		if showMarker {
			row = append(row, m.markerSymbol(entry.Line))
		}

		for _, col := range m.columnQueries {
			query, err := gojq.Parse(col)
			if err != nil {
				log.Errorf("Error parsing jq query '%s': %v", col, err)
				row = append(row, "ERR:PARSE")
				continue
			}

			iter := query.Run(entry.Data)
			val := ""
			v, ok := iter.Next()
			if !ok {
				val = ""
			} else if execErr, isErr := v.(error); isErr {
				log.Errorf("Error executing jq query '%s': %v", col, execErr)
				val = "ERR:EXEC"
			} else {
				switch v := v.(type) {
				case float64:
					val = fmt.Sprintf("%.0f", v)
				case []interface{}, map[string]interface{}:
					// For arrays and objects, display as JSON string
					if jsonBytes, err := json.Marshal(v); err == nil {
						val = string(jsonBytes)
					} else {
						val = fmt.Sprintf("%v", v)
					}
				default:
					val = fmt.Sprintf("%v", v)
				}
			}
			row = append(row, val)
		}
		rows = append(rows, table.Row(row))
		if selectedLine > 0 && entry.Line == selectedLine {
			cursorPos = idx
		}
	}

	oldColumnCount := len(m.table.Columns())
	newColumnCount := len(columns)

	if newColumnCount < oldColumnCount {
		m.table.SetRows(rows)
		m.table.SetColumns(columns)
	} else {
		m.table.SetColumns(columns)
		m.table.SetRows(rows)
	}
	if preserveSelection && cursorPos >= 0 {
		m.table.SetCursor(cursorPos)
	} else if !preserveSelection {
		m.table.SetCursor(0)
	}
	m.updateColumnWidths()
}

func (m *Model) markerSymbol(line int) string {
	if _, ok := m.marked[line]; ok {
		return "●"
	}
	return ""
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

func (m *Model) FilterQuery() string {
	return m.filterQuery
}

func (m *Model) TotalRows() int {
	return len(m.rawEntries)
}

func (m *Model) FilteredRows() int {
	return len(m.filteredEntries)
}

func (m *Model) MarkedCount() int {
	return len(m.marked)
}

func (m *Model) MarkedLines() []int {
	lines := make([]int, 0, len(m.marked))
	for line := range m.marked {
		lines = append(lines, line)
	}
	return lines
}

func (m *Model) SetColumnQueries(queries []string) {
	m.columnQueries = queries
	log.Debugf("Set column queries: %v", queries)
}

func (m *Model) SelectedOriginalLine() int {
	if len(m.filteredEntries) == 0 {
		return 0
	}

	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredEntries) {
		return 0
	}

	return m.filteredEntries[cursor].Line
}

func (m *Model) SelectedEntry() *editor.Entry {
	if len(m.filteredEntries) == 0 {
		return nil
	}

	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredEntries) {
		return nil
	}

	return &m.filteredEntries[cursor]
}

func (m *Model) ToggleMarkSelected() {
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredEntries) {
		return
	}

	entry := m.filteredEntries[cursor]
	if m.marked == nil {
		m.marked = make(map[int]struct{})
	}

	if _, ok := m.marked[entry.Line]; ok {
		delete(m.marked, entry.Line)
	} else {
		m.marked[entry.Line] = struct{}{}
	}

	m.rebuildTable()
}

func (m *Model) ClearMarks() {
	if len(m.marked) == 0 {
		return
	}

	m.marked = make(map[int]struct{})
	m.rebuildTable()
}

func (m *Model) DeleteMarkedOrSelected() int {
	if len(m.rawEntries) == 0 {
		return 0
	}

	linesToDelete := make(map[int]struct{})
	if len(m.marked) > 0 {
		for line := range m.marked {
			linesToDelete[line] = struct{}{}
		}
	} else {
		selected := m.SelectedEntry()
		if selected == nil {
			return 0
		}
		linesToDelete[selected.Line] = struct{}{}
	}

	if len(linesToDelete) == 0 {
		return 0
	}

	newEntries := make([]editor.Entry, 0, len(m.rawEntries)-len(linesToDelete))
	for _, entry := range m.rawEntries {
		if _, remove := linesToDelete[entry.Line]; remove {
			continue
		}
		newEntries = append(newEntries, entry)
	}

	if len(newEntries) == len(m.rawEntries) {
		return 0
	}

	for idx := range newEntries {
		newEntries[idx].Line = idx + 1
	}

	previousCursor := m.table.Cursor()
	m.rawEntries = newEntries
	m.marked = make(map[int]struct{})
	m.rebuildTable()

	if len(m.filteredEntries) == 0 {
		m.table.SetCursor(0)
	} else {
		newCursor := previousCursor
		if newCursor >= len(m.filteredEntries) {
			newCursor = len(m.filteredEntries) - 1
		}
		if newCursor < 0 {
			newCursor = 0
		}
		m.table.SetCursor(newCursor)
	}

	return len(linesToDelete)
}

func (m *Model) Entries() []editor.Entry {
	entries := make([]editor.Entry, len(m.rawEntries))
	copy(entries, m.rawEntries)
	return entries
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

	idealWidths := make([]int, numColumns)
	for colIdx := range columns {
		width := lipgloss.Width(columns[colIdx].Title)
		for _, row := range rows {
			if colIdx < len(row) {
				cellWidth := lipgloss.Width(row[colIdx])
				if cellWidth > width {
					width = cellWidth
				}
			}
		}
		idealWidths[colIdx] = width
	}

	available := m.width - paddingWidthOffset
	if available < 0 {
		available = 0
	}

	widths := make([]int, numColumns)
	totalIdeal := 0
	for _, w := range idealWidths {
		totalIdeal += w
	}

	if totalIdeal == 0 {
		per := 0
		remainder := 0
		if numColumns > 0 {
			per = available / numColumns
			remainder = available % numColumns
		}
		for i := 0; i < numColumns; i++ {
			widths[i] = per
			if remainder > 0 {
				widths[i]++
				remainder--
			}
		}
	} else {
		remainder := available
		for i := 0; i < numColumns; i++ {
			widths[i] = (idealWidths[i] * available) / totalIdeal
			remainder -= widths[i]
		}
		for i := 0; remainder > 0 && i < numColumns; i++ {
			widths[i]++
			remainder--
		}
	}

	// Ensure minimum width of 10 characters for each column (except marker column)
	const minWidth = 10
	totalNeeded := 0
	startCol := 0
	if len(columns) > 0 && columns[0].Title == "●" {
		startCol = 1 // Skip the marker column
	}

	for i := startCol; i < numColumns; i++ {
		if widths[i] < minWidth {
			totalNeeded += minWidth - widths[i]
			widths[i] = minWidth
		}
	}

	// If we need more space, take it from the widest columns
	if totalNeeded > 0 {
		for totalNeeded > 0 && numColumns > startCol {
			// Find the widest column that can give up space (excluding marker column)
			maxIdx := -1
			maxWidth := minWidth
			for i := startCol; i < numColumns; i++ {
				if widths[i] > maxWidth {
					maxWidth = widths[i]
					maxIdx = i
				}
			}

			// If no column can give up space, break
			if maxIdx == -1 {
				break
			}

			// Take one character from the widest column
			widths[maxIdx]--
			totalNeeded--
		}
	}

	if len(columns) > 0 && columns[0].Title == "●" && len(widths) > 1 {
		maxIdx := 1
		for i := 2; i < len(widths); i++ {
			if widths[i] > widths[maxIdx] {
				maxIdx = i
			}
		}
		if widths[maxIdx] >= 2 {
			widths[maxIdx] -= 2
		} else {
			widths[maxIdx] = 0
		}
	}

	for idx := range columns {
		columns[idx].Width = widths[idx]
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

func (m *Model) sortEntries() {
	if m.sortColumn < 0 || m.sortColumn >= len(m.columnQueries) {
		return
	}

	sortQuery := m.columnQueries[m.sortColumn]
	query, err := gojq.Parse(sortQuery)
	if err != nil {
		log.Errorf("Error parsing sort query '%s': %v", sortQuery, err)
		return
	}

	sort.Slice(m.filteredEntries, func(i, j int) bool {
		valI := m.extractSortValue(query, m.filteredEntries[i])
		valJ := m.extractSortValue(query, m.filteredEntries[j])

		result := m.compareSortValues(valI, valJ)
		if m.sortAscending {
			return result
		}
		return !result
	})
}

func (m *Model) extractSortValue(query *gojq.Query, entry editor.Entry) interface{} {
	iter := query.Run(entry.Data)
	v, ok := iter.Next()
	if !ok {
		return ""
	}
	if _, isErr := v.(error); isErr {
		return ""
	}
	return v
}

func (m *Model) compareSortValues(a, b interface{}) bool {
	// Convert to strings for comparison if types don't match
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Try to compare as numbers if both look like numbers
	if aNum, aErr := parseNumber(aStr); aErr == nil {
		if bNum, bErr := parseNumber(bStr); bErr == nil {
			return aNum < bNum
		}
	}

	// String comparison as fallback
	return strings.ToLower(aStr) < strings.ToLower(bStr)
}

func parseNumber(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	var result float64
	n, err := fmt.Sscanf(s, "%f", &result)
	if n != 1 || err != nil {
		return 0, err
	}
	return result, nil
}

func (m *Model) UpdateEntries(targetLines []int, values map[string]string, singleMode bool) error {
	updatedCount := 0
	
	if singleMode && len(targetLines) == 1 {
		// Update single entry
		targetLine := targetLines[0]
		for i := range m.rawEntries {
			if m.rawEntries[i].Line == targetLine {
				if err := m.updateEntryData(&m.rawEntries[i], values); err != nil {
					return err
				}
				updatedCount++
				break
			}
		}
	} else {
		// Update multiple entries
		for _, targetLine := range targetLines {
			for i := range m.rawEntries {
				if m.rawEntries[i].Line == targetLine {
					// Only update non-empty values for multi-line edit
					nonEmptyValues := make(map[string]string)
					for col, val := range values {
						if strings.TrimSpace(val) != "" {
							nonEmptyValues[col] = val
						}
					}
					if err := m.updateEntryData(&m.rawEntries[i], nonEmptyValues); err != nil {
						return err
					}
					updatedCount++
					break
				}
			}
		}
	}

	log.Debugf("UpdateEntries: Updated %d entries out of %d targets", updatedCount, len(targetLines))

	// Rebuild the table to reflect changes
	m.rebuildTable()
	return nil
}

func (m *Model) updateEntryData(entry *editor.Entry, values map[string]string) error {
	// entry.Data is of type any, so we need to cast it to map[string]interface{}
	dataMap, ok := entry.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("entry data is not a map")
	}

	log.Debugf("updateEntryData: Updating entry line %d with %d values", entry.Line, len(values))

	// Create a copy of the data to modify
	data := make(map[string]interface{})
	for k, v := range dataMap {
		data[k] = v
	}

	// Apply the updates using jq-like path setting
	for column, value := range values {
		log.Debugf("updateEntryData: Setting %s = %s", column, value)
		if err := m.setValueAtPath(data, column, value); err != nil {
			return fmt.Errorf("failed to set value for column %s: %w", column, err)
		}
	}

	entry.Data = data
	log.Debugf("updateEntryData: Successfully updated entry line %d", entry.Line)
	return nil
}

func (m *Model) setValueAtPath(data map[string]interface{}, path, value string) error {
	// Simple implementation for basic JSON paths
	// This handles simple property access like ".name", ".age", etc.

	if !strings.HasPrefix(path, ".") {
		return fmt.Errorf("path must start with '.'")
	}

	key := strings.TrimPrefix(path, ".")

	// Handle nested paths by splitting on dots
	parts := strings.Split(key, ".")
	current := data

	// Navigate to the parent of the target field
	for i, part := range parts[:len(parts)-1] {
		if current[part] == nil {
			current[part] = make(map[string]interface{})
		}

		nested, ok := current[part].(map[string]interface{})
		if !ok {
			return fmt.Errorf("cannot navigate through non-object at part %d (%s)", i, part)
		}
		current = nested
	}

	finalKey := parts[len(parts)-1]

	// Try to parse as different types
	if value == "" {
		current[finalKey] = nil
	} else if value == "true" {
		current[finalKey] = true
	} else if value == "false" {
		current[finalKey] = false
	} else if num, err := strconv.ParseFloat(value, 64); err == nil {
		// Check if it's actually an integer
		if num == float64(int64(num)) {
			current[finalKey] = int64(num)
		} else {
			current[finalKey] = num
		}
	} else if m.looksLikeJSON(value) {
		// Try to parse as JSON (for arrays and objects)
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			current[finalKey] = jsonValue
			log.Debugf("Parsed JSON value for %s: %v", finalKey, jsonValue)
		} else {
			// If JSON parsing fails, treat as string
			current[finalKey] = value
		}
	} else {
		// Default to string
		current[finalKey] = value
	}

	return nil
}

func (m *Model) looksLikeJSON(value string) bool {
	value = strings.TrimSpace(value)
	return (strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]")) ||
		   (strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}"))
}
