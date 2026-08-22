package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir string, name string, contents string) {
	t.Helper()

	filePath := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create fixture directory: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}
}

func serveRequest(handler http.Handler, method string, target string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	for name, value := range headers {
		request.Header.Set(name, value)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestStaticHandlerServesExistingFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.js", "console.log('ok')")

	response := serveRequest(
		newStaticHandler(dir, true, false, 0),
		http.MethodGet,
		"/app.js",
		nil,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if response.Body.String() != "console.log('ok')" {
		t.Fatalf("expected existing file contents, got %q", response.Body.String())
	}
}

func TestStaticHandlerSPAFallbackNavigation(t *testing.T) {
	dir := t.TempDir()
	const index = "<main id=\"app\">servexy</main>"
	writeFile(t, dir, "index.html", index)
	writeFile(t, dir, "docs/index.html", "<main>docs</main>")
	writeFile(t, dir, "empty/.keep", "fixture")

	handler := newStaticHandler(dir, true, false, 0)
	tests := []struct {
		name    string
		target  string
		headers map[string]string
		body    string
	}{
		{
			name:   "HTML route",
			target: "/dashboard/settings",
			headers: map[string]string{
				"Accept": "text/html,application/xhtml+xml",
			},
			body: index,
		},
		{
			name:   "dotted route",
			target: "/users/jane.doe",
			headers: map[string]string{
				"Accept": "text/html",
			},
			body: index,
		},
		{
			name:   "wildcard extensionless route",
			target: "/reports/weekly",
			headers: map[string]string{
				"Accept": "*/*",
			},
			body: index,
		},
		{
			name:   "browser navigation metadata",
			target: "/account/profile.json",
			headers: map[string]string{
				"Accept":         "*/*",
				"Sec-Fetch-Mode": "navigate",
			},
			body: index,
		},
		{
			name:   "physical directory without index",
			target: "/empty/",
			headers: map[string]string{
				"Accept": "text/html",
			},
			body: index,
		},
		{
			name:   "physical directory with index",
			target: "/docs/",
			headers: map[string]string{
				"Accept": "text/html",
			},
			body: "<main>docs</main>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := serveRequest(handler, http.MethodGet, test.target, test.headers)
			if response.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", response.Code)
			}
			if response.Body.String() != test.body {
				t.Fatalf("expected %q, got %q", test.body, response.Body.String())
			}
			if response.Header().Get("Cache-Control") != "no-cache" {
				t.Fatalf("expected no-cache, got %q", response.Header().Get("Cache-Control"))
			}
		})
	}
}

func TestStaticHandlerSPAFallbackDoesNotMaskNonNavigationRequests(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main id=\"app\"></main>")
	handler := newStaticHandler(dir, true, false, 0)

	tests := []struct {
		name   string
		target string
		accept string
		status int
	}{
		{name: "JavaScript asset", target: "/assets/missing.js", accept: "*/*", status: http.StatusNotFound},
		{name: "CSS asset", target: "/assets/missing.css", accept: "text/css,*/*;q=0.1", status: http.StatusNotFound},
		{name: "JSON request", target: "/api/session", accept: "application/json", status: http.StatusNotFound},
		{name: "HTML explicitly rejected", target: "/dashboard", accept: "text/html;q=0", status: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := serveRequest(handler, http.MethodGet, test.target, map[string]string{"Accept": test.accept})
			if response.Code != test.status {
				t.Fatalf("expected %d, got %d", test.status, response.Code)
			}
		})
	}
}

func TestStaticHandlerWithoutSPAFallbackReturns404(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main id=\"app\"></main>")

	response := serveRequest(
		newStaticHandler(dir, false, false, 0),
		http.MethodGet,
		"/dashboard/settings",
		map[string]string{"Accept": "text/html"},
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestStaticHandlerSPAFallbackSupportsHead(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main id=\"app\"></main>")

	response := serveRequest(
		newStaticHandler(dir, true, true, 0),
		http.MethodHead,
		"/dashboard",
		map[string]string{"Accept": "text/html", "Accept-Encoding": "gzip"},
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("expected an empty HEAD response, got %d bytes", response.Body.Len())
	}
	if response.Header().Get("Content-Encoding") != "" {
		t.Fatalf("HEAD requests should not be dynamically compressed")
	}
}

func TestStaticHandlerRejectsUnsupportedMethods(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "ok")

	response := serveRequest(newStaticHandler(dir, false, false, 0), http.MethodPost, "/", nil)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", response.Code)
	}
	if response.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("unexpected Allow header: %q", response.Header().Get("Allow"))
	}
}

