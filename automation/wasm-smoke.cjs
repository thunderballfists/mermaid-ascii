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
  assert.equal((await render("graph LR\nA(Start)", { boxPaddingY: 0 })).text, "╭───────╮\n│ Start │\n╰───────╯");
  assert.equal((await render("graph LR\nA[[Run]]", { boxPaddingY: 0 })).text, "┌─┬─────┬─┐\n│ │ Run │ │\n└─┴─────┴─┘");
  assert.equal((await render("graph LR\nA([Start])", { boxPaddingY: 0 })).text, "╭───────╮\n( Start )\n╰───────╯");
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
  const branchHeads = await render("graph LR\nA<-->B\nB-->C\nD-->C", {
    paddingX: 1, paddingY: 1, boxPaddingX: 0, boxPaddingY: 0,
  });
  assert.equal(branchHeads.type, "result");
  assert.equal((branchHeads.text.match(/[►◄▲▼]/g) || []).length, 4);
  const loop = await render("graph LR\nA-->|loop|A\nA-->B\nB-->C", {
    paddingX: 1, paddingY: 1, boxPaddingX: 0, boxPaddingY: 0,
  });
  assert.equal(loop.type, "result");
  assert.ok(loop.text.includes("loop"));
  assert.equal((loop.text.match(/[►◄▲▼]/g) || []).length, 3);
  const shapedLoop = await render("graph LR\nA(Start)-->|loop|A", { boxPaddingY: 0 });
  assert.equal(shapedLoop.type, "result");
  assert.ok(shapedLoop.text.includes("Start"));
  assert.ok(shapedLoop.text.includes("loop"));
  assert.equal((shapedLoop.text.match(/[►◄▲▼]/g) || []).length, 1);
  const spacedLabel = await render("graph LR\nA-->|Café 東京 returns|B", { boxPaddingY: 0 });
  assert.equal(spacedLabel.type, "result");
  assert.ok(spacedLabel.text.includes("Café 東京 returns"));
  assert.equal((spacedLabel.text.match(/[►◄▲▼]/g) || []).length, 1);
  const doubleEdge = await render("graph LR\nA==>B", { boxPaddingY: 0 });
  assert.equal(doubleEdge.type, "result");
  assert.equal(doubleEdge.text, "┌───┐     ┌───┐\n│ A ╞════►│ B │\n└───┘     └───┘");
  const verticalDoubleEdge = await render("graph TD\nA==>B", { boxPaddingY: 0 });
  assert.equal(verticalDoubleEdge.type, "result");
  assert.ok(verticalDoubleEdge.text.includes("└─╥─┘"));
  const coloredDoubleEdge = await render("graph LR\nclassDef box stroke:#00ff00\nA:::box ==> B\nlinkStyle 0 stroke:#ff0000", { colorMode: "always", boxPaddingY: 0 });
  assert.equal(coloredDoubleEdge.type, "result");
  assert.ok(coloredDoubleEdge.text.includes("\x1b[38;2;0;255;0m╞\x1b[0m"));
  assert.ok(coloredDoubleEdge.text.includes("\x1b[38;2;255;0;0m═\x1b[0m"));
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
  const colorSource = "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA[東京]:::box -->|two words| B:::box\nlinkStyle 0 stroke:#ff0000,color:#0000ff";
  const colored = await render(colorSource, { colorMode: "always", boxPaddingY: 0 });
  assert.equal(colored.type, "result");
  assert.ok(colored.text.includes("\x1b[38;2;0;255;0m├\x1b[0m"));
  assert.ok(colored.text.includes("\x1b[38;2;255;0;0m►\x1b[0m"));
  assert.ok(colored.text.includes("\x1b[38;2;0;0;255mt\x1b[0m"));
  const plainColor = await render(colorSource, { colorMode: "never", boxPaddingY: 0 });
  assert.equal(plainColor.type, "result");
  assert.equal(colored.text.replace(/\x1b\[[0-9;]*m/g, ""), plainColor.text);
  const separateClassSource = "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA[東京] -->|two words| B\nclass A, B box;\nlinkStyle 0 stroke:#ff0000,color:#0000ff";
  const classAssigned = await render(separateClassSource, { colorMode: "always", boxPaddingY: 0 });
  assert.equal(classAssigned.type, "result");
  assert.equal(classAssigned.text, colored.text);
  const invalidColor = await render("graph LR\nA", { colorMode: "invalid" });
  assert.equal(invalidColor.type, "error");
  assert.match(invalidColor.error, /color mode/);
  assert.equal((await render(source, { useAscii: true, colorMode: "always" })).text, ascii.text);
  const examplesContext = { window: {} };
  vm.runInNewContext(fs.readFileSync(path.join(__dirname, "../site/examples.js"), "utf8"), examplesContext);
  const colorExample = examplesContext.window.MERMAID_ASCII_EXAMPLES.find(example => example.id === "colored-flow");
  const shapeExample = examplesContext.window.MERMAID_ASCII_EXAMPLES.find(example => example.id === "node-shapes");
  assert.ok(shapeExample);
  const shapeResult = await render(shapeExample.mermaid, { boxPaddingY: 0, colorMode: "never" });
  assert.equal(shapeResult.type, "result");
  assert.ok(shapeResult.text.includes("( Start ├"));
  assert.ok(shapeResult.text.includes("│ │ Run │ ├"));
  assert.ok(shapeResult.text.includes("│ Done │"));
  assert.equal((shapeResult.text.match(/►/g) || []).length, 2);
  assert.ok(colorExample);
  const exampleResult = await render(colorExample.mermaid, { colorMode: "always", boxPaddingY: 0 });
  assert.equal(exampleResult.type, "result");
  assert.ok(exampleResult.text.includes("\x1b[38;2;0;170;102m├\x1b[0m"));
  assert.ok(exampleResult.text.includes("\x1b[38;2;204;51;68m►\x1b[0m"));
  assert.ok(exampleResult.text.includes("\x1b[38;2;153;102;204mt\x1b[0m"));
  clearTimeout(timeout);
  log("Built WASM and worker: defaults, compact boxes, Unicode text, colors, validation, uniform gaps, ASCII invariance passed");
  process.exit(0);
})().catch(error => { console.error(error); process.exit(1); });
