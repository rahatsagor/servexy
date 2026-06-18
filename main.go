package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type staticHandler struct {
	fileServer http.Handler
	fileSystem http.FileSystem
	spa        bool
}

func newStaticHandler(dir string, spa bool) http.Handler {
	fileSystem := http.Dir(dir)

	return &staticHandler{
		fileServer: http.FileServer(fileSystem),
		fileSystem: fileSystem,
		spa:        spa,
	}
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.shouldServeIndex(r) && !h.pathExists(r.URL.Path) {
		indexRequest := r.Clone(r.Context())
		indexRequest.URL.Path = "/"
		indexRequest.URL.RawPath = ""
		h.fileServer.ServeHTTP(w, indexRequest)
		return
	}

	h.fileServer.ServeHTTP(w, r)
}

func (h *staticHandler) shouldServeIndex(r *http.Request) bool {
	if !h.spa {
		return false
	}

	if path.Ext(r.URL.Path) != "" {
		return false
	}

	accept := r.Header.Get("Accept")
	return accept == "" || strings.Contains(accept, "text/html")
}

func (h *staticHandler) pathExists(requestPath string) bool {
	name := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	if name == "" {
		name = "."
	}

	file, err := h.fileSystem.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()

	return true
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func main() {
	flag.Usage = func() {
		fmt.Println(`servexy - static file server

A fast zero-config HTTP server for serving files and directories.

USAGE:
	servexy [OPTIONS]

OPTIONS:
	-p, -port <port>     Port to serve on (default: 8080)
	-d, -dir <directory> Directory to serve (default: current directory)
	-n, -network         Allow access from other devices on the network
	-s, -spa             Serve index.html for missing HTML routes
	-h, -help           Show this help message

EXAMPLES:
	servexy                          # Serve current directory on port 8080
	servexy -p 3000                  # Serve current directory on port 3000
	servexy -d /home/user/docs       # Serve /home/user/docs on port 8080
	servexy -p 9000 -d ./public      # Serve ./public on port 9000
	servexy -n                       # Allow network access from other devices
	servexy -p 8080 -n               # Serve on port 8080 with network access
	servexy -d ./dist -s             # Serve Vue/React router history fallback
	servexy -h                       # Show help

Visit http://localhost:<port> to access the served files.
With -n flag, access from other devices using http://<your-ip>:<port>`)
	}

	var port string
	var dir string
	var help bool
	var network bool
	var spa bool

	flag.StringVar(&port, "p", "8080", "Port to serve on")
	flag.StringVar(&port, "port", "8080", "Port to serve on")
	flag.StringVar(&dir, "d", ".", "Directory to serve")
	flag.StringVar(&dir, "dir", ".", "Directory to serve")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&network, "n", false, "Allow access from other devices on the network")
	flag.BoolVar(&network, "network", false, "Allow access from other devices on the network")
	flag.BoolVar(&spa, "s", false, "Serve index.html for single-page app routes")
	flag.BoolVar(&spa, "spa", false, "Serve index.html for single-page app routes")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	if err := validatePort(port); err != nil {
		log.Fatalf("invalid port: %v", err)
	}

	serveDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("error resolving directory: %v", err)
	}

	fileInfo, err := os.Stat(serveDir)
	if err != nil {
		log.Fatalf("error accessing directory: %v", err)
	}
	if !fileInfo.IsDir() {
		log.Fatalf("specified path is not a directory: %s", serveDir)
	}

	if spa && !fileExists(filepath.Join(serveDir, "index.html")) {
		log.Fatalf("spa mode requires an index.html file in: %s", serveDir)
	}

	fmt.Printf("servexy is running\n")
	fmt.Printf("serving directory: %s\n", serveDir)
	if spa {
		fmt.Printf("spa fallback enabled: missing HTML routes serve index.html\n")
	}

	host := "127.0.0.1"
	if network {
		host = ""
		fmt.Printf("server starting on all interfaces:\n")
		fmt.Printf("local:   http://localhost:%s\n", port)
		if ip := getLocalIP(); ip != "" {
			fmt.Printf("network: http://%s:%s\n", ip, port)
		}
	} else {
		fmt.Printf("server starting on http://localhost:%s\n", port)
	}
	fmt.Printf("press Ctrl+C to stop the server\n")

	server := &http.Server{
		Addr:              net.JoinHostPort(host, port),
		Handler:           newStaticHandler(serveDir, spa),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
