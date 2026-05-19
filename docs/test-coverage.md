# Test Coverage — digital.vasic.i18n (round-257)

> Verbatim 2026-05-19 operator mandate: *"all existing tests and Challenges do work in anti-bluff manner - they MUST confirm that all tested codebase really works as expected! We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completition and full usability by end users of the product!"*

CONST-050(B) symbol-to-test ledger. Every exported symbol in
`pkg/{i18n,loader,middleware}` is cross-referenced to the unit-test
name(s) that exercise it AND to the round-257 Challenge runner section
that exercises it against real OS facilities (multi-locale in-memory
bundles, `os.TempDir` JSON files, real `net/http` request/response
cycles). No metadata-only PASS — every entry below names the production
code path and the runtime evidence channel that proves it works.

## Anti-bluff posture (round-257)

- **Multi-locale bundle round-trip.** `challenges/runner/main.go`
  Section 1 builds a real `Bundle`, loads 5 locales (en, sr, ja, ar,
  zh-CN) via `loader.LoadMap`, and asserts every `GetMessage(lang,
  "hello")` returns the exact bytes loaded. The PASS line carries the
  locale code AND `utf8.RuneCountInString` so byte preservation is
  observable, not assumed.
- **Real on-disk JSON transport.** Section 2 writes one JSON file per
  locale into `os.MkdirTemp`, exercises `loader.LoadJSON` (single file)
  AND `loader.LoadJSONDir` (directory walk), and asserts byte
  equivalence.
- **Rune-safe template substitution.** Section 3 verifies every
  locale's `"hello"` template renders the supplied non-ASCII `Name`
  placeholder verbatim — proving `{{Name}}` substitution is rune-safe
  for Cyrillic, Japanese, Arabic, and Han scripts.
- **Fallback semantics.** Section 4 plants a key only in `en`, then
  asserts `GetMessage("xx", key)` returns the English value (default
  fallback), and that a truly missing key returns the key verbatim
  (including on an empty bundle).
- **Real HTTP middleware transport.** Section 5 spins up
  `httptest.NewServer` wrapping `middleware.New(DefaultConfig())` and
  fires real `http.Client.Do` requests with `Accept-Language` headers
  (5 locales), `?lang=` query params (5 locales), the
  query-overrides-header invariant, and the default-fallback fall-back.
- **Paired mutation.** Running the gate with `--anti-bluff-mutate`
  plants a deliberate `GetMessage -> GetBogus_MUTATED` rename in a tmp
  copy of this ledger, reruns the cross-reference check, and asserts
  the gate exits 99. Proves the symbol-to-test ledger actually catches
  drift instead of rubber-stamping it.

## pkg/i18n

| Exported symbol | Unit-test coverage | Runner section |
|-----------------|--------------------|----------------|
| `type Bundle` | every `Test*` in `i18n_test.go` | Sections 1, 3, 4 |
| `func NewBundle(defaultLanguage string) *Bundle` | `TestNewBundle_Empty` | Sections 1, 3, 4 (constructed per section) |
| `func (*Bundle) DefaultLanguage() string` | `TestNewBundle_Empty` | Section 1 (asserts returns 'en') |
| `func (*Bundle) AddMessages(lang string, messages map[string]string)` | `TestBundle_AddMessages`, `TestBundle_MultipleLanguages`, `TestBundle_SupportedLanguages`, `TestBundle_TemplateSubstitution`, `TestBundle_MultipleParams` | Sections 3, 4 (planted bundles) |
| `func (*Bundle) GetMessage(lang, key string, params ...map[string]interface{}) string` | `TestBundle_AddMessages`, `TestBundle_FallbackToDefault`, `TestBundle_ReturnKeyIfMissing`, `TestBundle_TemplateSubstitution`, `TestBundle_MultipleLanguages`, `TestBundle_MultipleParams` | Sections 1, 3, 4 (byte-exact, template, fallback, missing-key) |
| `func (*Bundle) SupportedLanguages() []string` | `TestBundle_SupportedLanguages` | Section 1 (count assertion) |

## pkg/loader

| Exported symbol | Unit-test coverage | Runner section |
|-----------------|--------------------|----------------|
| `func LoadJSON(bundle *i18n.Bundle, lang, filePath string) error` | `TestLoadJSON`, `TestLoadJSON_FileNotFound` | Section 2 (real `os.MkdirTemp` + `os.WriteFile` then real `os.ReadFile` path) |
| `func LoadJSONDir(bundle *i18n.Bundle, dir string) error` | `TestLoadJSONDir` | Section 2 (directory walk over 5 locales) |
| `func LoadMap(bundle *i18n.Bundle, messages map[string]map[string]string)` | `TestLoadMap` | Section 1 (bundle hydration entry point) |

## pkg/middleware

| Exported symbol | Unit-test coverage | Runner section |
|-----------------|--------------------|----------------|
| `type Config` | `TestDetectLanguage_AcceptLanguageHeader`, `TestDetectLanguage_QueryParam`, `TestDetectLanguage_FallbackToDefault`, `TestDetectLanguage_QueryOverridesHeader` | Section 5 (DefaultConfig via real server) |
| `func DefaultConfig() *Config` | every `Test*` | Section 5 (used to construct middleware) |
| `func New(cfg *Config) func(http.Handler) http.Handler` | every `Test*` | Section 5 (wraps real `http.HandlerFunc` echoing language) |
| `func LanguageFromContext(ctx context.Context) string` | every `Test*` | Section 5 (handler reads context-injected language) |

## Round-257 artefacts inventory

| Artefact | Path | Purpose |
|----------|------|---------|
| Runner | `challenges/runner/main.go` | Real multi-locale exerciser (5 sections) |
| Mutation gate | `challenges/scripts/i18n_describe_challenge.sh` | Cross-reference + paired-mutation enforcement |
| Multi-locale fixture | `tests/fixtures/i18n/payloads.json` | 5 locales: en, sr, ja, ar, zh-CN |
| README guarantees | `README.md` | Anti-bluff section + quick start |
| Ledger | `docs/test-coverage.md` (this file) | Symbol → test cross-reference |

## Inherited governance challenges (still in scope)

| Script | Purpose |
|--------|---------|
| `challenges/scripts/no_suspend_calls_challenge.sh` | CONST-033 host-power scan |
| `challenges/scripts/host_no_auto_suspend_challenge.sh` | CONST-033 host hardening |
| `challenges/scripts/chaos_failure_injection_challenge.sh` | CONST-050(B) chaos type |
| `challenges/scripts/ddos_health_flood_challenge.sh` | CONST-050(B) DDoS type |
| `challenges/scripts/scaling_horizontal_challenge.sh` | CONST-050(B) scaling type |
| `challenges/scripts/stress_sustained_load_challenge.sh` | CONST-050(B) stress type |
| `challenges/scripts/ui_terminal_interaction_challenge.sh` | CONST-050(B) UI type |
| `challenges/scripts/ux_end_to_end_flow_challenge.sh` | CONST-050(B) UX type |