func TestStaticHandlerCompressesLargeTextResponses(t *testing.T) {
	dir := t.TempDir()
	contents := strings.Repeat("const message = 'servexy';\n", 200)
	writeFile(t, dir, "assets/app.js", contents)

	response := serveRequest(
		newStaticHandler(dir, false, true, 0),
		http.MethodGet,
		"/assets/app.js",
		map[string]string{"Accept-Encoding": "br, gzip;q=1"},
	)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if response.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding, got %q", response.Header().Get("Content-Encoding"))
	}
	if !strings.Contains(response.Header().Get("Vary"), "Accept-Encoding") {
		t.Fatalf("expected Vary: Accept-Encoding, got %q", response.Header().Get("Vary"))
	}

	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		t.Fatalf("failed to open gzip response: %v", err)
	}
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress response: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("failed to close gzip reader: %v", err)
	}
	if string(decompressed) != contents {
		t.Fatalf("decompressed response did not match source")
	}
}

func TestStaticHandlerRespectsCompressionOptOutAndRanges(t *testing.T) {
	dir := t.TempDir()
	contents := strings.Repeat("servexy", 500)
	writeFile(t, dir, "asset.txt", contents)
	handler := newStaticHandler(dir, false, true, 0)

	t.Run("gzip rejected", func(t *testing.T) {
		response := serveRequest(handler, http.MethodGet, "/asset.txt", map[string]string{"Accept-Encoding": "gzip;q=0, *;q=1"})
		if response.Header().Get("Content-Encoding") != "" {
			t.Fatalf("expected an uncompressed response")
		}
		if response.Body.String() != contents {
			t.Fatalf("expected original response body")
		}
	})

	t.Run("range request", func(t *testing.T) {
		response := serveRequest(handler, http.MethodGet, "/asset.txt", map[string]string{
			"Accept-Encoding": "gzip",
			"Range":           "bytes=0-6",
		})
		if response.Code != http.StatusPartialContent {
			t.Fatalf("expected 206, got %d", response.Code)
		}
		if response.Header().Get("Content-Encoding") != "" {
			t.Fatalf("range responses must not be dynamically compressed")
		}
		if response.Body.String() != "servexy" {
			t.Fatalf("unexpected range body: %q", response.Body.String())
		}
	})
}

func TestStaticHandlerCachePolicy(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main>app</main>")
	writeFile(t, dir, "assets/app.js", "console.log('ok')")
	handler := newStaticHandler(dir, true, false, time.Hour)

	asset := serveRequest(handler, http.MethodGet, "/assets/app.js", nil)
	if asset.Header().Get("Cache-Control") != "public, max-age=3600" {
		t.Fatalf("unexpected asset cache header: %q", asset.Header().Get("Cache-Control"))
	}

	index := serveRequest(handler, http.MethodGet, "/index.html", nil)
	if index.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("unexpected HTML cache header: %q", index.Header().Get("Cache-Control"))
	}

	fallback := serveRequest(handler, http.MethodGet, "/dashboard", map[string]string{"Accept": "text/html"})
	if fallback.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("unexpected fallback cache header: %q", fallback.Header().Get("Cache-Control"))
	}
}

func TestParseOptionsSupportsFlagsAnywhere(t *testing.T) {
	options, err := parseOptions([]string{"./dist", "--spa", "--port=3000", "--cache", "1h", "--network"})
	if err != nil {
		t.Fatalf("parseOptions returned an error: %v", err)
	}

	if options.dir != "./dist" || options.port != "3000" || !options.spa || !options.network {
		t.Fatalf("unexpected options: %+v", options)
	}
	if !options.compress || options.cacheDuration != time.Hour {
		t.Fatalf("unexpected production options: %+v", options)
	}
}

func TestParseOptionsPreservesLegacyAliases(t *testing.T) {
	options, err := parseOptions([]string{"-spa", "-network", "-port", "9000", "-dir", "./public", "--no-compress"})
	if err != nil {
		t.Fatalf("parseOptions returned an error: %v", err)
	}
	if options.dir != "./public" || options.port != "9000" || !options.spa || !options.network || options.compress {
		t.Fatalf("unexpected options: %+v", options)
	}
}

func TestParseOptionsRejectsInvalidInput(t *testing.T) {
	tests := [][]string{
		{"--unknown"},
		{"--port"},
		{"--port", "0"},
		{"--cache", "forever"},
		{"--cache", "500ms"},
		{"--dir", "public", "dist"},
		{"public", "dist"},
		{"--spa=true"},
	}

	for _, args := range tests {
		if _, err := parseOptions(args); err == nil {
			t.Fatalf("expected parsing to fail for %q", args)
		}
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
