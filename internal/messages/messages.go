package messages

type ColumnQueryChanged struct {
	Queries []string
}

type InputFileLoaded struct {
	Content []any
}

type InputFileLoadError struct {
	Error error
}