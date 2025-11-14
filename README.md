
<img height="200" alt="cutl-logo" src="https://github.com/user-attachments/assets/e4c799fe-3993-4c3d-8f0e-ece573162816" />


# cutl - Your friendly jq-based JSONL Dataset Editor

cutl is a fast, interactive TUI tool to view, filter, and edit JSONL (JSON Lines) files – ideal for NLP and machine learning datasets, or any structured line-oriented data.
It's goal is to be pleasant to use. The primary intended use-case is to work in NLP dataset / corpus files that need to be refined, filtered or edited. (e.g. the ones created for https://spacy.io)

Most cutl operations—including filtering, column selection, and row manipulation—are powered by [jq](https://stedolan.github.io/jq/) selectors and syntax. Users familiar with jq will find expressive power for querying and editing structured data, and all filtering follows jq-compatible rules. For a primer, see the jq [manual](https://stedolan.github.io/jq/manual/).


![CleanShot 2025-11-11 at 17 08 20](https://github.com/user-attachments/assets/8acc4747-4f64-47d7-8eaa-4bbead27b72c)


## Features

- Interactive table view for large JSONL files
- Live filtering and JQ-style queries
- Optional AI-assisted filter prompts (requires `OPENAI_API_KEY`)
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

## AI-assisted filtering

If you export `OPENAI_API_KEY`, cutl unlocks an “AI Filter” prompt that can turn natural language instructions into jq filters:

```bash
export OPENAI_API_KEY=sk-your-key
# Optional overrides
export CUTL_OPENAI_MODEL=gpt-4.1-mini
export OPENAI_BASE_URL=https://api.openai.com/v1
```

While viewing the table, press `P` to open the AI prompt. Describe what you want to filter (e.g. “rows where language is German and score > 0.8”) and press `Enter`. The currently selected row is sent as context so the model understands your data structure. The assistant responds with a jq expression that is immediately applied as the active filter. If no API key is available, the shortcut is hidden and the regular filtering workflow stays unchanged.

## Contributing

Pull requests are welcome. The tool is heavily vibe-coded up to this point, so don't expect the finest of code.

## License

MIT
