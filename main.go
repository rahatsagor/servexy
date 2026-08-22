package main

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultPort       = "8080"
	minimumGzipSize   = 1024
	shutdownTimeout   = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
	idleTimeout       = 60 * time.Second
)

var version = "dev"

type cliOptions struct {
	port          string
	dir           string
	network       bool
	spa           bool
	compress      bool
	cacheDuration time.Duration
	help          bool
	version       bool
}

type staticHandler struct {
	fileServer    http.Handler
	fileSystem    http.FileSystem
	spa           bool
	compress      bool
	cacheDuration time.Duration
}

type gzipResponseWriter struct {
	http.ResponseWriter
	compressor  *gzip.Writer
	wroteHeader bool
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		writer, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		if err != nil {
			panic(err)
		}
		return writer
	},
}

func newStaticHandler(dir string, spa bool, compress bool, cacheDuration time.Duration) http.Handler {
	fileSystem := http.Dir(dir)

	return &staticHandler{
		fileServer:    http.FileServer(fileSystem),
		fileSystem:    fileSystem,
		spa:           spa,
		compress:      compress,
		cacheDuration: cacheDuration,
	}
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	request := r
	fallback := h.shouldServeIndex(r) && !h.pathIsServable(r.URL.Path)
	if fallback {
		request = r.Clone(r.Context())
		request.URL.Path = "/"
		request.URL.RawPath = ""
	}

	h.setCacheHeader(w, r.URL.Path, fallback)

	if h.compress && request.Method == http.MethodGet && request.Header.Get("Range") == "" && acceptsGzip(request.Header.Get("Accept-Encoding")) {
		compressed := &gzipResponseWriter{ResponseWriter: w}
		h.fileServer.ServeHTTP(compressed, request)
		compressed.close()
		return
	}

	h.fileServer.ServeHTTP(w, request)
}

func (h *staticHandler) shouldServeIndex(r *http.Request) bool {
	return h.spa && isNavigationRequest(r)
}

func (h *staticHandler) pathIsServable(requestPath string) bool {
	name := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	if name == "" {
		name = "."
	}

	file, err := h.fileSystem.Open(name)
	if err != nil {
		return false
	}

	info, err := file.Stat()
	file.Close()
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return info.Mode().IsRegular()
	}

	index, err := h.fileSystem.Open(path.Join(name, "index.html"))
	if err != nil {
		return false
	}
	defer index.Close()

	indexInfo, err := index.Stat()
	return err == nil && indexInfo.Mode().IsRegular()
}

func (h *staticHandler) setCacheHeader(w http.ResponseWriter, requestPath string, fallback bool) {
	extension := strings.ToLower(path.Ext(requestPath))
	isHTML := extension == "" || extension == ".html" || extension == ".htm"

	if fallback || isHTML || h.cacheDuration <= 0 {
		w.Header().Set("Cache-Control", "no-cache")
		return
	}

	maxAge := int64(h.cacheDuration / time.Second)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
}

func isNavigationRequest(r *http.Request) bool {
	switch strings.ToLower(r.Header.Get("Sec-Fetch-Mode")) {
	case "navigate":
		return true
	}

	switch strings.ToLower(r.Header.Get("Sec-Fetch-Dest")) {
	case "document", "frame", "iframe":
		return true
	}

	accept := r.Header.Get("Accept")
	if accept == "" {
		return path.Ext(r.URL.Path) == ""
	}

	wildcardAccepted := false
	for _, item := range strings.Split(accept, ",") {
		mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(item))
		if err != nil {
			continue
		}

		quality := 1.0
		if value, ok := params["q"]; ok {
			quality, err = strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
		}
		if quality <= 0 {
			continue
		}

		switch strings.ToLower(mediaType) {
		case "text/html", "application/xhtml+xml":
			return true
		case "*/*":
			wildcardAccepted = true
		}
	}

	return wildcardAccepted && path.Ext(r.URL.Path) == ""
}

