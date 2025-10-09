package editor

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/charmbracelet/log"
)

type Entry struct {
	Data any
	Line int
}

func LoadJSONL(filePath string) ([]Entry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Fehler beim Öffnen der JSONL-Datei:", "error", err)
		return nil, err
	}
	defer file.Close()

	var result []Entry
	lineNumber := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var obj any
		if err := json.Unmarshal(line, &obj); err == nil {
			result = append(result, Entry{Data: obj, Line: lineNumber})
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error("Fehler beim Lesen der JSONL-Datei:", "error", err)
		return nil, err
	}
	log.Debugf("Erfolgreich %d JSON-Objekte aus der Datei %s geladen.", len(result), filePath)

	return result, nil
}
