# servexy

Fast zero-config static file server with optional SPA history fallback. Built with Go and distributed through npm with prebuilt binaries.

## Installation

### Global Installation

```bash
# Using bun
bun install -g servexy

# Using npm
npm install -g servexy
```

### Run without Installing

```bash
# Using bunx
bunx servexy -p 8080

# Using npx
npx servexy -p 8080
```

## Usage

```bash
servexy [OPTIONS]
```

### Options

- `-p, -port <port>` - Port to serve on (default: 8080)
- `-d, -dir <directory>` - Directory to serve (default: current directory)
- `-n, -network` - Allow access from other devices on the network
- `-s, -spa` - Serve `index.html` for missing HTML routes
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

# Serve a Vue/Vite build with history routing fallback
servexy -d ./dist -s

# Allow network access from other devices
servexy -n

# Serve on port 8080 with network access
servexy -p 8080 -n

# Show help
servexy -h
```

## Features

- **Fast**: Built with Go and a small wrapper
- **Simple**: Serve any directory with a single command
- **SPA fallback**: Optional `index.html` fallback for Vue/React router history mode
- **Network access**: Optional network access for sharing across devices
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Install-script free**: npm packages ship prebuilt binaries instead of downloading at install time
- **No runtime dependencies**: Self-contained binary, no runtime npm dependencies

## Platform Support

servexy supports the following platforms:

- **Windows**: x64, ARM64
- **macOS**: x64 (Intel), ARM64 (Apple Silicon)
- **Linux**: x64, ARM64

## Development

### Building from Source

This project uses bun for package scripts and Go for the server binary.

```bash
git clone https://github.com/rahatsagor/servexy.git
cd servexy

bun run test
bun run build:current
bun run build:binaries
```

`bun run build:current` writes the binary for your current machine. `bun run build:binaries` writes every release binary to `dist/`. The npm package verifies those binaries during packing and never downloads or builds anything during user install.

## Publishing

Publishing is handled by GitHub Actions using npm trusted publishing. On npmjs.com, configure package `servexy` with this trusted publisher:

- Publisher: GitHub Actions
- Owner: `rahatsagor`
- Repository: `servexy`
- Workflow filename: `publish.yml`
- Allowed action: `npm publish`

Then create and push a version tag:

```bash
git tag v1.1.0
git push origin v1.1.0
```

The workflow runs tests, rebuilds every release binary, verifies the package, and publishes to npm. To publish manually from a local machine, run:

```bash
bun run test
bun run build:binaries
bun publish --dry-run
bun publish --access public
```

If cross-compiling fails on your local machine, run the build and publish steps on a Linux CI runner with Go 1.26.4 and Bun installed. Do not publish while `bun publish --dry-run` reports missing or stale binaries.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
