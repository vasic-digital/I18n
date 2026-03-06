package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"digital.vasic.i18n/pkg/i18n"
)

func TestNewBundle_Empty(t *testing.T) {
	b := i18n.NewBundle("en")
	assert.NotNil(t, b)
	assert.Equal(t, "en", b.DefaultLanguage())
}

func TestBundle_AddMessages(t *testing.T) {
	b := i18n.NewBundle("en")
	b.AddMessages("en", map[string]string{
		"hello": "Hello!",
		"bye":   "Goodbye!",
	})
	assert.Equal(t, "Hello!", b.GetMessage("en", "hello"))
}

func TestBundle_FallbackToDefault(t *testing.T) {
	b := i18n.NewBundle("en")
	b.AddMessages("en", map[string]string{"hello": "Hello!"})
	assert.Equal(t, "Hello!", b.GetMessage("fr", "hello"))
}

func TestBundle_ReturnKeyIfMissing(t *testing.T) {
	b := i18n.NewBundle("en")
	assert.Equal(t, "nonexistent", b.GetMessage("en", "nonexistent"))
}

func TestBundle_TemplateSubstitution(t *testing.T) {
	b := i18n.NewBundle("en")
	b.AddMessages("en", map[string]string{
		"greeting": "Hello, {{Name}}! Welcome to {{App}}.",
	})
	msg := b.GetMessage("en", "greeting", map[string]interface{}{
		"Name": "Alice",
		"App":  "Bear Messenger",
	})
	assert.Equal(t, "Hello, Alice! Welcome to Bear Messenger.", msg)
}

func TestBundle_MultipleLanguages(t *testing.T) {
	b := i18n.NewBundle("en")
	b.AddMessages("en", map[string]string{"hello": "Hello!"})
	b.AddMessages("ru", map[string]string{"hello": "Привет!"})
	b.AddMessages("es", map[string]string{"hello": "¡Hola!"})

	assert.Equal(t, "Hello!", b.GetMessage("en", "hello"))
	assert.Equal(t, "Привет!", b.GetMessage("ru", "hello"))
	assert.Equal(t, "¡Hola!", b.GetMessage("es", "hello"))
}

func TestBundle_SupportedLanguages(t *testing.T) {
	b := i18n.NewBundle("en")
	b.AddMessages("en", map[string]string{"hello": "Hello!"})
	b.AddMessages("ru", map[string]string{"hello": "Привет!"})

	langs := b.SupportedLanguages()
	assert.Contains(t, langs, "en")
	assert.Contains(t, langs, "ru")
}

func TestBundle_MultipleParams(t *testing.T) {
	b := i18n.NewBundle("en")
	b.AddMessages("en", map[string]string{
		"invite": "{{Inviter}} invited you to {{App}}. Code: {{Code}}",
	})
	msg := b.GetMessage("en", "invite", map[string]interface{}{
		"Inviter": "Bob",
		"App":     "Bear",
		"Code":    12345,
	})
	assert.Equal(t, "Bob invited you to Bear. Code: 12345", msg)
}
