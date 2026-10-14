<p align="center">
  <a href="https://servexy.rsapplication.com">
    <img src="https://raw.githubusercontent.com/rahatsagor/servexy/main/docs/media/icon.svg" width="96" height="96" alt="servexy icon">
  </a>
</p>

<h1 align="center">servexy</h1>

<h3 align="center">Your files. Now serving.</h3>

<p align="center">
  A tiny, zero-config static file server for previewing builds, testing SPA routes,<br>
  and sharing a folder with the devices on your network.
</p>

<p align="center">
  <b>One command. One native binary. Nothing to configure.</b>
</p>

<p align="center">
  <a href="https://www.npmjs.com/package/servexy"><img src="https://img.shields.io/npm/v/servexy?style=for-the-badge&logo=npm&label=npm&color=d6f77a&labelColor=10120f" alt="npm version"></a>
  <a href="https://servexy.rsapplication.com"><img src="https://img.shields.io/badge/Website-servexy.rsapplication.com-10120f?style=for-the-badge&logo=safari&logoColor=white" alt="Website"></a>
</p>

<p align="center">
  <a href="https://github.com/rahatsagor/servexy/releases/latest"><img src="https://img.shields.io/github/v/release/rahatsagor/servexy?style=flat-square&label=release&color=d6f77a&labelColor=1f1f1f" alt="Latest release"></a>
  <a href="https://github.com/rahatsagor/servexy/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/rahatsagor/servexy/ci.yml?branch=main&style=flat-square&label=CI&labelColor=1f1f1f" alt="CI status"></a>
  <a href="#requirements"><img src="https://img.shields.io/badge/macOS%20%7C%20Linux%20%7C%20Windows-x64%20%2B%20ARM64-1f1f1f?style=flat-square" alt="macOS, Linux and Windows on x64 and ARM64"></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/rahatsagor/servexy?style=flat-square&logo=go&labelColor=1f1f1f&color=1f1f1f" alt="Go version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-1f1f1f?style=flat-square" alt="MIT License"></a>
</p>

<p align="center">
  <a href="https://servexy.rsapplication.com">
    <img src="https://raw.githubusercontent.com/rahatsagor/servexy/main/docs/media/hero.svg" width="900" alt="A terminal runs servexy ./dist --spa --network. It reports the folder being served, that SPA fallback and gzip are on, and prints a local and a network URL.">
  </a>
</p>

<table>
  <tr>
    <td width="33%" valign="top">
      <b>1. Install it</b><br>
      <code>bun add -g servexy</code><br>
      or <code>npm i -g servexy</code>. Or download one file.
    </td>
    <td width="33%" valign="top">
      <b>2. Point it at a folder</b><br>
      <code>servexy ./dist --spa</code><br>
      Flags go before or after the folder.
    </td>
    <td width="33%" valign="top">
      <b>3. Open your browser</b><br>
      <code>http://localhost:8080</code><br>
      That's the whole setup.
    </td>
  </tr>
</table>

## Everything a file server should do

<table>
  <tr>
    <td width="50%" valign="top">
      <h4>SPA routing that tells the truth</h4>
      Refresh <code>/dashboard/settings</code> or <code>/users/jane.doe</code> and your app loads. A missing
      script, a typo in an image path or a <code>fetch()</code> to an API that doesn't exist still gets a real
      <code>404</code>, so broken paths never quietly turn into HTML.
    </td>
    <td width="50%" valign="top">
      <h4>Safe by default</h4>
      Listens on <code>127.0.0.1</code> only. <code>.env</code>, <code>.git/</code> and every other hidden
      path return <code>404</code>. Symlinks can't reach outside the folder you serve, and only
      <code>GET</code> and <code>HEAD</code> are allowed.
    </td>
  </tr>
  <tr>
    <td width="50%" valign="top">
      <h4>Your phone is invited</h4>
      Add <code>--network</code> and servexy prints a LAN address like <code>http://192.168.1.42:8080</code>
      you can open on any phone, tablet or laptop on the same Wi-Fi.
    </td>
    <td width="50%" valign="top">
      <h4>Fast where it counts</h4>
      A compiled Go server with on-the-fly gzip for text, and zero-copy <code>sendfile</code> transfers
      for images, video and other binaries. It's ready to serve in milliseconds.
    </td>
  </tr>
  <tr>
    <td width="50%" valign="top">
      <h4>Caching done right</h4>
      HTML is always revalidated, so a new build shows up on the next refresh. Add
      <code>--cache 24h</code> to let browsers keep your hashed assets.
    </td>
    <td width="50%" valign="top">
      <h4>Runs anywhere, installs nothing</h4>
      One self-contained binary for macOS, Linux and Windows on x64 and ARM64. The npm package has no
      dependencies and no install scripts.
    </td>
  </tr>
