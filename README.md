# I18n — digital.vasic.i18n

Generic, reusable internationalization (i18n) library for Go. Three small,
focused packages — `pkg/i18n`, `pkg/loader`, `pkg/middleware` — each fully
decoupled from any consuming project (CONST-051(B)). All exported behaviour
is exercised end-to-end by the round-257 Challenge runner against real
in-memory bundles, real `os.TempDir` JSON files, and real `net/http`
request/response cycles. No mocks beyond unit tests (CONST-050(A)).

## Module facts

| Fact | Value |
|------|-------|
| Module path | `digital.vasic.i18n` |
| Go version | 1.25+ (see `go.mod`) |
| Dependencies | `stretchr/testify` (test-only), `gopkg.in/yaml.v3` (test-only via testify) |
| Test-bank command | `go test ./... -v -race -count=1` |
| Challenge command | `bash challenges/scripts/i18n_describe_challenge.sh` |
| Mutation gate | `bash challenges/scripts/i18n_describe_challenge.sh --anti-bluff-mutate` (exit 99 on PASS) |
| Round | 257 |

## Package overview

### `pkg/i18n` — Bundle + message lookup

The core type is `Bundle`. It owns:

- A `defaultLanguage` set at construction (`NewBundle("en")`).
- A `messages` map of `lang -> key -> message` guarded by `sync.RWMutex`.

Exported surface:

- `NewBundle(defaultLanguage string) *Bundle`
- `(*Bundle).DefaultLanguage() string`
- `(*Bundle).AddMessages(lang string, messages map[string]string)` — additive
  merge per language; safe to call concurrently.
- `(*Bundle).GetMessage(lang, key string, params ...map[string]interface{}) string`
  — returns the localized message; falls back to the default language when
  the key is missing in `lang`; returns the key verbatim when missing in
  both; performs `{{Name}}` template substitution when `params[0]` is
  non-nil.
- `(*Bundle).SupportedLanguages() []string`

### `pkg/loader` — Bundle hydration

Loads messages from JSON files, directories, or Go maps into a `Bundle`:

- `LoadJSON(bundle, lang, filePath) error` — reads a single JSON file
  (`{"key": "value"}` shape) and merges into the bundle for `lang`.
- `LoadJSONDir(bundle, dir) error` — reads every `*.json` in `dir`, using
  the basename (without `.json`) as the language code.
- `LoadMap(bundle, messages)` — loads a nested Go map directly. Useful for
  embedded fixtures and tests.

All three are real `os.ReadFile` / `encoding/json` paths — no in-memory
fakes.

### `pkg/middleware` — HTTP language detection

`net/http`-compatible middleware that detects the request language from
the query parameter (default `?lang=`) or the `Accept-Language` header,
falling back to `Config.DefaultLanguage`. The chosen language is stored
in the request context.

- `type Config struct { DefaultLanguage, QueryParam string }`
- `DefaultConfig() *Config` — returns `{"en", "lang"}`.
- `New(cfg *Config) func(http.Handler) http.Handler` — middleware factory.
- `LanguageFromContext(ctx context.Context) string`

The header parser handles RFC 7231 quality-value lists like
`ru-RU,ru;q=0.9,en;q=0.8` by taking the first entry, stripping the
quality value, and trimming the region suffix.

## Quick start

```go
import (
    "digital.vasic.i18n/pkg/i18n"
    "digital.vasic.i18n/pkg/loader"
)

bundle := i18n.NewBundle("en")
loader.LoadMap(bundle, map[string]map[string]string{
    "en":    {"hello": "Hello, {{Name}}!"},
    "ru":    {"hello": "Привет, {{Name}}!"},
    "ja":    {"hello": "こんにちは、{{Name}}さん!"},
    "ar":    {"hello": "مرحبا، {{Name}}!"},
    "zh-CN": {"hello": "你好, {{Name}}!"},
})

msg := bundle.GetMessage("ja", "hello", map[string]interface{}{"Name": "アリス"})
// "こんにちは、アリスさん!"
```

HTTP middleware:

```go
import (
    "net/http"
    "digital.vasic.i18n/pkg/middleware"
)

mw := middleware.New(middleware.DefaultConfig())
http.Handle("/", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    lang := middleware.LanguageFromContext(r.Context())
    w.Write([]byte("detected: " + lang))
})))
```

## Anti-bluff guarantees (round-257)

The round-257 Challenge runner (`challenges/runner/main.go`) and its
paired-mutation gate (`challenges/scripts/i18n_describe_challenge.sh`)
together enforce six invariants drawn from Article XI §11.9, CONST-035,
and CONST-050(B):

1. **Real bundle round-trip per locale.** Section 1 of the runner builds a
   `Bundle`, loads five locales (en, sr, ja, ar, zh-CN) via
   `loader.LoadMap`, and asserts every `GetMessage(lang, key)` returns the
   exact bytes loaded — `utf8.RuneCountInString` is captured in each PASS
   line. No translation table is hardcoded in the runner; everything is
   loaded from `tests/fixtures/i18n/payloads.json`.
2. **Real on-disk JSON round-trip.** Section 2 writes one JSON file per
   locale into `os.MkdirTemp`, calls `loader.LoadJSON` and
   `loader.LoadJSONDir` on the real bytes, and asserts the messages survive
   the read.
3. **Real template substitution.** Section 3 verifies every locale's
   `"hello"` template renders the supplied non-ASCII `Name` placeholder
   verbatim — proving `{{Name}}` substitution is rune-safe.
4. **Real fallback semantics.** Section 4 plants a key only in `en`, then
   asserts `GetMessage("xx", key)` returns the English value (default
   fallback), and that a truly missing key returns the key verbatim.
5. **Real HTTP middleware transport.** Section 5 spins up
   `httptest.NewServer` wrapping `middleware.New(DefaultConfig())`, fires a
   real `http.Client.Do` request per locale with both
   `Accept-Language` headers and `?lang=` query params, and asserts the
   handler observes the expected language via `LanguageFromContext`. The
   query-overrides-header invariant is asserted explicitly.
6. **Paired mutation.** Running the gate with `--anti-bluff-mutate` plants
   a deliberate symbol-rename in a tmp copy of `docs/test-coverage.md`
   (`GetMessage -> GetBogus_MUTATED`), reruns the cross-reference check,
   and asserts the gate exits 99. Proves the symbol-to-test ledger
   actually catches drift instead of rubber-stamping it.

A Section that returns success without producing the corresponding PASS
line is a §11.9 violation regardless of how green the summary line looks.

## Test bank

```bash
# Unit tests (mocks allowed, per CONST-050(A))
go test ./... -v -race -count=1

# Challenge: deep-doc + runner gate (clean mode)
bash challenges/scripts/i18n_describe_challenge.sh

# Paired-mutation gate (must exit 99 on PASS)
bash challenges/scripts/i18n_describe_challenge.sh --anti-bluff-mutate

# Inherited governance challenges
bash challenges/scripts/no_suspend_calls_challenge.sh
bash challenges/scripts/host_no_auto_suspend_challenge.sh
```

## Governance

This submodule inherits the constitution submodule's universal rules.
See `CLAUDE.md`, `AGENTS.md`, `CONSTITUTION.md` for the cascaded clauses
(CONST-033, CONST-035, CONST-042, CONST-043, CONST-047..061).

## License

Private — vasic-digital.
