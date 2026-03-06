// Package i18n provides internationalization support with message bundles,
// template substitution, and language fallback.
package i18n

import (
	"fmt"
	"strings"
	"sync"
)

// Bundle holds localized messages for multiple languages.
type Bundle struct {
	mu              sync.RWMutex
	defaultLanguage string
	messages        map[string]map[string]string // lang -> key -> message
}

// NewBundle creates a new message Bundle with the given default language.
func NewBundle(defaultLanguage string) *Bundle {
	return &Bundle{
		defaultLanguage: defaultLanguage,
		messages:        make(map[string]map[string]string),
	}
}

// DefaultLanguage returns the bundle's default (fallback) language.
func (b *Bundle) DefaultLanguage() string {
	return b.defaultLanguage
}

// AddMessages adds or merges messages for the given language.
func (b *Bundle) AddMessages(lang string, messages map[string]string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.messages[lang] == nil {
		b.messages[lang] = make(map[string]string)
	}
	for k, v := range messages {
		b.messages[lang][k] = v
	}
}

// GetMessage retrieves a localized message by language and key.
// Template variables in {{VarName}} format are replaced using the optional params map.
// Falls back to the default language if the key is not found in the requested language.
// Returns the key itself if the message is not found in any language.
func (b *Bundle) GetMessage(lang, key string, params ...map[string]interface{}) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msg := b.lookup(lang, key)
	if msg == "" {
		msg = b.lookup(b.defaultLanguage, key)
	}
	if msg == "" {
		return key
	}

	if len(params) > 0 && params[0] != nil {
		msg = substituteParams(msg, params[0])
	}
	return msg
}

// SupportedLanguages returns a list of languages that have messages loaded.
func (b *Bundle) SupportedLanguages() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	langs := make([]string, 0, len(b.messages))
	for lang := range b.messages {
		langs = append(langs, lang)
	}
	return langs
}

func (b *Bundle) lookup(lang, key string) string {
	langMessages := b.messages[lang]
	if langMessages == nil {
		return ""
	}
	return langMessages[key]
}

func substituteParams(msg string, params map[string]interface{}) string {
	for k, v := range params {
		placeholder := "{{" + k + "}}"
		msg = strings.ReplaceAll(msg, placeholder, fmt.Sprint(v))
	}
	return msg
}
