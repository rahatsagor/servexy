# 🚀 Servexy - The Fastest HTTP File Server

Lightning-fast HTTP server for serving files and directories with zero configuration. Built with Go and distributed via npm.

## Installation

### Global Installation

```bash
# Using npm
npm install -g servexy

# Using bun
bun install -g servexy

# Using yarn
yarn global add servexy
```

### Run without Installing (npx/bunx)

```bash
# Using npx
npx servexy -p 8080

# Using bunx
bunx servexy -p 8080
```

## Usage

```bash
servexy [OPTIONS]
```

### Options

- `-p, -port <port>` - Port to serve on (default: 8080)
- `-d, -dir <directory>` - Directory to serve (default: current directory)
- `-n, -network` - Allow access from other devices on the network
- `-h, -help` - Show help message

### Examples

```bash
# Serve current directory on port 8080
servexy

# Serve current directory on port 3000
servexy -p 3000

# Serve specific directory
servexy -d /home/user/docs

# Serve with custom port and directory
servexy -p 9000 -d ./public

# Allow network access from other devices
servexy -n

# Serve on port 8080 with network access
servexy -p 8080 -n

# Show help
servexy -h
```

## Features

- ⚡ **Lightning Fast**: Built with Go for unmatched performance
- 📁 **Simple**: Serve any directory with a single command
- 🌐 **Network Access**: Optional network access for sharing across devices
- 🔧 **Cross-platform**: Works on Windows, macOS, and Linux
- 📦 **Easy Installation**: Install via npm/bun package managers
- ⚡ **No Dependencies**: Self-contained binary, no runtime dependencies

## Platform Support

Servexy supports the following platforms:

- **Windows**: x64, ARM64
- **macOS**: x64 (Intel), ARM64 (Apple Silicon)
- **Linux**: x64, ARM64

## Development

### Building from Source

If you have Go installed, you can build the project:

```bash
# Clone the repository
git clone https://github.com/rahatsagor/servexy.git
cd servexy

# Build for current platform
go build -o servexy main.go
```

For maintainers: See development documentation for release processes.

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
