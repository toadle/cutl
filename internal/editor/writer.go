package editor

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/charmbracelet/log"
)

func WriteJSONL(filePath string, entries []Entry) error {
	file, err := os.Create(filePath)
	if err != nil {
		log.Error("Failed to open JSONL file for writing:", "error", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, entry := range entries {
		data, err := json.Marshal(entry.Data)
		if err != nil {
			log.Error("Failed to serialize JSON object:", "error", err)
			return err
		}

		if _, err := writer.Write(data); err != nil {
			log.Error("Failed to write JSONL file:", "error", err)
			return err
		}

		if err := writer.WriteByte('\n'); err != nil {
			log.Error("Failed to append newline:", "error", err)
			return err
		}
	}

	if err = writer.Flush(); err != nil {
		log.Error("Failed to flush JSONL writer:", "error", err)
		return err
	}

	log.Debugf("Erfolgreich %d JSON-Objekte in %s geschrieben.", len(entries), filePath)

	return nil
}
