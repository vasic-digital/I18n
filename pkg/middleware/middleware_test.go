package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"digital.vasic.i18n/pkg/middleware"
)

func TestDetectLanguage_AcceptLanguageHeader(t *testing.T) {
	mw := middleware.New(middleware.DefaultConfig())
	var detectedLang string

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detectedLang = middleware.LanguageFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.8")
	handler.ServeHTTP(w, req)

	assert.Equal(t, "ru", detectedLang)
}

func TestDetectLanguage_QueryParam(t *testing.T) {
	mw := middleware.New(middleware.DefaultConfig())
	var detectedLang string

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detectedLang = middleware.LanguageFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?lang=es", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, "es", detectedLang)
}

func TestDetectLanguage_FallbackToDefault(t *testing.T) {
	mw := middleware.New(middleware.DefaultConfig())
	var detectedLang string

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detectedLang = middleware.LanguageFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, "en", detectedLang)
}

func TestDetectLanguage_QueryOverridesHeader(t *testing.T) {
	mw := middleware.New(middleware.DefaultConfig())
	var detectedLang string

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detectedLang = middleware.LanguageFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?lang=zh", nil)
	req.Header.Set("Accept-Language", "ru")
	handler.ServeHTTP(w, req)

	assert.Equal(t, "zh", detectedLang)
}
