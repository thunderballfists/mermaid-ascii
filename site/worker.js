/* eslint-disable no-restricted-globals */
importScripts("wasm_exec.js");

const go = new Go();
let runtimeReady = false;
let runtimeInit = null;

async function ensureRuntime() {
  if (runtimeReady) {
    return;
  }
  if (!runtimeInit) {
    runtimeInit = (async () => {
      let instance;
      if (WebAssembly.instantiateStreaming) {
        try {
          const streamed = await WebAssembly.instantiateStreaming(
            fetch("main.wasm"),
            go.importObject
          );
          instance = streamed.instance;
        } catch (streamErr) {
          console.warn("instantiateStreaming failed, falling back:", streamErr);
        }
      }

      if (!instance) {
        const response = await fetch("main.wasm");
        if (!response.ok) {
          throw new Error(`HTTP ${response.status} while fetching main.wasm`);
        }
        const bytes = await response.arrayBuffer();
        const result = await WebAssembly.instantiate(bytes, go.importObject);
        instance = result.instance;
      }

      go.run(instance);
      await new Promise((resolve) => setTimeout(resolve, 0));

      const exported = globalThis.mermaidAsciiRender;
      if (typeof exported !== "function") {
        throw new Error("mermaidAsciiRender export is missing");
      }
      runtimeReady = true;
    })();
  }
  await runtimeInit;
}

ensureRuntime()
  .then(() => {
    self.postMessage({ type: "ready" });
  })
  .catch((err) => {
    self.postMessage({
      type: "fatal",
      error: err && err.message ? err.message : String(err),
    });
  });

self.addEventListener("message", async (event) => {
  const data = event.data || {};
  if (data.type !== "render") {
    return;
  }

  try {
    await ensureRuntime();
  } catch (err) {
    self.postMessage({
      type: "error",
      id: data.id,
      error: err && err.message ? err.message : String(err),
    });
    return;
  }

  try {
    const mermaid = typeof data.mermaid === "string" ? data.mermaid : "";
    const options = data.options || {};
    const payload = JSON.stringify(options);
    const response = globalThis.mermaidAsciiRender(mermaid, payload);

    if (response && typeof response === "object" && "error" in response) {
      self.postMessage({
        type: "error",
        id: data.id,
        error: response.error || "Unknown error",
      });
      return;
    }

    let text = "";
    if (typeof response === "string") {
      text = response;
    } else if (response && typeof response === "object") {
      text = response.text || "";
    }

    self.postMessage({
      type: "result",
      id: data.id,
      text,
    });
  } catch (err) {
    self.postMessage({
      type: "error",
      id: data.id,
      error: err && err.message ? err.message : String(err),
    });
  }
});
