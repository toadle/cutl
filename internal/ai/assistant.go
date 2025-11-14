package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/log"
	openai "github.com/sashabaranov/go-openai"
)

var (
	// ErrMissingAPIKey is returned when OPENAI_API_KEY is not available in the environment.
	ErrMissingAPIKey = errors.New("OPENAI_API_KEY not set")
)

const (
	defaultModel          = "gpt-4.1-mini"
	defaultRequestTimeout = 30 * time.Second
)

// Client wraps the OpenAI API client with CUTL specific helpers.
type Client struct {
	api   *openai.Client
	model string
}

// NewFromEnv attempts to create a client using environment variables.
// It honours OPENAI_API_KEY (required), CUTL_OPENAI_MODEL (optional) and
// OPENAI_BASE_URL (optional, useful for Azure/OpenAI-compatible endpoints).
func NewFromEnv() (*Client, error) {
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		return nil, ErrMissingAPIKey
	}

	cfg := openai.DefaultConfig(key)
	if baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	model := strings.TrimSpace(os.Getenv("CUTL_OPENAI_MODEL"))
	if model == "" {
		model = defaultModel
	}

	return &Client{
		api:   openai.NewClientWithConfig(cfg),
		model: model,
	}, nil
}

// FilterRequest bundles the information we send to the assistant so it can
// craft an appropriate jq filter.
type FilterRequest struct {
	Prompt      string
	SampleJSON  string
	ColumnHints []string
}

// GenerateFilterQuery asks the LLM to return a jq filter based on the provided
// prompt and the structural hints from the currently selected row.
func (c *Client) GenerateFilterQuery(ctx context.Context, req FilterRequest) (string, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return "", errors.New("prompt is empty")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	columnHint := "(no column hints)"
	if len(req.ColumnHints) > 0 {
		columnHint = strings.Join(req.ColumnHints, ", ")
	}

	sampleJSON := strings.TrimSpace(req.SampleJSON)
	if sampleJSON == "" {
		sampleJSON = "{}"
	}

	userContent := fmt.Sprintf("User intent:\n%s\n\nAnonymized JSONL row structure:\n%s\n\nColumn hints: %s\n\nReturn ONLY a valid jq boolean expression that can be placed inside select(...). Use '.' as the current row. Prefer piping to functions, e.g. (.input // \"\" | ascii_downcase | contains(\"paket\")) or (.language == \"fr\"). Do not emit select(), jq prefixes, extra explanation, or invalid function signatures.",
		req.Prompt,
		sampleJSON,
		columnHint,
	)

	log.Debug("Sending AI filter prompt",
		"model", c.model,
		"prompt", req.Prompt,
		"columns", strings.Join(req.ColumnHints, ", "),
		"sample", sampleJSON,
	)

	resp, err := c.api.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Temperature: 0,
		MaxTokens:   256,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You convert natural language requests and JSON examples into jq boolean expressions that are safe to place inside select(...). Use '.' to reference the current JSON object, prefer piping (e.g. (.input // \"\" | ascii_downcase | contains(\"paket\"))). Never emit select(), jq prefixes, invalid function signatures, or commentary.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userContent,
			},
		},
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("assistant returned no choices")
	}

	query := normalizeAssistantQuery(resp.Choices[0].Message.Content)
	log.Debug("Received AI filter response", "query", query)
	if query == "" {
		return "", errors.New("assistant returned an empty response")
	}

	return query, nil
}

func normalizeAssistantQuery(raw string) string {
	cleaned := stripCodeFences(raw)
	cleaned = strings.TrimSpace(cleaned)
	cleaned = stripPrefixesCI(cleaned, []string{
		"jq",
		"query:",
		"filter:",
		"answer:",
		"result:",
		"expression:",
		"condition:",
		"the jq query is",
		"the jq expression is",
		"the condition is",
		"here is the jq query:",
		"here is the condition:",
	})
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.Trim(cleaned, "`")
	cleaned = strings.TrimSpace(cleaned)

	if strings.HasPrefix(cleaned, ". |") {
		cleaned = strings.TrimSpace(cleaned[3:])
	} else if strings.HasPrefix(cleaned, ".|") {
		cleaned = strings.TrimSpace(cleaned[2:])
	} else if strings.HasPrefix(cleaned, "|") {
		cleaned = strings.TrimSpace(cleaned[1:])
	}

	cleaned = extractSelectCondition(cleaned)
	cleaned = strings.TrimSpace(cleaned)

	if strings.HasSuffix(cleaned, ";") {
		cleaned = strings.TrimSpace(cleaned[:len(cleaned)-1])
	}

	return cleaned
}

func stripCodeFences(s string) string {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	content := trimmed[3:]
	if newline := strings.Index(content, "\n"); newline >= 0 {
		content = content[newline+1:]
	} else {
		content = ""
	}
	if idx := strings.LastIndex(content, "```"); idx >= 0 {
		content = content[:idx]
	}
	return strings.TrimSpace(content)
}

func stripPrefixesCI(s string, prefixes []string) string {
	trimmed := strings.TrimSpace(s)
	for {
		changed := false
		for _, prefix := range prefixes {
			if len(trimmed) >= len(prefix) && strings.EqualFold(trimmed[:len(prefix)], prefix) {
				trimmed = strings.TrimSpace(trimmed[len(prefix):])
				changed = true
				break
			}
		}
		if !changed {
			break
		}
	}
	return trimmed
}

func extractSelectCondition(s string) string {
	trimmed := strings.TrimSpace(s)
	lower := strings.ToLower(trimmed)
	idx := strings.Index(lower, "select")
	if idx == -1 {
		return trimmed
	}
	i := idx + len("select")
	for i < len(trimmed) && unicode.IsSpace(rune(trimmed[i])) {
		i++
	}
	if i >= len(trimmed) || trimmed[i] != '(' {
		return trimmed
	}
	i++
	depth := 1
	start := i
	for i < len(trimmed) {
		switch trimmed[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return strings.TrimSpace(trimmed[start:i])
			}
		}
		i++
	}
	return strings.TrimSpace(trimmed)
}
