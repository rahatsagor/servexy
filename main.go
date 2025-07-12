package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

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

func main() {
	flag.Usage = func() {
		fmt.Println(`🚀 Servexy - Simple HTTP File Server

A lightweight HTTP server for serving files and directories.

USAGE:
	servexy [OPTIONS]

OPTIONS:
	-p, -port <port>     Port to serve on (default: 8080)
	-d, -dir <directory> Directory to serve (default: current directory)
	-n, -network         Allow access from other devices on the network
	-h, -help           Show this help message

EXAMPLES:
	servexy                          # Serve current directory on port 8080
	servexy -p 3000                  # Serve current directory on port 3000
	servexy -d /home/user/docs       # Serve /home/user/docs on port 8080
	servexy -p 9000 -d ./public      # Serve ./public on port 9000
	servexy -n                       # Allow network access from other devices
	servexy -p 8080 -n               # Serve on port 8080 with network access
	servexy -h                       # Show help

Visit http://localhost:<port> to access the served files.
With -n flag, access from other devices using http://<your-ip>:<port>`)
	}

	port := flag.String("p", "8080", "Port to serve on")
	portLong := flag.String("port", "8080", "Port to serve on")
	dir := flag.String("d", ".", "Directory to serve")
	dirLong := flag.String("dir", ".", "Directory to serve")
	help := flag.Bool("h", false, "Show help")
	helpLong := flag.Bool("help", false, "Show help")
	network := flag.Bool("n", false, "Allow access from other devices on the network")
	networkLong := flag.Bool("network", false, "Allow access from other devices on the network")
	flag.Parse()

	if *help || *helpLong {
		flag.Usage()
		return
	}

	finalPort := *port
	if *portLong != "8080" && *port == "8080" {
		finalPort = *portLong
	} else if *port != "8080" {
		finalPort = *port
	}

	finalDir := *dir
	if *dirLong != "." && *dir == "." {
		finalDir = *dirLong
	} else if *dir != "." {
		finalDir = *dir
	}

	serveDir, err := filepath.Abs(finalDir)
	if err != nil {
		log.Fatalf("❌ Error resolving directory: %v", err)
	}

	fileInfo, err := os.Stat(serveDir)
	if err != nil {
		log.Fatalf("❌ Error accessing directory: %v", err)
	}
	if !fileInfo.IsDir() {
		log.Fatalf("❌ The specified path is not a directory: %s", serveDir)
	}

	fileServer := http.FileServer(http.Dir(serveDir))

	http.Handle("/", fileServer)

	fmt.Printf("🔥 Servexy is running... \n")
	fmt.Printf("📁 Serving directory: %s\n", serveDir)
	host := "localhost"
	if *network || *networkLong {
		host = ""
		fmt.Printf("🌐 Server starting on all interfaces:\n")
		fmt.Printf("   Local:   http://localhost:%s\n", finalPort)
		if ip := getLocalIP(); ip != "" {
			fmt.Printf("   Network: http://%s:%s\n", ip, finalPort)
		}
	} else {
		fmt.Printf("🌐 Server starting on http://localhost:%s\n", finalPort)
	}
	fmt.Printf("⏹️  Press Ctrl+C to stop the server\n")

	err = http.ListenAndServe(host+":"+finalPort, nil)
	if err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
