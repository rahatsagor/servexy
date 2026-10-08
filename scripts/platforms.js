"use strict";

const releasePlatforms = [
	{ os: "linux", arch: "amd64" },
	{ os: "linux", arch: "arm64" },
	{ os: "darwin", arch: "amd64" },
	{ os: "darwin", arch: "arm64" },
	{ os: "windows", arch: "amd64", extension: ".exe" },
	{ os: "windows", arch: "arm64", extension: ".exe" },
];

const platformNames = {
	darwin: "darwin",
	linux: "linux",
	win32: "windows",
};

const architectureNames = {
	x64: "amd64",
	arm64: "arm64",
};

function binaryName(platform) {
	return `servexy-${platform.os}-${platform.arch}${platform.extension || ""}`;
}

function currentPlatform() {
	const os = platformNames[process.platform];
	const arch = architectureNames[process.arch];
	if (!os || !arch) {
		throw new Error(`Unsupported platform: ${process.platform}/${process.arch}`);
	}
	return { os, arch, extension: os === "windows" ? ".exe" : "" };
}

module.exports = { releasePlatforms, binaryName, currentPlatform };
