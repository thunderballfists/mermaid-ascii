const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

function loadApp(query = "") {
  const sitePath = path.join(__dirname, "../site");
  const html = fs.readFileSync(path.join(sitePath, "index.html"), "utf8");
  const elements = new Map();
  function createNode(tagName = "") {
    const listeners = new Map();
    let content = "";
    const node = {
      tagName, children: [], style: {},
      value: "", checked: false, disabled: false, dataset: {},
      options: [{ value: "" }],
      classList: { add() {}, remove() {}, toggle() {} },
      setAttribute() {}, removeAttribute() {}, focus() {},
      addEventListener(type, callback) { listeners.set(type, callback); },
      emit(type) { return listeners.get(type)?.({}); },
      click() {},
      appendChild(child) { this.children.push(child); },
      replaceChildren(...children) {
        this.textContent = "";
        this.children = children.flatMap(child => child.nodeType === 11 ? child.children : [child]);
      },
      get textContent() { return this.children.length ? this.children.map(child => child.textContent).join("") : content; },
      set textContent(value) { content = String(value); this.children = []; },
    };
    Object.defineProperty(node, "innerHTML", { set() { throw new Error("HTML injection is forbidden"); } });
    return node;
  }
  for (const match of html.matchAll(/\bid="([^"]+)"/g)) {
    elements.set(match[1], createNode());
  }
  elements.get("heads").options = ["pointer", "open", "none"].map(value => ({ value }));
  let initialize;
  const workerListeners = new Map();
  const requests = [];
  const copied = [], downloaded = [];
  const timers = new Map();
  let nextTimer = 0;
  const window = {
    location: { href: `http://localhost/${query}` },
    history: { replaceState(_state, _title, url) { window.location.href = String(url); } },
    addEventListener() {},
  };
  const context = vm.createContext({
    window, console, Blob,
    URL: class extends URL {
      static createObjectURL(blob) { downloaded.push(blob); return "blob:test"; }
      static revokeObjectURL() {}
    },
    navigator: { clipboard: { async writeText(text) { copied.push(text); } } },
    document: {
      getElementById(id) { return elements.get(id) || null; },
      addEventListener(type, callback) { if (type === "DOMContentLoaded") initialize = callback; },
      createElement: createNode,
      createTextNode(textContent) { return { textContent }; },
      createDocumentFragment() { return Object.assign(createNode(), { nodeType: 11 }); },
      body: { appendChild() {}, removeChild() {} },
    },
    Worker: class {
      addEventListener(type, callback) { workerListeners.set(type, callback); }
      postMessage(request) { requests.push(JSON.parse(JSON.stringify(request))); }
    },
    setTimeout(callback) { timers.set(++nextTimer, callback); return nextTimer; },
    clearTimeout(id) { timers.delete(id); },
  });
  vm.runInContext(fs.readFileSync(path.join(sitePath, "examples.js"), "utf8"), context);
  vm.runInContext(fs.readFileSync(path.join(sitePath, "app.js"), "utf8"), context);
  initialize();
  workerListeners.get("message")({ data: { type: "ready" } });
  return {
    elements, window, api: window.__MERMAID_ASCII_APP__, requests, copied, downloaded,
    receive(data) { workerListeners.get("message")({ data }); },
    async settle() {
      for (const callback of timers.values()) callback();
      timers.clear();
      await new Promise(setImmediate);
    },
  };
}

test("box defaults stay at one and ASCII dispatch omits extended layout options", async () => {
  const app = loadApp();
  await app.settle();
  assert.equal(app.elements.get("box-padding-x")?.value, "1");
  assert.equal(app.elements.get("box-padding-y")?.value, "1");
  assert.equal(app.elements.get("box-padding-y").disabled, true);
  assert.equal(app.elements.get("uniform-spacing").disabled, true);
  assert.deepEqual(app.requests.at(-1).options, { useAscii: true, paddingX: 5, paddingY: 5, heads: "pointer" });
});

test("compact and uniform controls dispatch independently from gaps", async () => {
  const app = loadApp("?ext=1");
  await app.settle();
  app.elements.get("box-padding-x").value = "3";
  app.elements.get("box-padding-x").emit("input");
  const initialRequests = app.requests.length;
  await app.settle();
  assert.ok(app.requests.length > initialRequests);
  assert.equal(app.requests.at(-1).options.boxPaddingX, 3);
  app.elements.get("box-padding-y").value = "0";
  app.elements.get("box-padding-y").emit("input");
  const horizontalRequests = app.requests.length;
  await app.settle();
  assert.ok(app.requests.length > horizontalRequests);
  assert.equal(app.requests.at(-1).options.boxPaddingY, 0);
  app.elements.get("uniform-spacing").checked = true;
  app.elements.get("uniform-spacing").emit("change");
  await app.settle();
  assert.deepEqual(app.requests.at(-1).options, {
    useAscii: false, paddingX: 5, paddingY: 5, heads: "pointer",
    boxPaddingX: 3, boxPaddingY: 0, uniformSpacing: true, colorMode: "always",
  });
});

