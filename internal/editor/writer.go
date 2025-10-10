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
		log.Error("Fehler beim Öffnen der JSONL-Datei zum Schreiben:", "error", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, entry := range entries {
		data, err := json.Marshal(entry.Data)
		if err != nil {
			log.Error("Fehler beim Serialisieren eines JSON-Objekts:", "error", err)
			return err
		}

		if _, err := writer.Write(data); err != nil {
			log.Error("Fehler beim Schreiben der JSONL-Datei:", "error", err)
			return err
		}

		if err := writer.WriteByte('\n'); err != nil {
			log.Error("Fehler beim Hinzufügen eines Zeilenumbruchs:", "error", err)
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		log.Error("Fehler beim Abschließen des Schreibens der JSONL-Datei:", "error", err)
		return err
	}

	log.Infof("Erfolgreich %d JSON-Objekte in %s geschrieben.", len(entries), filePath)

	return nil
}
