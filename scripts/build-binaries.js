#!/usr/bin/env bun

const { mkdirSync } = require("fs");
const { join } = require("path");
const { spawnSync } = require("child_process");

const releasePlatforms = [
	{ os: "linux", arch: "amd64" },
	{ os: "linux", arch: "arm64" },
	{ os: "darwin", arch: "amd64" },
	{ os: "darwin", arch: "arm64" },
	{ os: "windows", arch: "amd64", extension: ".exe" },
	{ os: "windows", arch: "arm64", extension: ".exe" },
];

const platformMap = {
	darwin: "darwin",
	linux: "linux",
	win32: "windows",
};

const archMap = {
	x64: "amd64",
	arm64: "arm64",
};

function run(command, args, options = {}) {
	const result = spawnSync(command, args, {
		stdio: "inherit",
		...options,
	});

	if (result.error) {
		throw result.error;
	}

	if (result.status !== 0) {
		throw new Error(`${command} ${args.join(" ")} exited with ${result.status}`);
	}
}

const goBinary = process.env.GO_BINARY || "go";

function currentPlatform() {
	const os = platformMap[process.platform];
	const arch = archMap[process.arch];

	if (!os || !arch) {
		throw new Error(`Unsupported platform: ${process.platform}/${process.arch}`);
	}

	return {
		os,
		arch,
		extension: os === "windows" ? ".exe" : "",
	};
}

const platforms = process.argv.includes("--current")
	? [currentPlatform()]
	: releasePlatforms;

mkdirSync("dist", { recursive: true });

run(goBinary, ["version"]);

for (const platform of platforms) {
	const output = join(
		"dist",
		`servexy-${platform.os}-${platform.arch}${platform.extension || ""}`,
	);

	console.log(`Building ${platform.os}/${platform.arch} -> ${output}`);
	run(goBinary, ["build", "-trimpath", "-ldflags=-s -w", "-o", output, "."], {
		env: {
			...process.env,
			CGO_ENABLED: "0",
			GOOS: platform.os,
			GOARCH: platform.arch,
		},
	});
}