</table>

## Contents

- [Requirements](#requirements)
- [Install](#install)
- [Usage](#usage)
- [Recipes](#recipes)
- [SPA routing](#spa-routing)
- [Safe by default](#safe-by-default)
- [How responses are served](#how-responses-are-served)
- [What servexy is not](#what-servexy-is-not)
- [Troubleshooting](#troubleshooting)
- [Build from source](#build-from-source)
- [License](#license)

## Requirements

- macOS, Linux or Windows, on x64 or ARM64.
- For the npm package: Node.js 18 or later, or Bun. The prebuilt binaries need nothing else.

## Install

**With Bun or npm**

```bash
bun add --global servexy
# or
npm install --global servexy
```

Run the same commands again to update. To try it without installing:

```bash
bunx servexy ./dist --spa
# or
npx servexy ./dist --spa
```

The package ships a small launcher plus a native binary for every supported platform. The launcher picks the right one, so there is no download step during install and nothing runs at install time.

**Or download the binary**

1. Download the file for your system from the [latest release](https://github.com/rahatsagor/servexy/releases/latest):

   | System                 | File                        |
   | ---------------------- | --------------------------- |
   | macOS, Apple silicon   | `servexy-darwin-arm64`      |
   | macOS, Intel           | `servexy-darwin-amd64`      |
   | Linux, x64             | `servexy-linux-amd64`       |
   | Linux, ARM64           | `servexy-linux-arm64`       |
   | Windows, x64           | `servexy-windows-amd64.exe` |
   | Windows, ARM64         | `servexy-windows-arm64.exe` |

2. Rename it to `servexy` (`servexy.exe` on Windows) and move it to a folder on your `PATH`. On macOS and Linux, make it executable too:

   ```bash
   chmod +x servexy
   sudo mv servexy /usr/local/bin/
   ```

3. On macOS, a binary downloaded with a browser is quarantined. Clear the flag once:

   ```bash
   xattr -d com.apple.quarantine /usr/local/bin/servexy
   ```

<details>
<summary><b>Verify the download</b></summary>

<br>

Every release includes a `SHA256SUMS` file. Put it next to the binary and run:

```bash
shasum -a 256 --check --ignore-missing SHA256SUMS   # macOS
sha256sum --check --ignore-missing SHA256SUMS       # Linux
```

On Windows, compare the output of `Get-FileHash servexy-windows-amd64.exe` with the matching line in `SHA256SUMS`.

npm releases are published from GitHub Actions with [provenance](https://docs.npmjs.com/generating-provenance-statements), so npmjs.com shows the exact commit and workflow that built each version.

</details>

**Or install with Go**

```bash
go install github.com/rahatsagor/servexy@latest
```

This needs Go 1.27 or later. Builds made this way report their version as `dev`.

## Usage

```text
servexy [OPTIONS] [DIRECTORY]
```

```bash
servexy                          # serve the current folder on http://localhost:8080
servexy ./public --port 3000     # another folder and port
servexy ./dist --spa             # a single-page app with client-side routing
servexy ./dist --spa --cache 1h  # also let browsers cache assets for an hour
servexy ./public --network       # reachable from other devices on your network
```

| Option                  | Default        | What it does                                                                  |
| ----------------------- | -------------- | ----------------------------------------------------------------------------- |
| `DIRECTORY`             | current folder | The folder to serve. Give it here or with `--dir`, not both.                  |
| `-p, --port <port>`     | `8080`         | Port to listen on, from 1 to 65535.                                           |
| `-d, --dir <directory>` | current folder | The folder to serve.                                                          |
| `-s, --spa`             | off            | Serve the root `index.html` for page navigations that don't match a file.     |
| `-n, --network`         | off            | Listen on every network interface instead of only `127.0.0.1`.                |
| `--cache <duration>`    | off            | Let browsers cache files other than HTML, for example `10m`, `1h` or `24h`.   |
| `--no-compress`         | off            | Turn off gzip compression.                                                    |
| `--dotfiles`            | off            | Serve hidden files and folders such as `.env`.                                |
| `-v, --version`         |                | Print the version.                                                            |
| `-h, --help`            |                | Print help.                                                                   |

- Options can come before or after the folder, and values can be written as `--port 3000` or `--port=3000`.
- `--` ends the options, which lets you serve a folder whose name starts with a dash: `servexy -- -drafts`.
- The single-dash forms `-port`, `-dir`, `-spa` and `-network` from older versions still work.

Press **Ctrl+C** to stop. Requests that are still running get up to five seconds to finish; press **Ctrl+C** again to quit at once.

## Recipes

### Preview a production build

| Built with                         | Output folder         | Command                               |
| ---------------------------------- | --------------------- | ------------------------------------- |
| Vite (Vue, React, Svelte, Solid)   | `dist/`               | `servexy ./dist --spa`                |
| Create React App                   | `build/`              | `servexy ./build --spa`               |
| Angular 17 or later                | `dist/<app>/browser/` | `servexy ./dist/my-app/browser --spa` |
| Next.js with `output: 'export'`    | `out/`                | `servexy ./out`                       |
| Astro (static output)              | `dist/`               | `servexy ./dist`                      |
| Hugo                               | `public/`             | `servexy ./public`                    |
| Eleventy                           | `_site/`              | `servexy ./_site`                     |

Vue Router's history mode, React Router's `createBrowserRouter`, and any other router that uses real URLs need `--spa` so a refresh on a deep link works. Hash routing (`/#/settings`) works without it.

To pin it to your project, add it as a dev dependency with `bun add --dev servexy` (or `npm install --save-dev servexy`) and use it in a script:

```json
{
  "scripts": {
    "build": "vite build",
    "preview": "servexy ./dist --spa"
  }
}
```

### Test on your phone

```bash
servexy ./dist --spa --network
```

Open the `network:` address servexy prints on any device on the same Wi-Fi. If nothing loads, check that your firewall allows incoming connections for servexy. Everyone on that network can reach the folder while it runs, so only share what you mean to.

### Run end-to-end tests against the real build

Point [Playwright](https://playwright.dev/docs/test-webserver) at your build output, so tests exercise exactly what you ship:

```ts
// playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  webServer: {
    command: 'bunx servexy ./dist --spa --port 4173',
    url: 'http://127.0.0.1:4173',
    reuseExistingServer: !process.env.CI,
  },
  use: { baseURL: 'http://127.0.0.1:4173' },
});
```

Because missing assets and `fetch()` calls stay `404`, a broken import or API path fails the test instead of hiding behind your `index.html`.

### Cache hashed assets

```bash
servexy ./dist --spa --cache 24h
```

HTML keeps `Cache-Control: no-cache`, so a new deploy shows up on the next refresh. Everything else gets `public, max-age=86400`. Use this when your bundler puts a content hash in file names, like `assets/index-3f9a1c.js`. Without hashes, browsers may keep an old file until the cache runs out.

## SPA routing

<p align="center">
  <img src="https://raw.githubusercontent.com/rahatsagor/servexy/main/docs/media/spa-routing.svg" width="900" alt="With --spa, navigations to /dashboard/settings and /users/jane.doe get index.html. An existing asset is served as itself. A missing script, a fetch() to /api/users and /.env get 404.">
</p>

Most SPA servers send `index.html` for every path they can't find. That keeps refreshes working, but it also answers a missing `app.js` or a mistyped API call with `200 OK` and a page of HTML, which turns a clear `404` into a confusing parse error. servexy only sends your app to requests that are actually page loads.

With `--spa`, each request is answered like this:

1. If the path is a file, or a folder with its own `index.html`, that is served.
2. Otherwise, if the request is a page navigation, the root `index.html` is served with `200 OK`.
3. Anything else gets `404 Not Found`.

<details>
<summary><b>How servexy recognizes a page navigation</b></summary>

<br>

- **Modern browsers** send [Fetch Metadata](https://developer.mozilla.org/docs/Glossary/Fetch_metadata_request_header) headers. A request is a navigation when `Sec-Fetch-Mode` is `navigate`, or when `Sec-Fetch-Dest` is `document`, `iframe` or `frame`. When these headers are present, they decide on their own, so `fetch()`, images, scripts and styles never get the app.
- **Otherwise**, a request whose `Accept` header includes `text/html` or `application/xhtml+xml` counts as a navigation.
- **Tools like curl** send `Accept: */*` or no `Accept` at all. Those get the app only when the path has no file extension, so `curl localhost:8080/settings` returns your app and `curl localhost:8080/missing.js` returns `404`.

That is why dotted routes like `/users/jane.doe` load in the browser, while `/assets/app.jss` stays `404`. `GET` and `HEAD` are treated the same way.

SPA mode needs an `index.html` at the root of the folder and won't start without one. A page load of a folder that has no `index.html` of its own also gets the app, not a file listing.

</details>

## Safe by default

<table>
  <tr>
    <td width="50%" valign="top">
      <h4>Only your computer, unless you say so</h4>
      servexy listens on <code>127.0.0.1</code>. Nothing else on the network can connect until you pass
      <code>--network</code>, which listens on every interface, IPv6 included.
    </td>
    <td width="50%" valign="top">
      <h4>Locked to one folder</h4>
      Files are opened through Go's <a href="https://pkg.go.dev/os#Root"><code>os.Root</code></a>, which
      refuses any path or symlink that leads outside the folder. Symlinks that stay inside it work as usual.
    </td>
  </tr>
  <tr>
    <td width="50%" valign="top">
      <h4>Hidden files stay hidden</h4>
      Any path with a part that starts with a dot, like <code>.env</code>, <code>.git/config</code> or
      <code>public/.htpasswd</code>, returns <code>404</code> and is left out of file listings, in SPA mode
      too. <code>.well-known/</code> is allowed, but hidden files inside it are not.
    </td>
    <td width="50%" valign="top">
      <h4>Read-only</h4>
      Only <code>GET</code> and <code>HEAD</code> are accepted. There is no upload, no delete and nothing
      that runs code from the folder. Every response carries <code>X-Content-Type-Options: nosniff</code>.
    </td>
  </tr>
</table>

Pass `--dotfiles` only when you really want hidden files served.

On Windows, path parts that look like 8.3 short names, such as `GIT~1`, are treated as hidden too, because NTFS can resolve them to `.git` or `.env`. As a result, a file whose name contains `~` followed by a digit isn't served on Windows unless you pass `--dotfiles`.

## How responses are served

<details>
<summary><b>Compression</b></summary>

<br>

- Responses are gzipped when the browser accepts gzip, the body is at least 1 KiB, and the content is text: HTML, CSS, JavaScript, JSON, web manifests, WebAssembly, XML or SVG.
- Images, video, fonts and archives are already compressed, so they are sent as they are, using zero-copy `sendfile` where the system supports it.
- `HEAD` and range requests are never compressed, and compressed responses don't offer byte ranges.
- Every compressible response includes `Vary: Accept-Encoding`, so a cache in front of servexy keeps the gzipped and plain versions apart.

</details>

<details>
<summary><b>Caching</b></summary>

<br>

| Response                                                  | `Cache-Control`                                           |
| --------------------------------------------------------- | --------------------------------------------------------- |
| HTML, paths without an extension, folders, SPA fallback   | `no-cache`                                                |
| Every other file                                          | `no-cache`, or `public, max-age=<seconds>` with `--cache` |
| Errors such as `404`                                      | none                                                      |

Files are sent with `Last-Modified`, so browsers revalidate with a quick `304 Not Modified` instead of downloading the file again.

</details>

<details>
<summary><b>Connections, limits and shutdown</b></summary>

<br>

- Request headers must arrive within 5 seconds and are limited to 1 MiB. Idle keep-alive connections close after 60 seconds.
- There is no overall write timeout, so large downloads on slow connections are never cut off.
- Without `--spa`, a folder without an `index.html` shows a simple file listing, with hidden files left out.
- On Ctrl+C or `SIGTERM`, servexy stops accepting connections and waits up to 5 seconds for running requests before closing them. A second Ctrl+C exits at once.

</details>

## What servexy is not

- **A dev server.** It serves files that already exist. There's no bundling and no hot reload; keep using `vite`, `next dev` and friends while you write code, and use servexy to check the build.
- **An HTTPS server.** It speaks plain HTTP. For HTTPS, put it behind a reverse proxy such as Caddy, or use a tunnel.
- **A full web server.** There are no request logs, rewrites, headers config or virtual hosts. If you need those, nginx or Caddy are the right tools.

## Troubleshooting

<details>
<summary><b><code>port 8080 is already in use</code></b></summary>

<br>

Another program is listening on that port, often another dev server. Stop it, or pick another port with `--port 3000`.

</details>

<details>
<summary><b><code>permission denied for port 80</code></b></summary>

<br>

Most systems only let administrators use ports below 1024. Use a higher port, or run servexy behind a reverse proxy. Windows also reserves some port ranges; list them with:

```powershell
netsh interface ipv4 show excludedportrange protocol=tcp
```

</details>

<details>
<summary><b><code>SPA mode requires index.html</code></b></summary>

<br>

`--spa` needs to point at your build output, such as `dist/` or `build/`, not the project folder. Run your build first.

</details>

<details>
<summary><b>A deep link shows 404 after a refresh</b></summary>

<br>

Add `--spa`. Without it, servexy only serves files that exist, so `/dashboard/settings` is a `404`.

</details>

<details>
<summary><b>My phone can't open the network address</b></summary>

<br>

- Make sure you passed `--network` and both devices are on the same network. Guest networks and some office or hotel Wi-Fi stop devices from reaching each other.
- Use the `network:` address servexy prints, not `localhost`.
- Allow incoming connections for servexy in your firewall. If the macOS firewall is on, it asks the first time; on Windows, allow it for private networks.

</details>

<details>
<summary><b>A file exists but returns 404</b></summary>

<br>

- Check whether its path has a hidden part, such as `.cache/` or `.vitepress/`. Pass `--dotfiles` if you really mean to serve it.
- Check whether it's a symlink that points outside the folder you're serving.
- On Windows, check whether a part of the path looks like `NAME~1`.

</details>

<details>
<summary><b><code>binary not found</code> from the npm package</b></summary>

<br>

The package manager skipped some of the package's files. Reinstall servexy. If you run it from a clone of this repository, build the binaries first with `bun run build:binaries`.

</details>

Still stuck? [Open an issue](https://github.com/rahatsagor/servexy/issues/new) with your operating system, the command you ran and what servexy printed.

## Build from source

You need [Go](https://go.dev/dl/) 1.27.1 or later, [Bun](https://bun.sh/) 1.4.2 or later, and Node.js 18 or later for the launcher tests.

```bash
git clone https://github.com/rahatsagor/servexy.git
cd servexy
go run . ./some-folder --spa   # run from source
```

| Command                  | What it does                                                   |
| ------------------------ | -------------------------------------------------------------- |
| `bun run check`          | Go tests and `go vet`.                                         |
| `bun run test:race`      | Tests with the race detector.                                  |
| `bun run vuln`           | Scan for known vulnerabilities with `govulncheck`.             |
| `bun run build:current`  | Build the binary for this machine into `dist/`.                |
| `bun run verify:current` | Check that binary runs and reports the package version.        |
| `bun run test:launcher`  | Test the npm launcher, including Ctrl+C handling.              |
| `bun run build:binaries` | Build all six release binaries and `dist/SHA256SUMS`.          |

<details>
<summary><b>Project layout</b></summary>

<br>

| Path                 | What's inside                                                          |
| -------------------- | ---------------------------------------------------------------------- |
| `main.go`            | The whole server: options, file system, SPA routing and compression.   |
| `main_test.go`       | Tests for routing, compression, caching, hidden files and options.     |
| `bin/servexy`        | The npm launcher that starts the right native binary.                  |
| `scripts/`           | Build, package verification and launcher test scripts.                 |
| `docs/media/`        | Images used in this README.                                            |
| `.github/workflows/` | CI on Linux, macOS and Windows, and the release workflow.              |

</details>

<details>
<summary><b>Releasing</b></summary>

<br>

Releases are built and published by GitHub Actions when a version tag is pushed. npm trusted publishing is set up for the `rahatsagor/servexy` repository and the `publish.yml` workflow, so no npm token is stored anywhere.

1. Set the new `version` in `package.json` and commit it.
2. Tag the commit and push both:

   ```bash
   git tag vX.Y.Z
   git push origin main vX.Y.Z
   ```

The workflow tests, vets and scans the code, builds all six binaries, checks their checksums and that the tag matches `package.json`, creates the GitHub release, and publishes to npm with provenance. Every workflow action is pinned to a commit SHA.

</details>

Bug reports and ideas are welcome as [issues](https://github.com/rahatsagor/servexy/issues).

## License

Free and open source under the [MIT License](LICENSE). © 2026 Rahat Sagor.

<p align="center">
  <br>
  <a href="https://servexy.rsapplication.com">
    <img src="https://raw.githubusercontent.com/rahatsagor/servexy/main/docs/media/icon.svg" width="40" height="40" alt="servexy">
  </a>
  <br>
  <a href="https://servexy.rsapplication.com"><b>servexy.rsapplication.com</b></a>
</p>
