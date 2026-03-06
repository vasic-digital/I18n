// Package middleware provides HTTP middleware for language detection
// from Accept-Language headers and query parameters.
package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const languageKey contextKey = "i18n_language"

// Config holds language detection middleware configuration.
type Config struct {
	DefaultLanguage string
	QueryParam      string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultLanguage: "en",
		QueryParam:      "lang",
	}
}

// New creates language detection middleware.
func New(cfg *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := detectLanguage(r, cfg)
			ctx := context.WithValue(r.Context(), languageKey, lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LanguageFromContext extracts the detected language from the request context.
func LanguageFromContext(ctx context.Context) string {
	lang, ok := ctx.Value(languageKey).(string)
	if !ok {
		return ""
	}
	return lang
}

func detectLanguage(r *http.Request, cfg *Config) string {
	if lang := r.URL.Query().Get(cfg.QueryParam); lang != "" {
		return lang
	}
	if accept := r.Header.Get("Accept-Language"); accept != "" {
		lang := parseAcceptLanguage(accept)
		if lang != "" {
			return lang
		}
	}
	return cfg.DefaultLanguage
}

func parseAcceptLanguage(header string) string {
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return ""
	}
	lang := strings.TrimSpace(parts[0])
	if idx := strings.Index(lang, ";"); idx != -1 {
		lang = lang[:idx]
	}
	if idx := strings.Index(lang, "-"); idx != -1 {
		lang = lang[:idx]
	}
	return strings.TrimSpace(lang)
}
