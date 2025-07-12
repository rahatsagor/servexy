#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const https = require("https");
const { execSync } = require("child_process");

const GITHUB_REPO = "rahatsagor/servexy";
const VERSION = require("../package.json").version;

function getPlatform() {
	const platform = process.platform;
	const arch = process.arch;

	let osName, archName;

	switch (platform) {
		case "darwin":
			osName = "darwin";
			break;
		case "linux":
			osName = "linux";
			break;
		case "win32":
			osName = "windows";
			break;
		default:
			throw new Error(`Unsupported platform: ${platform}`);
	}

	switch (arch) {
		case "x64":
			archName = "amd64";
			break;
		case "arm64":
			archName = "arm64";
			break;
		default:
			throw new Error(`Unsupported architecture: ${arch}`);
	}

	return { osName, archName };
}

function downloadBinary() {
	const { osName, archName } = getPlatform();
	const binaryName = osName === "windows" ? "servexy.exe" : "servexy-binary";
	const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/servexy-${osName}-${archName}${osName === "windows" ? ".exe" : ""}`;

	const binDir = path.join(__dirname, "..", "bin");
	const binaryPath = path.join(binDir, binaryName);

	if (!fs.existsSync(binDir)) {
		fs.mkdirSync(binDir, { recursive: true });
	}

	console.log(`Downloading servexy binary for ${osName}-${archName}...`);
	console.log(`URL: ${downloadUrl}`);

	return new Promise((resolve, reject) => {
		const file = fs.createWriteStream(binaryPath);

		https
			.get(downloadUrl, (response) => {
				if (response.statusCode === 302 || response.statusCode === 301) {
					https
						.get(response.headers.location, (redirectResponse) => {
							if (redirectResponse.statusCode === 200) {
								redirectResponse.pipe(file);
								file.on("finish", () => {
									file.close();
									fs.chmodSync(binaryPath, "755");
									console.log("✅ Servexy binary downloaded successfully!");
									resolve();
								});
							} else {
								reject(
									new Error(
										`Failed to download binary: ${redirectResponse.statusCode}`,
									),
								);
							}
						})
						.on("error", reject);
				} else if (response.statusCode === 200) {
					response.pipe(file);
					file.on("finish", () => {
						file.close();
						fs.chmodSync(binaryPath, "755");
						console.log("✅ Servexy binary downloaded successfully!");
						resolve();
					});
				} else {
					reject(
						new Error(`Failed to download binary: ${response.statusCode}`),
					);
				}
			})
			.on("error", reject);

		file.on("error", (err) => {
			fs.unlink(binaryPath, () => {});
			reject(err);
		});
	});
}

function buildLocally() {
	const { osName, archName } = getPlatform();
	const binaryName = osName === "windows" ? "servexy.exe" : "servexy-binary";
	const binDir = path.join(__dirname, "..", "bin");
	const binaryPath = path.join(binDir, binaryName);

	if (!fs.existsSync(binDir)) {
		fs.mkdirSync(binDir, { recursive: true });
	}

	console.log("Building servexy locally...");

	try {
		const env = { ...process.env };
		if (osName !== process.platform) {
			env.GOOS = osName;
			env.GOARCH = archName === "amd64" ? "amd64" : "arm64";
		}

		execSync(`go build -ldflags="-s -w" -o "${binaryPath}" main.go`, {
			stdio: "inherit",
			env: env,
			cwd: path.join(__dirname, ".."),
		});

		if (process.platform !== "win32") {
			fs.chmodSync(binaryPath, "755");
		}

		console.log("✅ Servexy built successfully!");
		return true;
	} catch (error) {
		console.log(
			"⚠️  Local build failed, will try to download from GitHub releases...",
		);
		return false;
	}
}

async function install() {
	try {
		if (fs.existsSync(path.join(__dirname, "..", "main.go"))) {
			const built = buildLocally();
			if (built) return;
		}

		await downloadBinary();
	} catch (error) {
		console.error("❌ Failed to install servexy:", error.message);
		console.error("Please check that:");
		console.error("1. You have an internet connection");
		console.error("2. The GitHub release exists for your platform");
		console.error(
			"3. Your platform is supported (Windows, macOS, Linux on x64/arm64)",
		);
		process.exit(1);
	}
}

if (require.main === module) {
	install();
}

module.exports = { install };
