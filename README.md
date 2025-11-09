# CUTL - Terminal JSONL Dataset Editor

CUTL is a fast, interactive TUI tool to view, filter, and edit JSONL (JSON Lines) files – ideal for NLP and machine learning datasets, or any structured line-oriented data.

Most CUTL operations—including filtering, column selection, and row manipulation—are powered by [jq](https://stedolan.github.io/jq/) selectors and syntax. Users familiar with jq will find expressive power for querying and editing structured data, and all filtering follows jq-compatible rules. For a primer, see the jq [manual](https://stedolan.github.io/jq/manual/).

## Features

- Interactive table view for large JSONL files
- Live filtering and JQ-style queries
- Easy field/row editing, supports multi-line edit
- Keyboard-friendly navigation (vim- and arrow keys)
- Batch delete, mark/clear, save back to file
- Detail and column configuration views
- Works anywhere Go runs (no runtime dependencies)

## Quick Start

```bash
git clone <repository-url>
cd cutl
go mod tidy
go build -o cutl
./cutl --input path/to/data.jsonl
```

## Usage Example

```bash
./cutl --input data.jsonl         # Open and edit
diff data.jsonl data.jsonl.bak    # After save, compare your edits
```

For keyboard shortcuts, see in-app help or [manual](MANUAL.md) if available.

## Contributing

Pull requests are welcome. Please ensure new features include tests, follow Go code conventions, and update documentation as needed.

## License

Specify license file in repository root (e.g., MIT, Apache-2.0, etc.).

---

For detailed usage and advanced features, refer to the full manual or in-app help.
