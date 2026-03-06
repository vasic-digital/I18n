# I18n

Generic internationalization module for Go applications.

## Features
- Message bundles with multi-language support
- Template variable substitution (`{{Name}}`)
- Language fallback chain
- JSON file and directory loading
- HTTP middleware for Accept-Language detection

## Quick Start

```go
import (
    "digital.vasic.i18n/pkg/i18n"
    "digital.vasic.i18n/pkg/loader"
)

bundle := i18n.NewBundle("en")
loader.LoadMap(bundle, map[string]map[string]string{
    "en": {"hello": "Hello, {{Name}}!"},
    "ru": {"hello": "Привет, {{Name}}!"},
})

msg := bundle.GetMessage("ru", "hello", map[string]interface{}{"Name": "Alice"})
// "Привет, Alice!"
```

## License
Private - vasic-digital
