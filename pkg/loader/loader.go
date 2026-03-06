// Package loader provides functions to load localized messages from various sources
// (JSON files, directories, Go maps) into an i18n Bundle.
package loader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"digital.vasic.i18n/pkg/i18n"
)

// LoadJSON loads messages from a single JSON file into the bundle for the given language.
func LoadJSON(bundle *i18n.Bundle, lang, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return err
	}
	bundle.AddMessages(lang, messages)
	return nil
}

// LoadJSONDir loads all JSON files from a directory. Each file's name (without extension)
// is used as the language code.
func LoadJSONDir(bundle *i18n.Bundle, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		lang := strings.TrimSuffix(entry.Name(), ".json")
		if err := LoadJSON(bundle, lang, filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// LoadMap loads messages from a Go map directly into the bundle.
func LoadMap(bundle *i18n.Bundle, messages map[string]map[string]string) {
	for lang, msgs := range messages {
		bundle.AddMessages(lang, msgs)
	}
}
