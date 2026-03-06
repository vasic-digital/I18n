package loader_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"digital.vasic.i18n/pkg/i18n"
	"digital.vasic.i18n/pkg/loader"
)

func TestLoadJSON(t *testing.T) {
	dir := t.TempDir()
	enFile := filepath.Join(dir, "en.json")
	data, _ := json.Marshal(map[string]string{
		"hello": "Hello!",
		"bye":   "Goodbye!",
	})
	os.WriteFile(enFile, data, 0644)

	bundle := i18n.NewBundle("en")
	err := loader.LoadJSON(bundle, "en", enFile)
	require.NoError(t, err)

	assert.Equal(t, "Hello!", bundle.GetMessage("en", "hello"))
	assert.Equal(t, "Goodbye!", bundle.GetMessage("en", "bye"))
}

func TestLoadJSONDir(t *testing.T) {
	dir := t.TempDir()
	for lang, msgs := range map[string]map[string]string{
		"en": {"hello": "Hello!"},
		"ru": {"hello": "Привет!"},
	} {
		data, _ := json.Marshal(msgs)
		os.WriteFile(filepath.Join(dir, lang+".json"), data, 0644)
	}

	bundle := i18n.NewBundle("en")
	err := loader.LoadJSONDir(bundle, dir)
	require.NoError(t, err)

	assert.Equal(t, "Hello!", bundle.GetMessage("en", "hello"))
	assert.Equal(t, "Привет!", bundle.GetMessage("ru", "hello"))
}

func TestLoadJSON_FileNotFound(t *testing.T) {
	bundle := i18n.NewBundle("en")
	err := loader.LoadJSON(bundle, "en", "/nonexistent.json")
	assert.Error(t, err)
}

func TestLoadMap(t *testing.T) {
	bundle := i18n.NewBundle("en")
	loader.LoadMap(bundle, map[string]map[string]string{
		"en": {"hello": "Hello!"},
		"es": {"hello": "¡Hola!"},
	})

	assert.Equal(t, "Hello!", bundle.GetMessage("en", "hello"))
	assert.Equal(t, "¡Hola!", bundle.GetMessage("es", "hello"))
}
