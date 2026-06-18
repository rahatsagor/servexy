#!/usr/bin/env bun

const { existsSync, readFileSync } = require("fs");
const { join } = require("path");

const requiredBinaries = [
	"servexy-linux-amd64",
	"servexy-linux-arm64",
	"servexy-darwin-amd64",
	"servexy-darwin-arm64",
	"servexy-windows-amd64.exe",
	"servexy-windows-arm64.exe",
];
const binaryMarker = "spa fallback enabled";

const missing = requiredBinaries.filter((name) => !existsSync(join("dist", name)));

if (missing.length > 0) {
	console.error("Missing release binaries:");
	for (const name of missing) {
		console.error(`- dist/${name}`);
	}
	console.error("Run bun run build:binaries before publishing.");
	process.exit(1);
}

const stale = requiredBinaries.filter((name) => {
	const contents = readFileSync(join("dist", name));
	return !contents.includes(binaryMarker);
});

if (stale.length > 0) {
	console.error("Stale release binaries detected:");
	for (const name of stale) {
		console.error(`- dist/${name}`);
	}
	console.error("Rebuild all binaries before publishing.");
	process.exit(1);
}