func acceptsGzip(header string) bool {
	if header == "" {
		return false
	}

	gzipQuality := -1.0
	wildcardQuality := -1.0
	for _, item := range strings.Split(header, ",") {
		parts := strings.Split(strings.TrimSpace(item), ";")
		encoding := strings.ToLower(strings.TrimSpace(parts[0]))
		quality := 1.0

		for _, parameter := range parts[1:] {
			name, value, found := strings.Cut(strings.TrimSpace(parameter), "=")
			if !found || !strings.EqualFold(name, "q") {
				continue
			}

			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil {
				quality = 0
			} else {
				quality = parsed
			}
		}

		switch encoding {
		case "gzip":
			gzipQuality = quality
		case "*":
			wildcardQuality = quality
		}
	}

	if gzipQuality >= 0 {
		return gzipQuality > 0
	}
	return wildcardQuality > 0
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if shouldCompressResponse(w.Header(), statusCode) {
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")
		addVaryHeader(w.Header(), "Accept-Encoding")

		w.compressor = gzipWriterPool.Get().(*gzip.Writer)
		w.compressor.Reset(w.ResponseWriter)
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(contents []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.compressor != nil {
		return w.compressor.Write(contents)
	}
	return w.ResponseWriter.Write(contents)
}

func (w *gzipResponseWriter) close() {
	if w.compressor == nil {
		return
	}

	_ = w.compressor.Close()
	w.compressor.Reset(io.Discard)
	gzipWriterPool.Put(w.compressor)
}

func shouldCompressResponse(header http.Header, statusCode int) bool {
	if statusCode != http.StatusOK || header.Get("Content-Encoding") != "" {
		return false
	}

	if value := header.Get("Content-Length"); value != "" {
		length, err := strconv.ParseInt(value, 10, 64)
		if err == nil && length < minimumGzipSize {
			return false
		}
	}

	mediaType, _, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return false
	}

	if strings.HasPrefix(mediaType, "text/") {
		return true
	}

	switch mediaType {
	case "application/javascript",
		"application/json",
		"application/manifest+json",
		"application/wasm",
		"application/xhtml+xml",
		"application/xml",
		"image/svg+xml":
		return true
	default:
		return false
	}
}

func addVaryHeader(header http.Header, value string) {
	for _, item := range header.Values("Vary") {
		for _, existing := range strings.Split(item, ",") {
			if strings.EqualFold(strings.TrimSpace(existing), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}

func parseOptions(args []string) (cliOptions, error) {
	options := cliOptions{
		port:     defaultPort,
		dir:      ".",
		compress: true,
	}

	var positional []string
	directorySet := false
	flagsEnded := false

	for index := 0; index < len(args); index++ {
		argument := args[index]
		if flagsEnded {
			positional = append(positional, argument)
			continue
		}
		if argument == "--" {
			flagsEnded = true
			continue
		}

		name, inlineValue, hasInlineValue := strings.Cut(argument, "=")
		takeValue := func() (string, error) {
			if hasInlineValue {
				if inlineValue == "" {
					return "", fmt.Errorf("%s requires a value", name)
				}
				return inlineValue, nil
			}
			if index+1 >= len(args) {
				return "", fmt.Errorf("%s requires a value", name)
			}
			index++
			return args[index], nil
		}
		rejectValue := func() error {
			if hasInlineValue {
				return fmt.Errorf("%s does not accept a value", name)
			}
			return nil
		}

		switch name {
		case "-p", "--port", "-port":
			value, err := takeValue()
			if err != nil {
				return options, err
			}
			options.port = value
		case "-d", "--dir", "-dir":
			value, err := takeValue()
			if err != nil {
				return options, err
			}
			options.dir = value
			directorySet = true
		case "-n", "--network", "-network":
			if err := rejectValue(); err != nil {
				return options, err
			}
			options.network = true
		case "-s", "--spa", "-spa":
			if err := rejectValue(); err != nil {
				return options, err
			}
			options.spa = true
		case "--no-compress":
			if err := rejectValue(); err != nil {
				return options, err
			}
			options.compress = false
		case "--cache":
			value, err := takeValue()
			if err != nil {
				return options, err
			}
			options.cacheDuration, err = time.ParseDuration(value)
			if err != nil || options.cacheDuration < 0 {
				return options, fmt.Errorf("invalid cache duration %q (examples: 0, 10m, 1h)", value)
			}
			if options.cacheDuration > 0 && options.cacheDuration < time.Second {
				return options, errors.New("cache duration must be zero or at least one second")
			}
		case "-h", "--help", "-help":
			if err := rejectValue(); err != nil {
				return options, err
			}
			options.help = true
		case "-v", "--version", "-version":
			if err := rejectValue(); err != nil {
				return options, err
			}
			options.version = true
		default:
			if strings.HasPrefix(argument, "-") {
				return options, fmt.Errorf("unknown option: %s", argument)
			}
			positional = append(positional, argument)
		}
	}

	if len(positional) > 1 {
		return options, errors.New("only one directory can be served")
	}
	if len(positional) == 1 {
		if directorySet {
			return options, errors.New("directory cannot be provided both positionally and with --dir")
		}
		options.dir = positional[0]
	}

	if err := validatePort(options.port); err != nil {
		return options, fmt.Errorf("invalid port: %w", err)
	}

	return options, nil
}

func usage() string {
	return `servexy - fast, dependency-free static file server

USAGE:
  servexy [OPTIONS] [DIRECTORY]

OPTIONS:
  -p, --port <port>       Port to serve on (default: 8080)
  -d, --dir <directory>   Directory to serve (default: current directory)
  -n, --network           Allow access from other devices on the network
  -s, --spa               Fall back to index.html for SPA navigation routes
      --cache <duration>  Cache non-HTML files (examples: 10m, 1h, 24h)
      --no-compress       Disable gzip compression
  -v, --version           Show the installed version
  -h, --help              Show this help message

EXAMPLES:
  servexy
  servexy ./public --port 3000
  servexy ./dist --spa
  servexy ./dist --spa --cache 1h
  servexy ./public --network

Options may appear before or after DIRECTORY.`
}

func getLocalIP() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addresses {
		ipNetwork, ok := address.(*net.IPNet)
		if ok && !ipNetwork.IP.IsLoopback() && ipNetwork.IP.To4() != nil {
			return ipNetwork.IP.String()
		}
	}
	return ""
}

func validatePort(port string) error {
	value, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be a number: %s", port)
	}
	if value < 1 || value > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && info.Mode().IsRegular()
}

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	if options.help {
		fmt.Println(usage())
		return nil
	}
	if options.version {
		fmt.Printf("servexy %s\n", version)
		return nil
	}

	serveDir, err := filepath.Abs(options.dir)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}

	fileInfo, err := os.Stat(serveDir)
	if err != nil {
		return fmt.Errorf("access directory: %w", err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", serveDir)
	}
	if options.spa && !fileExists(filepath.Join(serveDir, "index.html")) {
		return fmt.Errorf("SPA mode requires index.html in: %s", serveDir)
	}

	host := "127.0.0.1"
	if options.network {
		host = "0.0.0.0"
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(host, options.port))
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	fmt.Printf("servexy %s\n", version)
	fmt.Printf("serving %s\n", serveDir)
	if options.spa {
		fmt.Println("SPA fallback enabled")
	}
	if options.compress {
		fmt.Println("gzip compression enabled")
	}
	if options.cacheDuration > 0 {
		fmt.Printf("asset cache enabled: %s\n", options.cacheDuration)
	}

	if options.network {
		fmt.Printf("local:   http://localhost:%s\n", options.port)
		if ip := getLocalIP(); ip != "" {
			fmt.Printf("network: http://%s:%s\n", ip, options.port)
		}
	} else {
		fmt.Printf("local: http://localhost:%s\n", options.port)
	}
	fmt.Println("press Ctrl+C to stop")

	server := &http.Server{
		Handler:           newStaticHandler(serveDir, options.spa, options.compress, options.cacheDuration),
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    1 << 20,
		ErrorLog:          log.New(os.Stderr, "servexy: ", 0),
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	signalContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server failed: %w", err)
		}
		return nil
	case <-signalContext.Done():
		fmt.Println("\nshutting down")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shut down server: %w", err)
	}

	if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "servexy: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'servexy --help' for usage.")
		os.Exit(1)
	}
}
