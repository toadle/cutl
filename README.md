
<img width="200" alt="53C8B79A-E6B6-4B9E-B275-23913333D816" src="https://github.com/user-attachments/assets/cd05cd74-c77a-45b9-b0de-bf142a05206a" />


# cutl - Terminal JSONL Dataset Editor

cutl is a fast, interactive TUI tool to view, filter, and edit JSONL (JSON Lines) files – ideal for NLP and machine learning datasets, or any structured line-oriented data.
It's goal is to be pleasant to use. The primary intended use-case is to work in NLP dataset / corpus files that need to be refined, filtered or edited. (e.g. the ones created for https://spacy.io)

Most cutl operations—including filtering, column selection, and row manipulation—are powered by [jq](https://stedolan.github.io/jq/) selectors and syntax. Users familiar with jq will find expressive power for querying and editing structured data, and all filtering follows jq-compatible rules. For a primer, see the jq [manual](https://stedolan.github.io/jq/manual/).

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
```

For keyboard shortcuts, see in-app help.

## Contributing

Pull requests are welcome. Please ensure new features include tests, follow Go code conventions, and update documentation as needed.

## License

MIT
