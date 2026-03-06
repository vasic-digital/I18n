# CLAUDE.md - I18n Module

## Overview
Generic internationalization (i18n) Go module providing message bundles, template substitution, language detection, and file-based message loading.

**Module**: `digital.vasic.i18n` (Go 1.24+)

## Build & Test
```bash
go test ./... -v -count=1
go test -race ./... -count=1
```

## Package Structure
| Package | Purpose |
|---------|---------|
| `pkg/i18n` | Core: Bundle, GetMessage with {{placeholder}} substitution, language fallback |
| `pkg/loader` | Load messages from JSON files, directories, or Go maps |
| `pkg/middleware` | HTTP middleware for Accept-Language detection and query param override |

## Key Patterns
- Bundle is thread-safe (sync.RWMutex)
- GetMessage falls back: requested lang -> default lang -> return key
- Template substitution uses `{{VarName}}` placeholders
- Middleware stores language in context via `context.WithValue`
