package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir string, name string, contents string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create fixture directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}
}

func TestStaticHandlerServesExistingFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.js", "console.log('ok')")

	request := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	response := httptest.NewRecorder()

	newStaticHandler(dir, true).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	if !strings.Contains(response.Body.String(), "console.log('ok')") {
		t.Fatalf("expected existing file contents, got %q", response.Body.String())
	}
}

func TestStaticHandlerSPAFallbackServesIndexForHTMLRoute(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main id=\"app\"></main>")

	request := httptest.NewRequest(http.MethodGet, "/dashboard/settings", nil)
	request.Header.Set("Accept", "text/html")
	response := httptest.NewRecorder()

	newStaticHandler(dir, true).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	if !strings.Contains(response.Body.String(), "<main id=\"app\"></main>") {
		t.Fatalf("expected index fallback, got %q", response.Body.String())
	}
}

func TestStaticHandlerSPAFallbackDoesNotMaskMissingAssets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main id=\"app\"></main>")

	request := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)
	request.Header.Set("Accept", "*/*")
	response := httptest.NewRecorder()

	newStaticHandler(dir, true).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestStaticHandlerWithoutSPAFallbackReturns404(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main id=\"app\"></main>")

	request := httptest.NewRequest(http.MethodGet, "/dashboard/settings", nil)
	request.Header.Set("Accept", "text/html")
	response := httptest.NewRecorder()

	newStaticHandler(dir, false).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestValidatePort(t *testing.T) {
	if err := validatePort("8080"); err != nil {
		t.Fatalf("expected valid port: %v", err)
	}

	for _, port := range []string{"0", "65536", "abc"} {
		if err := validatePort(port); err == nil {
			t.Fatalf("expected invalid port %q", port)
		}
	}
}
