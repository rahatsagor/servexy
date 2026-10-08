# servexy

Fast, production-ready static file server with reliable SPA history fallback. The server is a dependency-free Go binary distributed through npm with no install scripts or runtime package dependencies.

## Install

```bash
# Bun
bun add --global servexy

# npm
npm install --global servexy
```

Run it without a global install:

```bash
bunx servexy ./dist --spa
npx servexy ./dist --spa
```

## Usage

```text
servexy [OPTIONS] [DIRECTORY]

OPTIONS:
  -p, --port <port>       Port to serve on (default: 8080)
  -d, --dir <directory>   Directory to serve (default: current directory)
  -n, --network           Allow access from other devices on the network
  -s, --spa               Fall back to index.html for SPA navigation routes
      --cache <duration>  Cache non-HTML files (examples: 10m, 1h, 24h)
      --no-compress       Disable gzip compression
      --dotfiles          Serve dotfiles such as .env (hidden by default)
  -v, --version           Show the installed version
  -h, --help              Show help
```

Options can appear before or after the directory:

```bash
servexy
servexy ./public --port 3000
servexy --network ./public
servexy ./dist --spa --cache 1h
```

The server listens only on `127.0.0.1` by default. Use `--network` only when the files should be reachable by other devices; it listens on all available network interfaces, including IPv6 where supported.

## Security

- Files are served through an `os.Root`, which rejects paths and symlinks that escape the served directory. Relative symlinks that stay inside it work normally; absolute symlinks are rejected.
- Paths containing dotfiles or dot-directories (`.env`, `.git/`, `.htpasswd`, ...) return `404`, including in SPA mode, and hidden names are excluded from directory listings. `.well-known` directories are allowed; hidden files inside them remain blocked. Pass `--dotfiles` to serve hidden paths.

## Vue and Other SPAs

Build the app, then enable SPA mode when serving its output directory:

```bash
bun run build
servexy ./dist --spa
```

SPA fallback is navigation-aware:

- Existing files and nested `index.html` files are served normally.
- Browser navigation routes fall back to the root `index.html`, including routes containing dots.
- Missing JavaScript, CSS, JSON, image, and API-like requests remain real `404` responses.
- `GET` and `HEAD` refreshes behave consistently.

For Vue Router, history mode works with `--spa`. Hash mode does not need fallback routing.

## Production Behavior

- Gzip compression is enabled for compressible responses larger than 1 KiB.
- Range requests are never dynamically compressed, and compressed responses do not advertise range support.
- Uncompressed files (images, video, archives) retain Go's zero-copy `sendfile` path where supported.
- HTML and SPA fallback responses use `Cache-Control: no-cache`.
- `--cache 1h` applies a public max-age to non-HTML assets.
- `X-Content-Type-Options: nosniff` is included on responses.
- The server uses bounded header and idle timeouts and shuts down cleanly on termination. A second Ctrl+C exits immediately.
- Only `GET` and `HEAD` are accepted.

servexy supports Linux, macOS, and Windows on x64 and ARM64.

## Development

Requirements: Go 1.27.1, Bun 1.4.2, and Node.js 18 or newer for npm launcher tests.

```bash
git clone https://github.com/rahatsagor/servexy.git
cd servexy

bun run check
bun run test:race
bun run vuln
bun run build:current
bun run verify:current
bun run test:launcher
```

Build every release binary and verify the package:

```bash
bun run build:binaries
bun scripts/verify-package.js
bun publish --dry-run --access public
```

Full builds produce `dist/SHA256SUMS`. Package verification checks every binary's version, executable mode, size, and checksum before npm can pack or publish it.

## Release

GitHub Actions publishes npm and GitHub releases from version tags using npm trusted publishing. The npm trusted publisher must reference:

- Owner: `rahatsagor`
- Repository: `servexy`
- Workflow: `publish.yml`

Update `package.json`, commit the release, then push the matching tag:

```bash
git tag v1.3.0
git push origin main v1.3.0
```

CI tests Linux, macOS, and Windows and runs `govulncheck`. All workflow actions are pinned to commit SHAs. The release workflow tests, vets, and scans the source, rebuilds all six binaries, verifies checksums and tag/version alignment, creates the GitHub release, and publishes to npm with provenance.

## License

[MIT](LICENSE)