test("shared URLs retain zero box padding and uniform spacing", async () => {
  const app = loadApp("?ext=1&bx=3&by=0&u=1&x=7&y=2&unrelated=keep");
  await app.settle();
  assert.deepEqual(app.requests.at(-1).options, {
    useAscii: false, paddingX: 7, paddingY: 2, heads: "pointer",
    boxPaddingX: 3, boxPaddingY: 0, uniformSpacing: true, colorMode: "always",
  });
  app.api.syncURLFromForm();
  const params = new URL(app.window.location.href).searchParams;
  assert.equal(params.get("bx"), "3");
  assert.equal(params.get("by"), "0");
  assert.equal(params.get("u"), "1");
  assert.equal(params.get("unrelated"), "keep");
  const restored = loadApp(new URL(app.window.location.href).search);
  await restored.settle();
  assert.deepEqual(restored.requests.at(-1).options, app.requests.at(-1).options);
});

test("extended colors use safe spans while copy and download retain plain Unicode text", async () => {
  const app = loadApp("?ext=1");
  await app.settle();
  const text = "\x1b[38;2;0;255;0m├\x1b[0m\n\x1b[38;2;255;0;0m東京 <img src=x onerror=alert(1)>\x1b[0m";
  const plain = "├\n東京 <img src=x onerror=alert(1)>";
  app.receive({ type: "result", id: app.requests.at(-1).id, text });
  const output = app.elements.get("render-output");
  assert.equal(output.textContent, plain);
  const spans = output.children.filter(child => child.tagName === "span");
  assert.deepEqual(spans.map(span => span.style.color), ["rgb(0, 255, 0)", "rgb(255, 0, 0)"]);
  assert.equal(spans[1].textContent, "東京 <img src=x onerror=alert(1)>");
  assert.equal(app.requests.at(-1).options.colorMode, "always");
  await app.elements.get("copy-btn").emit("click");
  assert.deepEqual(app.copied, [plain]);
  app.elements.get("download-btn").emit("click");
  assert.equal(await app.downloaded[0].text(), plain);
  app.receive({ type: "error", id: app.requests.at(-1).id, error: text });
  assert.equal(output.textContent, text);
  assert.equal(output.children.length, 0);
  app.elements.get("extended-toggle").checked = false;
  app.elements.get("extended-toggle").emit("change");
  await app.settle();
  assert.equal(app.requests.at(-1).options.colorMode, undefined);
  app.receive({ type: "result", id: app.requests.at(-1).id, text });
  assert.equal(output.textContent, text);
});

test("a mode toggle keeps the pending response's color mode until the next dispatch", async () => {
  const app = loadApp("?ext=1");
  await app.settle();
  const previousId = app.requests.at(-1).id;
  const text = "\x1b[38;2;255;0;0m東京\x1b[0m";
  app.elements.get("extended-toggle").checked = false;
  app.elements.get("extended-toggle").emit("change");
  app.receive({ type: "result", id: previousId, text });
  assert.equal(app.elements.get("render-output").textContent, "東京");
  await app.elements.get("copy-btn").emit("click");
  assert.deepEqual(app.copied, ["東京"]);
  await app.settle();
  app.receive({ type: "result", id: app.requests.at(-1).id, text: "ASCII" });
  app.receive({ type: "result", id: previousId, text });
  assert.equal(app.elements.get("render-output").textContent, "ASCII");
});

test("the colored demo example dispatches its actual source in extended mode", async () => {
  const app = loadApp("?ext=1");
  await app.settle();
  const index = app.window.MERMAID_ASCII_EXAMPLES.findIndex(example => example.id === "colored-flow");
  assert.ok(index >= 0);
  app.elements.get("example-select").value = String(index);
  app.elements.get("example-select").emit("change");
  await app.settle();
  assert.equal(app.requests.at(-1).mermaid, app.window.MERMAID_ASCII_EXAMPLES[index].mermaid);
  assert.equal(app.requests.at(-1).options.colorMode, "always");
});

test("mode toggles preserve settings and Clear resets them", async () => {
  const app = loadApp("?ext=1&bx=3&by=0&u=1");
  await app.settle();
  const extended = app.elements.get("extended-toggle");
  extended.checked = false;
  extended.emit("change");
  await app.settle();
  assert.equal(app.elements.get("box-padding-x").disabled, true);
  assert.equal(app.requests.at(-1).options.boxPaddingX, undefined);
  assert.equal(app.requests.at(-1).options.uniformSpacing, undefined);
  extended.checked = true;
  extended.emit("change");
  await app.settle();
  assert.equal(app.requests.at(-1).options.boxPaddingY, 0);
  assert.equal(app.requests.at(-1).options.uniformSpacing, true);
  app.elements.get("clear-btn").emit("click");
  await app.settle();
  assert.equal(app.elements.get("box-padding-x").value, "1");
  assert.equal(app.elements.get("box-padding-y").value, "1");
  assert.equal(app.elements.get("uniform-spacing").checked, false);
  assert.equal(app.elements.get("box-padding-y").disabled, true);
  const params = new URL(app.window.location.href).searchParams;
  for (const key of ["bx", "by", "u", "ext"]) assert.equal(params.has(key), false);
});

test("invalid URL numbers and empty box fields preserve renderer defaults", async () => {
  const app = loadApp("?ext=1&bx=invalid&by=invalid");
  await app.settle();
  assert.equal(app.requests.at(-1).options.boxPaddingX, 1);
  assert.equal(app.requests.at(-1).options.boxPaddingY, 1);
  for (const id of ["box-padding-x", "box-padding-y"]) {
    app.elements.get(id).value = "";
    app.elements.get(id).emit("input");
  }
  await app.settle();
  assert.equal(app.requests.at(-1).options.boxPaddingX, undefined);
  assert.equal(app.requests.at(-1).options.boxPaddingY, undefined);
});
