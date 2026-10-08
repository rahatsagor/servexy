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

func newTestHandler(t *testing.T, dir string, options handlerOptions) http.Handler {
	t.Helper()

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("failed to open root: %v", err)
	}
	t.Cleanup(func() { root.Close() })
	return newStaticHandler(root, options)
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
		newTestHandler(t, dir, handlerOptions{spa: true}),
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

	handler := newTestHandler(t, dir, handlerOptions{spa: true})
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
	handler := newTestHandler(t, dir, handlerOptions{spa: true})

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
		newTestHandler(t, dir, handlerOptions{}),
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
		newTestHandler(t, dir, handlerOptions{spa: true, compress: true}),
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

	response := serveRequest(newTestHandler(t, dir, handlerOptions{}), http.MethodPost, "/", nil)
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
		newTestHandler(t, dir, handlerOptions{compress: true}),
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
	handler := newTestHandler(t, dir, handlerOptions{compress: true})

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
	handler := newTestHandler(t, dir, handlerOptions{spa: true, cacheDuration: time.Hour})

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

	if options.dir != "./dist" || options.port != "3000" || !options.handler.spa || !options.network {
		t.Fatalf("unexpected options: %+v", options)
	}
	if !options.handler.compress || options.handler.cacheDuration != time.Hour || options.handler.dotfiles {
		t.Fatalf("unexpected production options: %+v", options)
	}
}

func TestParseOptionsPreservesLegacyAliases(t *testing.T) {
	options, err := parseOptions([]string{"-spa", "-network", "-port", "9000", "-dir", "./public", "--no-compress"})
	if err != nil {
		t.Fatalf("parseOptions returned an error: %v", err)
	}
	if options.dir != "./public" || options.port != "9000" || !options.handler.spa || !options.network || options.handler.compress {
		t.Fatalf("unexpected options: %+v", options)
	}
}

func TestParseOptionsDotfilesAndPortNormalization(t *testing.T) {
	options, err := parseOptions([]string{"--dotfiles", "--port", "+08080"})
	if err != nil {
		t.Fatalf("parseOptions returned an error: %v", err)
	}
	if !options.handler.dotfiles {
		t.Fatalf("expected dotfiles to be enabled")
	}
	if options.port != "8080" {
		t.Fatalf("expected normalized port 8080, got %q", options.port)
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
		{"--dotfiles=1"},
	}

	for _, args := range tests {
		if _, err := parseOptions(args); err == nil {
			t.Fatalf("expected parsing to fail for %q", args)
		}
	}
}

func TestParsePort(t *testing.T) {
	if port, err := parsePort("8080"); err != nil || port != 8080 {
		t.Fatalf("expected valid port 8080, got %d (%v)", port, err)
	}

	for _, port := range []string{"0", "65536", "abc", "-1", ""} {
		if _, err := parsePort(port); err == nil {
			t.Fatalf("expected invalid port %q", port)
		}
	}
}

func TestStaticHandlerHidesDotfiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main>app</main>")
	writeFile(t, dir, ".env", "SECRET=1")
	writeFile(t, dir, ".git/config", "[core]")
	writeFile(t, dir, "public/.htpasswd", "user:hash")
	writeFile(t, dir, "public/visible.txt", "visible")
	writeFile(t, dir, ".well-known/security.txt", "Contact: mailto:security@example.com")
	writeFile(t, dir, ".well-known/.secret", "SECRET=2")

	handler := newTestHandler(t, dir, handlerOptions{})
	targets := []string{"/.env", "/%2eenv", "/.git/config", "/.git/", "/public/.htpasswd", "/.well-known/.secret"}
	for _, target := range targets {
		response := serveRequest(handler, http.MethodGet, target, nil)
		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for %s, got %d", target, response.Code)
		}
	}

	wellKnown := serveRequest(handler, http.MethodGet, "/.well-known/security.txt", nil)
	if wellKnown.Code != http.StatusOK {
		t.Fatalf("expected .well-known to be served, got %d", wellKnown.Code)
	}

	listing := serveRequest(handler, http.MethodGet, "/public/", nil)
	if listing.Code != http.StatusOK {
		t.Fatalf("expected directory listing, got %d", listing.Code)
	}
	if strings.Contains(listing.Body.String(), ".htpasswd") || !strings.Contains(listing.Body.String(), "visible.txt") {
		t.Fatalf("directory listing should hide dotfiles: %q", listing.Body.String())
	}

	spa := newTestHandler(t, dir, handlerOptions{spa: true})
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		for _, target := range targets {
			response := serveRequest(spa, method, target, map[string]string{"Sec-Fetch-Mode": "navigate", "Accept": "text/html"})
			if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), "SECRET=") || strings.Contains(response.Body.String(), "<main>") {
				t.Fatalf("expected hidden path to remain 404 in SPA mode for %s %s, got %d %q", method, target, response.Code, response.Body.String())
			}
		}
	}

	allowed := newTestHandler(t, dir, handlerOptions{dotfiles: true})
	if response := serveRequest(allowed, http.MethodGet, "/.env", nil); response.Body.String() != "SECRET=1" {
		t.Fatalf("expected --dotfiles to serve .env, got %d %q", response.Code, response.Body.String())
	}
}

func TestStaticHandlerDoesNotFollowSymlinksOutsideRoot(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "site")
	writeFile(t, parent, "secret.txt", "top secret")
	writeFile(t, dir, "inside.txt", "inside")
	if err := os.Symlink(filepath.Join(parent, "secret.txt"), filepath.Join(dir, "escape.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink("inside.txt", filepath.Join(dir, "alias.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	handler := newTestHandler(t, dir, handlerOptions{})

	escape := serveRequest(handler, http.MethodGet, "/escape.txt", nil)
	if escape.Code != http.StatusNotFound || strings.Contains(escape.Body.String(), "top secret") {
		t.Fatalf("expected 404 for escaping symlink, got %d %q", escape.Code, escape.Body.String())
	}

	alias := serveRequest(handler, http.MethodGet, "/alias.txt", nil)
	if alias.Code != http.StatusOK || alias.Body.String() != "inside" {
		t.Fatalf("expected symlink inside root to be served, got %d %q", alias.Code, alias.Body.String())
	}
}

func TestStaticHandlerCompressedResponseDropsRangeAdvertisement(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.css", strings.Repeat("body { color: red; }\n", 200))

	response := serveRequest(newTestHandler(t, dir, handlerOptions{compress: true}), http.MethodGet, "/app.css", map[string]string{"Accept-Encoding": "gzip"})
	if response.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding")
	}
	if response.Header().Get("Accept-Ranges") != "" || response.Header().Get("Content-Length") != "" {
		t.Fatalf("compressed response must not advertise ranges or original length: %v", response.Header())
	}
}

func TestStaticHandlerServesBinaryUncompressedThroughWrapper(t *testing.T) {
	dir := t.TempDir()
	contents := strings.Repeat("\x89PNG\x00\x01", 1000)
	writeFile(t, dir, "image.png", contents)

	response := serveRequest(newTestHandler(t, dir, handlerOptions{compress: true}), http.MethodGet, "/image.png", map[string]string{"Accept-Encoding": "gzip"})
	if response.Code != http.StatusOK || response.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected uncompressed 200, got %d %q", response.Code, response.Header().Get("Content-Encoding"))
	}
	if response.Body.String() != contents {
		t.Fatalf("binary response body was altered")
	}
}

func TestStaticHandlerMissingAssetIsNotCached(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "index.html", "<main>app</main>")

	response := serveRequest(newTestHandler(t, dir, handlerOptions{cacheDuration: time.Hour}), http.MethodGet, "/assets/missing.js", nil)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
	if cacheControl := response.Header().Get("Cache-Control"); strings.Contains(cacheControl, "max-age") {
		t.Fatalf("404 responses must not be cached, got %q", cacheControl)
	}
}
