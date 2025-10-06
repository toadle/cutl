package editor

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/charmbracelet/log"
)

func LoadJSONL(filePath string) ([]any, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Fehler beim Öffnen der JSONL-Datei:", "error", err)
		return nil, err
	}
	defer file.Close()

	var result []any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var obj any
		if err := json.Unmarshal(line, &obj); err == nil {
			result = append(result, obj)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error("Fehler beim Lesen der JSONL-Datei:", "error", err)
		return nil, err
	}
	log.Debugf("Erfolgreich %d JSON-Objekte aus der Datei %s geladen.", len(result), filePath)

	return result, nil
}