# User Guide -- digital.vasic.i18n

## Installation

```bash
go get digital.vasic.i18n
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "fmt"

    "digital.vasic.i18n/pkg/i18n"
    "digital.vasic.i18n/pkg/loader"
)

func main() {
    bundle := i18n.NewBundle("en")

    // Load from Go maps
    loader.LoadMap(bundle, map[string]map[string]string{
        "en": {
            "hello":    "Hello!",
            "greeting": "Hello, {{Name}}!",
        },
        "ru": {
            "hello":    "Привет!",
            "greeting": "Привет, {{Name}}!",
        },
    })

    fmt.Println(bundle.GetMessage("en", "hello"))
    // Output: Hello!

    fmt.Println(bundle.GetMessage("ru", "greeting", map[string]interface{}{
        "Name": "Alice",
    }))
    // Output: Привет, Alice!
}
```

## Loading Messages from JSON Files

Create one JSON file per language:

**locales/en.json**
```json
{
  "hello": "Hello!",
  "goodbye": "Goodbye!",
  "welcome": "Welcome to {{App}}, {{Name}}!"
}
```

**locales/ru.json**
```json
{
  "hello": "Привет!",
  "goodbye": "До свидания!",
  "welcome": "Добро пожаловать в {{App}}, {{Name}}!"
}
```

Load the entire directory at once:

```go
bundle := i18n.NewBundle("en")
if err := loader.LoadJSONDir(bundle, "locales/"); err != nil {
    log.Fatal(err)
}
```

Or load a single file:

```go
err := loader.LoadJSON(bundle, "es", "locales/es.json")
```

## Template Substitution

Messages can contain `{{VarName}}` placeholders. Pass a `map[string]interface{}` as the last argument to `GetMessage`:

```go
msg := bundle.GetMessage("en", "welcome", map[string]interface{}{
    "App":  "Bear Messenger",
    "Name": "Bob",
})
// "Welcome to Bear Messenger, Bob!"
```

Any type is accepted as a value; it is converted via `fmt.Sprint`. If no params map is provided, placeholders remain as-is in the output.

## Language Fallback

`GetMessage` follows a two-step fallback:

1. Look in the requested language.
2. If not found, look in the bundle's default language.
3. If still not found, return the key string itself.

```go
bundle := i18n.NewBundle("en")
bundle.AddMessages("en", map[string]string{"hello": "Hello!"})

bundle.GetMessage("fr", "hello")      // "Hello!" (falls back to "en")
bundle.GetMessage("fr", "unknown_key") // "unknown_key" (returns the key)
```

## HTTP Middleware

The `middleware` package detects the client's preferred language and stores it in the request context.

### With net/http

```go
import "digital.vasic.i18n/pkg/middleware"

cfg := middleware.DefaultConfig() // DefaultLanguage: "en", QueryParam: "lang"
mw := middleware.New(cfg)

http.Handle("/", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    lang := middleware.LanguageFromContext(r.Context())
    msg := bundle.GetMessage(lang, "hello")
    w.Write([]byte(msg))
})))
```

### With Gin

Wrap the standard middleware for use with Gin:

```go
import (
    "github.com/gin-gonic/gin"
    "digital.vasic.i18n/pkg/middleware"
)

func I18nMiddleware() gin.HandlerFunc {
    mw := middleware.New(middleware.DefaultConfig())
    return func(c *gin.Context) {
        mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            lang := middleware.LanguageFromContext(r.Context())
            c.Set("lang", lang)
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    }
}
```

### Language Detection Priority

1. **Query parameter** -- `?lang=ru` (highest priority)
2. **Accept-Language header** -- parses the first language code (e.g., `ru` from `ru-RU,ru;q=0.9`)
3. **Default language** -- configured in `Config.DefaultLanguage`

### Custom Configuration

```go
cfg := &middleware.Config{
    DefaultLanguage: "ru",      // Default to Russian
    QueryParam:      "locale",  // Use ?locale= instead of ?lang=
}
mw := middleware.New(cfg)
```

## Listing Supported Languages

```go
langs := bundle.SupportedLanguages()
// e.g., ["en", "ru", "es"]
```

The returned slice contains all languages that have at least one message loaded. Order is not guaranteed.

## Thread Safety

`Bundle` is safe for concurrent use. Multiple goroutines can call `GetMessage` simultaneously, and `AddMessages` can be called at any time (it acquires a write lock). This makes it safe to use as a long-lived singleton shared across HTTP handlers.

## Testing

```bash
go test ./... -v -count=1
go test -race ./... -count=1
```
