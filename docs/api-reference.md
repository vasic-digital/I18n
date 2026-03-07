# API Reference -- digital.vasic.i18n

## Package `i18n`

Import: `digital.vasic.i18n/pkg/i18n`

### Types

#### Bundle

```go
type Bundle struct { /* unexported fields */ }
```

Thread-safe container for localized messages across multiple languages. Uses `sync.RWMutex` internally for concurrent access.

### Functions

#### NewBundle

```go
func NewBundle(defaultLanguage string) *Bundle
```

Creates a new `Bundle` with the specified default (fallback) language. The bundle starts empty; add messages with `AddMessages`.

```go
bundle := i18n.NewBundle("en")
```

### Methods

#### (*Bundle) DefaultLanguage

```go
func (b *Bundle) DefaultLanguage() string
```

Returns the default language code set at construction time.

#### (*Bundle) AddMessages

```go
func (b *Bundle) AddMessages(lang string, messages map[string]string)
```

Adds or merges key-value message pairs for the given language. If messages already exist for that language, new entries are added and existing keys are overwritten.

```go
bundle.AddMessages("en", map[string]string{
    "hello":   "Hello!",
    "goodbye": "Goodbye!",
})
bundle.AddMessages("ru", map[string]string{
    "hello": "Привет!",
})
```

#### (*Bundle) GetMessage

```go
func (b *Bundle) GetMessage(lang, key string, params ...map[string]interface{}) string
```

Retrieves a localized message. Resolution order:

1. Look up `key` in the requested `lang`.
2. If not found, look up `key` in the default language.
3. If still not found, return the raw `key` string.

Template variables in `{{VarName}}` format are substituted from the optional `params` map. Values are converted to strings via `fmt.Sprint`.

```go
msg := bundle.GetMessage("en", "greeting", map[string]interface{}{
    "Name": "Alice",
    "App":  "Bear Messenger",
})
// "Hello, Alice! Welcome to Bear Messenger."
```

#### (*Bundle) SupportedLanguages

```go
func (b *Bundle) SupportedLanguages() []string
```

Returns a slice of all language codes that have messages loaded. Order is not guaranteed.

---

## Package `loader`

Import: `digital.vasic.i18n/pkg/loader`

### Functions

#### LoadJSON

```go
func LoadJSON(bundle *i18n.Bundle, lang, filePath string) error
```

Reads a JSON file containing a flat `{ "key": "value" }` map and loads it into the bundle for the specified language.

```go
err := loader.LoadJSON(bundle, "en", "locales/en.json")
```

**JSON file format:**
```json
{
  "hello": "Hello!",
  "goodbye": "Goodbye!"
}
```

#### LoadJSONDir

```go
func LoadJSONDir(bundle *i18n.Bundle, dir string) error
```

Scans a directory for `.json` files. Each filename (without extension) is used as the language code. Non-JSON files and subdirectories are skipped.

```go
err := loader.LoadJSONDir(bundle, "locales/")
// locales/en.json -> lang "en"
// locales/ru.json -> lang "ru"
```

#### LoadMap

```go
func LoadMap(bundle *i18n.Bundle, messages map[string]map[string]string)
```

Loads messages from a nested Go map (`lang -> key -> message`) directly into the bundle.

```go
loader.LoadMap(bundle, map[string]map[string]string{
    "en": {"hello": "Hello!"},
    "es": {"hello": "Hola!"},
})
```

---

## Package `middleware`

Import: `digital.vasic.i18n/pkg/middleware`

### Types

#### Config

```go
type Config struct {
    DefaultLanguage string // Fallback language (default: "en")
    QueryParam      string // URL query parameter name (default: "lang")
}
```

### Functions

#### DefaultConfig

```go
func DefaultConfig() *Config
```

Returns a `Config` with `DefaultLanguage: "en"` and `QueryParam: "lang"`.

#### New

```go
func New(cfg *Config) func(http.Handler) http.Handler
```

Creates standard `net/http` middleware that detects the client language and stores it in the request context. Detection priority:

1. URL query parameter (e.g., `?lang=ru`)
2. `Accept-Language` header (first language, base code only)
3. `Config.DefaultLanguage`

```go
mux := http.NewServeMux()
mux.Handle("/", middleware.New(middleware.DefaultConfig())(myHandler))
```

#### LanguageFromContext

```go
func LanguageFromContext(ctx context.Context) string
```

Extracts the detected language from the request context. Returns an empty string if the middleware was not applied.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    lang := middleware.LanguageFromContext(r.Context()) // "en", "ru", etc.
    msg := bundle.GetMessage(lang, "hello")
    w.Write([]byte(msg))
}
```
