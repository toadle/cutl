# CUTL - Command Line JSONL Editor

A powerful Terminal User Interface (TUI) application for viewing, filtering, editing, and manipulating JSONL (JSON Lines) files.

## Features

### 📊 Interactive Table View
- Display JSONL data in a clean, navigable table format
- Customizable column display using JQ-style queries
- Real-time filtering with JQ expressions
- Column sorting (press 1-9 to sort by visible columns)
- Visual sort indicators (↑/↓) in column headers

### ✏️ Data Editing
- **Single Line Edit**: Press `E` to edit the selected entry with pre-filled values
- **Multi-Line Edit**: Mark multiple lines with `Space`, then press `E` to batch edit
- Tab/Shift+Tab navigation between fields
- Automatic type detection (string, number, boolean, null)
- Support for nested JSON paths (e.g., `.meta.amount`, `.address.city`)

### 🗂️ Data Management
- **Delete entries**: Press `X` to delete selected or marked entries
- **Save changes**: Press `W` to write modified data back to file
- **Mark/Unmark**: Use `Space` to mark entries for batch operations
- **Clear marks**: Press `ESC` to clear all marks

### 🔍 Navigation & Views
- **Detail View**: Press `D` to see formatted JSON of selected entry
- **Filter Data**: Press `F` to apply JQ filter expressions
- **Column Setup**: Press `C` to configure visible columns
- Arrow keys and vim-like navigation

## Installation

### Prerequisites
- Go 1.19 or later

### Build from Source
```bash
git clone <repository-url>
cd cutl
go mod tidy
go build -o cutl
```

## Usage

### Basic Usage
```bash
# View a JSONL file
./cutl --input data.jsonl

# With debug logging
./cutl --input data.jsonl --debug
```

### Keyboard Shortcuts

#### Main Navigation
- `↑/↓` or `j/k` - Navigate table rows
- `←/→` or `h/l` - Navigate table columns (when applicable)
- `Space` - Mark/unmark current row for batch operations
- `ESC` - Clear all marks or return to table view
- `Q` or `Ctrl+C` - Quit application

#### Data Operations
- `E` - Edit current entry (or marked entries)
- `X` - Delete current entry (or marked entries) with confirmation
- `W` - Write changes to file with Y/N confirmation
- `D` - View detailed JSON of current entry

#### View Configuration
- `C` - Configure visible columns (comma-separated JQ expressions)
- `F` - Set filter expression (JQ query to filter rows)
- `1-9` - Sort by column number (press same key to toggle asc/desc)

#### Edit Mode
- `Tab/Shift+Tab` - Navigate between input fields
- `Enter` - Apply changes and return to table
- `ESC` - Cancel editing and return to table

### Column Configuration Examples
```bash
# Show basic fields
.id, .name, .email

# Show nested data
.user.name, .meta.created_at, .stats.count

# Show array elements
.tags[0], .categories[]

# Complex expressions
.price * .quantity, (.total // 0)
```

### Filter Examples
```bash
# Show only active users
.status == "active"

# Show entries with price > 100
.price > 100

# Show entries from last month
.created_at | strptime("%Y-%m-%d") | mktime > (now - 2592000)

# Show entries with specific tags
.tags | contains(["important"])
```

## File Format

CUTL works with JSONL (JSON Lines) files where each line contains a valid JSON object:

```jsonl
{"id": 1, "name": "Alice", "email": "alice@example.com", "active": true}
{"id": 2, "name": "Bob", "email": "bob@example.com", "active": false}
{"id": 3, "name": "Charlie", "email": "charlie@example.com", "active": true}
```

## Features in Detail

### Editing Capabilities

#### Single Entry Editing
1. Navigate to desired row
2. Press `E` to enter edit mode
3. Fields are pre-filled with current values
4. Modify as needed and press `Enter` to save

#### Batch Editing
1. Mark multiple entries with `Space`
2. Press `E` to enter batch edit mode
3. Enter new values (empty fields remain unchanged)
4. Press `Enter` to apply changes to all marked entries

### Data Types
The editor automatically detects and preserves data types:
- **Strings**: Regular text values
- **Numbers**: Integer and floating-point values
- **Booleans**: `true`/`false` values
- **Null**: Empty fields become `null`
- **Objects**: Nested JSON structures (via dot notation)

### Sorting
- Press number keys `1-9` to sort by the corresponding visible column
- Press the same key again to reverse sort order
- Sort indicators (↑/↓) appear in column headers
- Sorting is visual-only and doesn't affect save order

## Debugging

Enable debug logging to troubleshoot issues:
```bash
./cutl --input data.jsonl --debug

# View debug log in separate terminal
tail -f debug.log
```

## Examples

### Basic Workflow
1. **Open file**: `./cutl --input customers.jsonl`
2. **Set columns**: Press `C`, enter `.id, .name, .email, .status`
3. **Filter data**: Press `F`, enter `.status == "active"`
4. **Edit entry**: Navigate to row, press `E`, modify values, press `Enter`
5. **Save changes**: Press `W`, confirm with `Y`

### Batch Operations
1. **Mark entries**: Use `Space` to mark multiple rows
2. **Batch edit**: Press `E`, enter new values, press `Enter`
3. **Batch delete**: Press `X`, confirm deletion
4. **Save changes**: Press `W` to persist changes

## Tips & Tricks

- Use `.` to show all fields when configuring columns
- Combine filters with logical operators: `.active and .age > 18`
- Use `//` for default values: `.price // 0`
- Navigate large datasets efficiently with filtering before editing
- The detail view (`D`) is helpful for understanding data structure
- Sort by different columns to find patterns in your data
- Mark entries visually before batch operations to avoid mistakes

## Requirements

- Terminal with UTF-8 support
- Minimum terminal size: 80x24 characters
- JSONL files with valid JSON objects per line

## License

[Add your license information here]

## Contributing

[Add contributing guidelines here]