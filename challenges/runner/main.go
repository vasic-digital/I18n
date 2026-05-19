// Round-257 challenge runner for digital.vasic.i18n.
//
// Drives every public surface of pkg/{i18n,loader,middleware} through
// real in-memory bundles, real os.TempDir JSON files, and real
// net/http request/response cycles. The runner reads its bilingual
// inputs from tests/fixtures/i18n/payloads.json (5 locales:
// en, sr, ja, ar, zh-CN) — no translation table is hardcoded here.
//
// Sections:
//
//  1. pkg/i18n + pkg/loader (LoadMap):   bundle round-trip per locale.
//     Asserts GetMessage returns the exact bytes loaded; captures
//     utf8.RuneCountInString in the PASS line.
//  2. pkg/loader (LoadJSON + LoadJSONDir): real on-disk transport.
//     Writes one JSON file per locale into os.MkdirTemp, calls
//     LoadJSON for one file and LoadJSONDir for the directory, and
//     asserts the messages survive the byte round-trip.
//  3. pkg/i18n (template substitution):   rune-safe {{Name}} render
//     against the same 5 locales' "hello" templates.
//  4. pkg/i18n (fallback + missing key):  plants a key only in en,
//     asserts fallback returns en value, and missing key returns
//     key verbatim.
//  5. pkg/middleware (real HTTP):         spins up httptest.NewServer
//     wrapping middleware.New(DefaultConfig()), fires real
//     http.Client.Do per locale with Accept-Language headers, query
//     params, and the query-overrides-header invariant.
//
// Anti-bluff invariants enforced (Article XI §11.9 + CONST-035 + CONST-050(B)):
//
//   - No metadata-only / grep-only PASS. Every PASS line is preceded by
//     the locale code, the package exercised, and the actual rune count
//     of the round-tripped string (proves bytes survived, not just that
//     no error was returned).
//   - Real os.MkdirTemp + real os.WriteFile + real os.ReadFile (loader).
//   - Real net/http transport via httptest.NewServer + http.Client.
//   - Failure to round-trip non-ASCII bytes, failure to fall back to
//     default language, or middleware ignoring query-overrides-header
//     is a hard FAIL — exit non-zero.
//   - No mocks injected into the library; no patched HTTP client; no
//     stubs. The runner uses each package's public surface exactly as
//     a downstream consumer would.
//
// Verbatim 2026-05-19 operator mandate: "all existing tests and Challenges
// do work in anti-bluff manner - they MUST confirm that all tested codebase
// really works as expected! We had been in position that all tests do execute
// with success and all Challenges as well, but in reality the most of the
// features does not work and can't be used! This MUST NOT be the case and
// execution of tests and Challenges MUST guarantee the quality, the
// completition and full usability by end users of the product!"
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"unicode/utf8"

	"digital.vasic.i18n/pkg/i18n"
	"digital.vasic.i18n/pkg/loader"
	"digital.vasic.i18n/pkg/middleware"
)

type fixtureInput struct {
	Locale           string `json:"locale"`
	Hello            string `json:"hello"`
	Name             string `json:"name"`
	ExpectedRendered string `json:"expected_rendered"`
	ExpectedMinRunes int    `json:"expected_min_runes"`
}

type fixtureFile struct {
	Inputs []fixtureInput `json:"inputs"`
}

var (
	passCount int
	failCount int
)

func pass(format string, args ...interface{}) {
	passCount++
	fmt.Printf("  PASS: "+format+"\n", args...)
}

func fail(format string, args ...interface{}) {
	failCount++
	fmt.Printf("  FAIL: "+format+"\n", args...)
}

func main() {
	fixturesPath := flag.String("fixtures", "tests/fixtures/i18n/payloads.json", "path to bilingual fixture JSON")
	flag.Parse()

	fmt.Printf("=== Round-257 i18n Challenge Runner ===\n")
	fmt.Printf("Fixture: %s\n", *fixturesPath)
	fmt.Println()

	raw, err := os.ReadFile(*fixturesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read fixture %s: %v\n", *fixturesPath, err)
		os.Exit(2)
	}
	var fx fixtureFile
	if err := json.Unmarshal(raw, &fx); err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse fixture: %v\n", err)
		os.Exit(2)
	}
	if len(fx.Inputs) < 3 {
		fmt.Fprintf(os.Stderr, "fixture has only %d inputs; need >=3\n", len(fx.Inputs))
		os.Exit(2)
	}

	section1Bundle(fx)
	section2OnDisk(fx)
	section3Template(fx)
	section4Fallback()
	section5Middleware(fx)

	fmt.Println()
	fmt.Printf("=== Summary: %d PASS, %d FAIL ===\n", passCount, failCount)
	if failCount > 0 {
		os.Exit(1)
	}
}

