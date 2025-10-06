package messages


type InputFileLoaded struct {
	Content []any
}

type InputFileLoadError struct {
	Error error
}