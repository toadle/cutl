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

type InputFileWritten struct {
	Path  string
	Count int
}

type InputFileWriteError struct {
	Error error
}

type SortByColumn struct {
	ColumnIndex int
}