// -----------------------------------------------------------------------------
// Section 1 — pkg/i18n + pkg/loader.LoadMap: bundle round-trip per locale.
// -----------------------------------------------------------------------------

func section1Bundle(fx fixtureFile) {
	fmt.Println("Section 1: pkg/i18n Bundle round-trip via loader.LoadMap")

	bundle := i18n.NewBundle("en")
	payload := map[string]map[string]string{}
	for _, in := range fx.Inputs {
		payload[in.Locale] = map[string]string{"hello": in.Hello}
	}
	loader.LoadMap(bundle, payload)

	// SupportedLanguages must reflect everything we loaded.
	got := bundle.SupportedLanguages()
	if len(got) >= len(fx.Inputs) {
		pass("[i18n][SupportedLanguages] %d languages registered", len(got))
	} else {
		fail("[i18n][SupportedLanguages] got %d, expected >=%d", len(got), len(fx.Inputs))
	}
	if bundle.DefaultLanguage() == "en" {
		pass("[i18n][DefaultLanguage] returns 'en'")
	} else {
		fail("[i18n][DefaultLanguage] got %q, expected 'en'", bundle.DefaultLanguage())
	}

	for _, in := range fx.Inputs {
		msg := bundle.GetMessage(in.Locale, "hello")
		runes := utf8.RuneCountInString(msg)
		if msg == in.Hello {
			pass("[i18n][%s] GetMessage byte-exact (%d runes)", in.Locale, runes)
		} else {
			fail("[i18n][%s] GetMessage got %q, expected %q", in.Locale, msg, in.Hello)
		}
		if runes >= in.ExpectedMinRunes {
			pass("[i18n][%s] rune count %d >= expected_min %d", in.Locale, runes, in.ExpectedMinRunes)
		} else {
			fail("[i18n][%s] rune count %d < expected_min %d", in.Locale, runes, in.ExpectedMinRunes)
		}
	}
}

// -----------------------------------------------------------------------------
// Section 2 — pkg/loader.LoadJSON + LoadJSONDir: real on-disk transport.
// -----------------------------------------------------------------------------

func section2OnDisk(fx fixtureFile) {
	fmt.Println()
	fmt.Println("Section 2: pkg/loader on-disk JSON (real os.MkdirTemp + os.WriteFile)")

	dir, err := os.MkdirTemp("", "i18n-round257-*")
	if err != nil {
		fail("[loader][mktempdir] %v", err)
		return
	}
	defer os.RemoveAll(dir)

	// Write one JSON file per locale.
	for _, in := range fx.Inputs {
		path := filepath.Join(dir, in.Locale+".json")
		data, err := json.Marshal(map[string]string{"hello": in.Hello})
		if err != nil {
			fail("[loader][%s] marshal: %v", in.Locale, err)
			continue
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fail("[loader][%s] write: %v", in.Locale, err)
			continue
		}
	}

	// LoadJSON for the first locale only — proves single-file path.
	firstBundle := i18n.NewBundle("en")
	first := fx.Inputs[0]
	if err := loader.LoadJSON(firstBundle, first.Locale, filepath.Join(dir, first.Locale+".json")); err != nil {
		fail("[loader][LoadJSON][%s] %v", first.Locale, err)
	} else {
		got := firstBundle.GetMessage(first.Locale, "hello")
		if got == first.Hello {
			pass("[loader][LoadJSON][%s] byte-exact (%d runes)", first.Locale, utf8.RuneCountInString(got))
		} else {
			fail("[loader][LoadJSON][%s] got %q, expected %q", first.Locale, got, first.Hello)
		}
	}

	// LoadJSONDir for the whole tmp dir.
	dirBundle := i18n.NewBundle("en")
	if err := loader.LoadJSONDir(dirBundle, dir); err != nil {
		fail("[loader][LoadJSONDir] %v", err)
		return
	}
	for _, in := range fx.Inputs {
		got := dirBundle.GetMessage(in.Locale, "hello")
		if got == in.Hello {
			pass("[loader][LoadJSONDir][%s] byte-exact (%d runes)", in.Locale, utf8.RuneCountInString(got))
		} else {
			fail("[loader][LoadJSONDir][%s] got %q, expected %q", in.Locale, got, in.Hello)
		}
	}
}

// -----------------------------------------------------------------------------
// Section 3 — pkg/i18n template substitution: rune-safe {{Name}} render.
// -----------------------------------------------------------------------------

