const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

function loadApp(query = "") {
  const sitePath = path.join(__dirname, "../site");
  const html = fs.readFileSync(path.join(sitePath, "index.html"), "utf8");
  const elements = new Map();
  for (const match of html.matchAll(/\bid="([^"]+)"/g)) {
    const listeners = new Map();
    elements.set(match[1], {
      value: "", checked: false, disabled: false, textContent: "", dataset: {},
      options: [{ value: "" }],
      classList: { add() {}, remove() {}, toggle() {} },
      setAttribute() {}, removeAttribute() {}, focus() {},
      addEventListener(type, callback) { listeners.set(type, callback); },
      emit(type) { listeners.get(type)?.({}); },
    });
  }
  elements.get("heads").options = ["pointer", "open", "none"].map(value => ({ value }));
  let initialize;
  const workerListeners = new Map();
  const requests = [];
  const timers = new Map();
  let nextTimer = 0;
  const window = {
    location: { href: `http://localhost/${query}` },
    history: { replaceState(_state, _title, url) { window.location.href = String(url); } },
    addEventListener() {},
  };
  const context = vm.createContext({
    window, URL, console,
    document: {
      getElementById(id) { return elements.get(id) || null; },
      addEventListener(type, callback) { if (type === "DOMContentLoaded") initialize = callback; },
    },
    Worker: class {
      addEventListener(type, callback) { workerListeners.set(type, callback); }
      postMessage(request) { requests.push(JSON.parse(JSON.stringify(request))); }
    },
    setTimeout(callback) { timers.set(++nextTimer, callback); return nextTimer; },
    clearTimeout(id) { timers.delete(id); },
  });
  vm.runInContext(fs.readFileSync(path.join(sitePath, "app.js"), "utf8"), context);
  initialize();
  workerListeners.get("message")({ data: { type: "ready" } });
  return {
    elements, window, api: window.__MERMAID_ASCII_APP__, requests,
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
    boxPaddingX: 3, boxPaddingY: 0, uniformSpacing: true,
  });
});

test("shared URLs retain zero box padding and uniform spacing", async () => {
  const app = loadApp("?ext=1&bx=3&by=0&u=1&x=7&y=2&unrelated=keep");
  await app.settle();
  assert.deepEqual(app.requests.at(-1).options, {
    useAscii: false, paddingX: 7, paddingY: 2, heads: "pointer",
    boxPaddingX: 3, boxPaddingY: 0, uniformSpacing: true,
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
