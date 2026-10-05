const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

const artifactPath = path.resolve(process.argv[2] || path.join(__dirname, "../dist"));
if (!globalThis.crypto) globalThis.crypto = require("node:crypto").webcrypto;
const log = console.log;
console.log = () => {};
globalThis.self = globalThis;
globalThis.importScripts = filename => require(path.join(artifactPath, filename));
globalThis.fetch = async filename => new Response(fs.readFileSync(path.join(artifactPath, filename)), {
  headers: { "Content-Type": "application/wasm" },
});
let onMessage;
let resolveReady, rejectReady;
const ready = new Promise((resolve, reject) => { resolveReady = resolve; rejectReady = reject; });
const pending = new Map();
let requestId = 0;
self.addEventListener = (type, callback) => { if (type === "message") onMessage = callback; };
self.postMessage = message => {
  if (message.type === "ready") resolveReady();
  else if (message.type === "fatal") rejectReady(new Error(message.error));
  else pending.get(message.id)?.(message);
};
const timeout = setTimeout(() => { console.error("WASM smoke timed out"); process.exit(1); }, 15000);
vm.runInThisContext(fs.readFileSync(path.join(artifactPath, "worker.js"), "utf8"));

async function render(source, options = {}) {
  const id = ++requestId;
  const response = new Promise(resolve => pending.set(id, resolve));
  await onMessage({ data: { type: "render", id, mermaid: source, options } });
  const result = await response;
  pending.delete(id);
  return result;
}

(async () => {
  await ready;
  const defaultBox = "┌───┐\n│   │\n│ A │\n│   │\n└───┘";
  assert.equal((await render("graph LR\nA")).text, defaultBox);
  assert.equal((await render("graph LR\nA", { boxPaddingY: 0 })).text, "┌───┐\n│ A │\n└───┘");
  assert.equal((await render("graph LR\nA", { boxPaddingX: 0, boxPaddingY: 0 })).text, "┌─┐\n│A│\n└─┘");
  assert.equal((await render("graph LR\nA[東京]", { boxPaddingX: 0, boxPaddingY: 0 })).text, "┌────┐\n│東京│\n└────┘");
  assert.equal((await render("graph LR\nA", { boxPaddingX: 3, boxPaddingY: 0 })).text, "┌───────┐\n│   A   │\n└───────┘");
  const invalid = await render("graph LR\nA", { boxPaddingY: -1 });
  assert.equal(invalid.type, "error");
  assert.match(invalid.error, /non-negative/);
  assert.equal((await render("graph LR\nA")).text, defaultBox);
  const source = "graph LR\nA-->|long_edge_label|B\nB-->C";
  const branch = await render("graph LR\nA<-->B\nB-->C\nD-->C", {
    paddingX: 1, paddingY: 1, boxPaddingX: 0, boxPaddingY: 0, heads: "none",
  });
  assert.equal(branch.type, "result");
  assert.ok(branch.text.includes("│D│"));
  const uniform = await render(source, { paddingX: 1, boxPaddingY: 0, uniformSpacing: true });
  assert.equal(uniform.type, "result");
  assert.ok(uniform.text.includes("long_edge_label"));
  const top = [...uniform.text.split("\n")[0]];
  const starts = [], ends = [];
  top.forEach((cell, index) => { if (cell === "┌") starts.push(index); if (cell === "┐") ends.push(index); });
  assert.equal(starts.length, 3);
  assert.equal(starts[1] - ends[0], starts[2] - ends[1]);
  const ascii = await render(source, { useAscii: true });
  const asciiWithOptions = await render(source, { useAscii: true, boxPaddingX: -1, boxPaddingY: 0, uniformSpacing: true });
  assert.equal(asciiWithOptions.type, "result");
  assert.equal(asciiWithOptions.text, ascii.text);
  clearTimeout(timeout);
  log("Built WASM and worker: defaults, compact boxes, Unicode text, validation, uniform gaps, ASCII invariance passed");
  process.exit(0);
})().catch(error => { console.error(error); process.exit(1); });
