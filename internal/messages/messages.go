package messages

import "cutl/internal/editor"

type ColumnQueryChanged struct {
	Queries []string
}

type FilterQueryChanged struct {
	Query string
}

type InputFileLoaded struct {
	Content []editor.Entry
}

type InputFileLoadError struct {
	Error error
}
