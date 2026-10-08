#!/usr/bin/env bun

"use strict";

const assert = require("node:assert/strict");
const { spawn, spawnSync } = require("node:child_process");
const { once } = require("node:events");
const { closeSync, ftruncateSync, mkdtempSync, openSync, rmSync } = require("node:fs");
const { createConnection, createServer } = require("node:net");
const { tmpdir } = require("node:os");
const { join } = require("node:path");
const { setTimeout: delay } = require("node:timers/promises");
const packageMetadata = require("../package.json");

const launcherPath = join(__dirname, "..", "bin", "servexy");
const version = spawnSync("node", [launcherPath, "--version"], { encoding: "utf8" });
assert.ifError(version.error);
assert.equal(version.status, 0, version.stderr);
assert.equal(version.stdout.trim(), `servexy ${packageMetadata.version}`);

async function availablePort() {
	const listener = createServer();
	listener.listen(0, "127.0.0.1");
	await once(listener, "listening");
	const port = listener.address().port;
	await new Promise((resolve, reject) => listener.close((error) => error ? reject(error) : resolve()));
	return port;
}

function connect(port) {
	return new Promise((resolve, reject) => {
		const socket = createConnection({ host: "127.0.0.1", port });
		socket.once("error", reject);
		socket.setTimeout(200, () => socket.destroy(new Error("connection timed out")));
		socket.once("connect", () => {
			socket.setTimeout(0);
			resolve(socket);
		});
	});
}

async function testInterrupts(directory, mode) {
	const port = await availablePort();
	const launcher = spawn("node", [launcherPath, directory, "--port", String(port)], {
		detached: true,
		stdio: ["ignore", "pipe", "pipe"],
	});
	let output = "";
	launcher.stdout.on("data", (chunk) => { output += chunk; });
	launcher.stderr.on("data", (chunk) => { output += chunk; });
	const exited = once(launcher, "exit");
	let socket;
	let nativePid;
	let timer;
	let passed = false;
	try {
		for (let attempt = 0; attempt < 100; attempt++) {
			try {
				socket = await connect(port);
				break;
			} catch {
				assert.equal(launcher.exitCode, null, output);
				await delay(30);
			}
		}
		assert.ok(socket, `launcher did not start: ${output}`);
		socket.on("error", () => {});

		const children = spawnSync("pgrep", ["-P", String(launcher.pid)], { encoding: "utf8" });
		assert.ifError(children.error);
		assert.equal(children.status, 0, "could not find the launcher's native child");
		nativePid = Number(children.stdout.trim());
		assert.ok(Number.isInteger(nativePid) && nativePid > 0);

		// Pause a large transfer so graceful shutdown has an active request to drain.
		const response = once(socket, "data");
		socket.once("data", () => socket.pause());
		socket.write("GET /slow.bin HTTP/1.1\r\nHost: localhost\r\n\r\n");
		const [chunk] = await Promise.race([
			response,
			new Promise((_, reject) => { timer = setTimeout(() => reject(new Error("server did not start the transfer")), 3000); }),
		]);
		clearTimeout(timer);
		assert.match(chunk.toString(), /^HTTP\/1\.1 200 /);

		const target = mode === "terminal" ? -launcher.pid : launcher.pid;
		process.kill(target, "SIGINT");
		await delay(200);
		assert.equal(launcher.exitCode, null, `first ${mode} interrupt must drain the request: ${output}`);
		assert.equal(launcher.signalCode, null, `first ${mode} interrupt killed the launcher`);
		assert.match(output, /shutting down/);

		process.kill(target, "SIGINT");
		const [code, signal] = await Promise.race([
			exited,
			new Promise((_, reject) => { timer = setTimeout(() => reject(new Error("second interrupt did not stop the server")), 2000); }),
		]);
		assert.equal(code, 130, output);
		assert.equal(signal, null);
		passed = true;
		console.log(`Launcher ${mode} interrupts: graceful first signal, immediate second signal.`);
	} finally {
		clearTimeout(timer);
		socket?.destroy();
		if (!passed) {
			for (const pid of [-launcher.pid, nativePid]) {
				if (!pid) continue;
				try { process.kill(pid, "SIGKILL"); } catch (error) {
					if (error.code !== "ESRCH") throw error;
				}
			}
			if (launcher.exitCode === null && launcher.signalCode === null) await exited;
		}
	}
}

async function main() {
	if (process.platform === "win32") {
		console.log("Verified launcher version. Ctrl+C integration requires a Windows console.");
		return;
	}
	const directory = mkdtempSync(join(tmpdir(), "servexy-launcher-"));
	try {
		const file = openSync(join(directory, "slow.bin"), "w");
		try { ftruncateSync(file, 512 * 1024 * 1024); } finally { closeSync(file); }
		await testInterrupts(directory, "terminal");
		await testInterrupts(directory, "parent");
	} finally {
		rmSync(directory, { recursive: true, force: true });
	}
}

main().catch((error) => {
	console.error(error);
	process.exitCode = 1;
});