func section3Template(fx fixtureFile) {
	fmt.Println()
	fmt.Println("Section 3: pkg/i18n template substitution (rune-safe {{Name}})")

	bundle := i18n.NewBundle("en")
	for _, in := range fx.Inputs {
		bundle.AddMessages(in.Locale, map[string]string{"hello": in.Hello})
	}

	for _, in := range fx.Inputs {
		rendered := bundle.GetMessage(in.Locale, "hello", map[string]interface{}{"Name": in.Name})
		if rendered == in.ExpectedRendered {
			pass("[template][%s] %q -> %d runes", in.Locale, rendered, utf8.RuneCountInString(rendered))
		} else {
			fail("[template][%s] got %q, expected %q", in.Locale, rendered, in.ExpectedRendered)
		}
	}
}

// -----------------------------------------------------------------------------
// Section 4 — pkg/i18n fallback + missing-key semantics.
// -----------------------------------------------------------------------------

func section4Fallback() {
	fmt.Println()
	fmt.Println("Section 4: pkg/i18n fallback + missing-key")

	bundle := i18n.NewBundle("en")
	bundle.AddMessages("en", map[string]string{"only_en": "english-only"})

	// Fallback: request non-existent locale, expect default-lang value.
	got := bundle.GetMessage("xx", "only_en")
	if got == "english-only" {
		pass("[fallback] non-existent locale falls back to default")
	} else {
		fail("[fallback] got %q, expected 'english-only'", got)
	}

	// Missing key: return key verbatim.
	missing := bundle.GetMessage("en", "totally.absent.key")
	if missing == "totally.absent.key" {
		pass("[missing-key] returns key verbatim")
	} else {
		fail("[missing-key] got %q, expected 'totally.absent.key'", missing)
	}

	// Empty bundle: still returns key.
	empty := i18n.NewBundle("en")
	if empty.GetMessage("xx", "anything") == "anything" {
		pass("[missing-key][empty-bundle] returns key verbatim")
	} else {
		fail("[missing-key][empty-bundle] did not return key verbatim")
	}
}

// -----------------------------------------------------------------------------
// Section 5 — pkg/middleware: real HTTP transport via httptest.NewServer.
// -----------------------------------------------------------------------------

func section5Middleware(fx fixtureFile) {
	fmt.Println()
	fmt.Println("Section 5: pkg/middleware real HTTP (httptest.NewServer + http.Client)")

	mw := middleware.New(middleware.DefaultConfig())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := middleware.LanguageFromContext(r.Context())
		_, _ = io.WriteString(w, lang)
	}))
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := ts.Client()

	// 5a: each locale via Accept-Language header. Strip region suffix.
	for _, in := range fx.Inputs {
		req, err := http.NewRequest("GET", ts.URL+"/", nil)
		if err != nil {
			fail("[middleware][%s] new request: %v", in.Locale, err)
			continue
		}
		req.Header.Set("Accept-Language", in.Locale+",en;q=0.5")
		resp, err := client.Do(req)
		if err != nil {
			fail("[middleware][%s] do: %v", in.Locale, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		// Header parser strips region suffix (zh-CN -> zh).
		expected := stripRegion(in.Locale)
		if string(body) == expected {
			pass("[middleware][%s][Accept-Language] -> %q (%d bytes)", in.Locale, expected, len(body))
		} else {
			fail("[middleware][%s][Accept-Language] got %q, expected %q", in.Locale, string(body), expected)
		}
	}

	// 5b: ?lang= query param verbatim (no region stripping for query).
	for _, in := range fx.Inputs {
		req, err := http.NewRequest("GET", ts.URL+"/?lang="+in.Locale, nil)
		if err != nil {
			fail("[middleware][%s][query] new request: %v", in.Locale, err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			fail("[middleware][%s][query] do: %v", in.Locale, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if string(body) == in.Locale {
			pass("[middleware][%s][query] -> %q verbatim", in.Locale, in.Locale)
		} else {
			fail("[middleware][%s][query] got %q, expected %q", in.Locale, string(body), in.Locale)
		}
	}

	// 5c: query-overrides-header invariant.
	req, _ := http.NewRequest("GET", ts.URL+"/?lang=zh", nil)
	req.Header.Set("Accept-Language", "ru")
	resp, err := client.Do(req)
	if err != nil {
		fail("[middleware][query-overrides-header] do: %v", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) == "zh" {
		pass("[middleware][query-overrides-header] query 'zh' beat header 'ru'")
	} else {
		fail("[middleware][query-overrides-header] got %q, expected 'zh'", string(body))
	}

	// 5d: fallback to DefaultLanguage when neither query nor header given.
	resp2, err := client.Get(ts.URL + "/")
	if err != nil {
		fail("[middleware][default-fallback] do: %v", err)
		return
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if string(body2) == "en" {
		pass("[middleware][default-fallback] returned 'en'")
	} else {
		fail("[middleware][default-fallback] got %q, expected 'en'", string(body2))
	}
}

func stripRegion(locale string) string {
	for i := 0; i < len(locale); i++ {
		if locale[i] == '-' {
			return locale[:i]
		}
	}
	return locale
}
