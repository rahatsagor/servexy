#!/usr/bin/env bun

"use strict";

const { createHash } = require("node:crypto");
const { existsSync, readFileSync, statSync } = require("node:fs");
const { join } = require("node:path");
const { spawnSync } = require("node:child_process");
const { releasePlatforms, binaryName, currentPlatform } = require("./platforms.js");

const rootDirectory = join(__dirname, "..");
const distDirectory = join(rootDirectory, "dist");
const packageMetadata = JSON.parse(
	readFileSync(join(rootDirectory, "package.json"), "utf8"),
);

function fail(message) {
	console.error(`servexy package verification failed: ${message}`);
	process.exit(1);
}

function verifyPackageMetadata() {
	for (const script of ["preinstall", "install", "postinstall"]) {
		if (packageMetadata.scripts?.[script]) {
			fail(`unsafe lifecycle script is not allowed: ${script}`);
		}
	}

	if (process.env.GITHUB_REF_TYPE === "tag") {
		const expectedTag = `v${packageMetadata.version}`;
		if (process.env.GITHUB_REF_NAME !== expectedTag) {
			fail(
				`tag ${process.env.GITHUB_REF_NAME} does not match package version ${packageMetadata.version}`,
			);
		}
	}
}

function readChecksums() {
	const checksumPath = join(distDirectory, "SHA256SUMS");
	if (!existsSync(checksumPath)) {
		fail("dist/SHA256SUMS is missing; run 'bun run build:binaries'");
	}

	const entries = new Map();
	for (const line of readFileSync(checksumPath, "utf8").trim().split("\n")) {
		const match = line.match(/^([a-f0-9]{64})  (.+)$/);
		if (!match) {
			fail(`invalid checksum line: ${line}`);
		}
		entries.set(match[2], match[1]);
	}
	return entries;
}

function verifyBinary(platform, checksums) {
	const name = binaryName(platform);
	const binaryPath = join(distDirectory, name);
	if (!existsSync(binaryPath)) {
		fail(`dist/${name} is missing`);
	}

	const stats = statSync(binaryPath);
	if (!stats.isFile() || stats.size < 1_000_000) {
		fail(`dist/${name} is not a valid release binary`);
	}
	if (platform.os !== "windows" && (stats.mode & 0o111) === 0) {
		fail(`dist/${name} is not executable`);
	}

	const contents = readFileSync(binaryPath);
	if (!contents.includes(Buffer.from(packageMetadata.version))) {
		fail(`dist/${name} does not contain version ${packageMetadata.version}`);
	}

	if (checksums) {
		const expected = checksums.get(name);
		const actual = createHash("sha256").update(contents).digest("hex");
		if (!expected || actual !== expected) {
			fail(`dist/${name} does not match dist/SHA256SUMS`);
		}
	}
}

function verifyCurrentBinary(platform) {
	const binaryPath = join(distDirectory, binaryName(platform));
	const result = spawnSync(binaryPath, ["--version"], {
		cwd: rootDirectory,
		encoding: "utf8",
	});

	if (result.error || result.status !== 0) {
		fail(`current binary could not run: ${result.error?.message || result.stderr}`);
	}
	const expected = `servexy ${packageMetadata.version}`;
	if (result.stdout.trim() !== expected) {
		fail(`current binary reported ${JSON.stringify(result.stdout.trim())}, expected ${expected}`);
	}
}

verifyPackageMetadata();

const currentOnly = process.argv.includes("--current");
const platforms = currentOnly ? [currentPlatform()] : releasePlatforms;
const checksums = currentOnly ? null : readChecksums();

for (const platform of platforms) {
	verifyBinary(platform, checksums);
}
verifyCurrentBinary(currentPlatform());

console.log(
	currentOnly
		? `Verified servexy ${packageMetadata.version} for the current platform.`
		: `Verified servexy ${packageMetadata.version} release binaries and checksums.`,
);
