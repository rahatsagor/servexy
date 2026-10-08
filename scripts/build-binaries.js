#!/usr/bin/env bun

"use strict";

const { createHash } = require("node:crypto");
const {
	chmodSync,
	existsSync,
	mkdirSync,
	readFileSync,
	rmSync,
	writeFileSync,
} = require("node:fs");
const { join } = require("node:path");
const { spawnSync } = require("node:child_process");
const { releasePlatforms, binaryName, currentPlatform } = require("./platforms.js");

const rootDirectory = join(__dirname, "..");
const distDirectory = join(rootDirectory, "dist");
const packageMetadata = JSON.parse(
	readFileSync(join(rootDirectory, "package.json"), "utf8"),
);

function run(command, args, options = {}) {
	const result = spawnSync(command, args, {
		cwd: rootDirectory,
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

function writeChecksums() {
	const checksums = releasePlatforms.map((platform) => {
		const name = binaryName(platform);
		const contents = readFileSync(join(distDirectory, name));
		const digest = createHash("sha256").update(contents).digest("hex");
		return `${digest}  ${name}`;
	});

	writeFileSync(join(distDirectory, "SHA256SUMS"), `${checksums.join("\n")}\n`);
}

const currentOnly = process.argv.includes("--current");
const platforms = currentOnly ? [currentPlatform()] : releasePlatforms;
const goBinary = process.env.GO_BINARY || "go";
const ldflags = `-s -w -buildid= -X main.version=${packageMetadata.version}`;

mkdirSync(distDirectory, { recursive: true });
run(goBinary, ["version"]);

if (!currentOnly) {
	for (const platform of releasePlatforms) {
		const output = join(distDirectory, binaryName(platform));
		if (existsSync(output)) {
			rmSync(output);
		}
	}
	if (existsSync(join(distDirectory, "SHA256SUMS"))) {
		rmSync(join(distDirectory, "SHA256SUMS"));
	}
}

for (const platform of platforms) {
	const name = binaryName(platform);
	const output = join(distDirectory, name);

	console.log(`Building ${platform.os}/${platform.arch} -> dist/${name}`);
	run(
		goBinary,
		[
			"build",
			"-buildvcs=false",
			"-trimpath",
			`-ldflags=${ldflags}`,
			"-o",
			output,
			".",
		],
		{
			env: {
				...process.env,
				CGO_ENABLED: "0",
				GOOS: platform.os,
				GOARCH: platform.arch,
			},
		},
	);

	if (platform.os !== "windows") {
		chmodSync(output, 0o755);
	}
}

if (!currentOnly) {
	writeChecksums();
}
