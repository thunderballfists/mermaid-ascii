(function () {
  const elements = {
    input: document.getElementById("mermaid-input"),
    extended: document.getElementById("extended-toggle"),
    paddingX: document.getElementById("padding-x"),
    paddingY: document.getElementById("padding-y"),
    exampleSelect: document.getElementById("example-select"),
    heads: document.getElementById("heads"),
    renderBtn: document.getElementById("render-btn"),
    clearBtn: document.getElementById("clear-btn"),
    copyBtn: document.getElementById("copy-btn"),
    downloadBtn: document.getElementById("download-btn"),
    output: document.getElementById("render-output"),
    status: document.getElementById("status-message"),
  };

  const SAMPLE_MERMAID = "graph LR\nA-->B";
  const DEFAULT_PADDING = 5;
  const state = {
    go: null,
    renderFn: null,
    ready: false,
  };

  const URL_KEYS = {
    mermaid: "m",
    paddingX: "x",
    paddingY: "y",
    extended: "ext",
    heads: "heads",
  };

  function setStatus(message, tone) {
    elements.status.textContent = message || "";
    elements.status.classList.remove("status--error", "status--success");
    if (tone === "error") {
      elements.status.classList.add("status--error");
    } else if (tone === "success") {
      elements.status.classList.add("status--success");
    }
  }

  function updateOutput(text, isError) {
    elements.output.textContent = text || "";
    elements.output.classList.toggle("output--error", Boolean(isError));
    const hasText = Boolean(text) && !isError;
    elements.copyBtn.disabled = !hasText;
    elements.downloadBtn.disabled = !hasText;
  }

  function getOptions() {
    const opts = {
      useAscii: !elements.extended.checked,
    };
    const paddingX = parseInt(elements.paddingX.value, 10);
    if (!Number.isNaN(paddingX)) {
      opts.paddingX = paddingX;
    }
    const paddingY = parseInt(elements.paddingY.value, 10);
    if (!Number.isNaN(paddingY)) {
      opts.paddingY = paddingY;
    }
    if (elements.heads && elements.heads.value) {
      opts.heads = elements.heads.value;
    }
    return opts;
  }

  function readStateFromURL() {
    const params = new URL(window.location.href).searchParams;
    const parsed = {};
    if (params.has(URL_KEYS.mermaid)) {
      parsed.mermaid = params.get(URL_KEYS.mermaid) || "";
    }
    if (params.has(URL_KEYS.paddingX)) {
      const parsedX = parseInt(params.get(URL_KEYS.paddingX), 10);
      if (!Number.isNaN(parsedX)) {
        parsed.paddingX = parsedX;
      }
    }
    if (params.has(URL_KEYS.paddingY)) {
      const parsedY = parseInt(params.get(URL_KEYS.paddingY), 10);
      if (!Number.isNaN(parsedY)) {
        parsed.paddingY = parsedY;
      }
    }
    if (params.has(URL_KEYS.extended)) {
      const raw = params.get(URL_KEYS.extended);
      parsed.useExtended = raw === "1" || raw === "true";
    }
    if (params.has(URL_KEYS.heads)) {
      parsed.heads = params.get(URL_KEYS.heads) || "";
    }
    return parsed;
  }

  function applyStateToForm(parsed) {
    if (typeof parsed.mermaid === "string") {
      elements.input.value = parsed.mermaid;
    }
    if (typeof parsed.paddingX === "number" && !Number.isNaN(parsed.paddingX)) {
      elements.paddingX.value = String(parsed.paddingX);
    }
    if (typeof parsed.paddingY === "number" && !Number.isNaN(parsed.paddingY)) {
      elements.paddingY.value = String(parsed.paddingY);
    }
    if (typeof parsed.useExtended === "boolean") {
      elements.extended.checked = parsed.useExtended;
    }
    if (parsed.heads && elements.heads) {
      const matchingOption = Array.from(elements.heads.options || []).some(
        (opt) => opt.value === parsed.heads
      );
      if (matchingOption) {
        elements.heads.value = parsed.heads;
      }
    }
  }

  function currentStateSnapshot() {
    const paddingX = parseInt(elements.paddingX.value, 10);
    const paddingY = parseInt(elements.paddingY.value, 10);
    const heads = elements.heads ? elements.heads.value : "pointer";
    return {
      mermaid: elements.input.value || "",
      paddingX: Number.isNaN(paddingX) ? null : paddingX,
      paddingY: Number.isNaN(paddingY) ? null : paddingY,
      useExtended: Boolean(elements.extended.checked),
      heads,
    };
  }

  function syncURLFromForm() {
    if (typeof window === "undefined" || !window.history || !window.location) {
      return;
    }
    const url = new URL(window.location.href);
    const params = url.searchParams;
    const snapshot = currentStateSnapshot();

    if (snapshot.mermaid) {
      params.set(URL_KEYS.mermaid, snapshot.mermaid);
    } else {
      params.delete(URL_KEYS.mermaid);
    }

    if (snapshot.paddingX !== null && snapshot.paddingX !== DEFAULT_PADDING) {
      params.set(URL_KEYS.paddingX, String(snapshot.paddingX));
    } else {
      params.delete(URL_KEYS.paddingX);
    }

    if (snapshot.paddingY !== null && snapshot.paddingY !== DEFAULT_PADDING) {
      params.set(URL_KEYS.paddingY, String(snapshot.paddingY));
    } else {
      params.delete(URL_KEYS.paddingY);
    }

    if (snapshot.useExtended) {
      params.set(URL_KEYS.extended, "1");
    } else {
      params.delete(URL_KEYS.extended);
    }

    if (snapshot.heads && snapshot.heads !== "pointer") {
      params.set(URL_KEYS.heads, snapshot.heads);
    } else {
      params.delete(URL_KEYS.heads);
    }

    window.history.replaceState(null, "", url);
  }

  function populateExamples() {
    if (!elements.exampleSelect) {
      return;
    }
    const examples = Array.isArray(window.MERMAID_ASCII_EXAMPLES)
      ? window.MERMAID_ASCII_EXAMPLES
      : [];

    while (elements.exampleSelect.options.length > 1) {
      elements.exampleSelect.remove(1);
    }

    examples.forEach((example, index) => {
      const option = document.createElement("option");
      option.value = String(index);
      option.textContent = example.name;
      elements.exampleSelect.appendChild(option);
    });
  }

  async function loadWasm() {
    setStatus("Loading renderer…");
    try {
      state.go = new Go();
      let instance;

      if (WebAssembly.instantiateStreaming) {
        try {
          const streamed = await WebAssembly.instantiateStreaming(
            fetch("main.wasm"),
            state.go.importObject
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
        const result = await WebAssembly.instantiate(bytes, state.go.importObject);
        instance = result.instance;
      }

      state.go.run(instance);
      await new Promise((resolve) => setTimeout(resolve, 0));

      const exported = globalThis.mermaidAsciiRender;
      if (typeof exported !== "function") {
        throw new Error("mermaidAsciiRender export is missing");
      }

      state.renderFn = exported;
      state.ready = true;
      elements.renderBtn.disabled = false;
      setStatus("Renderer ready.", "success");
      await render();
    } catch (err) {
      console.error(err);
      elements.renderBtn.disabled = true;
      setStatus(`Failed to load renderer: ${err.message}`, "error");
    }
  }

  async function render() {
    if (!state.ready || typeof state.renderFn !== "function") {
      return;
    }
    const mermaid = elements.input.value;
    const payload = JSON.stringify(getOptions());
    const response = state.renderFn(mermaid, payload);

    if (response && typeof response === "object" && "error" in response) {
      updateOutput(response.error || "Unknown error", true);
      setStatus("Render failed.", "error");
      syncURLFromForm();
      return;
    }

    let text = "";
    if (typeof response === "string") {
      text = response;
    } else if (response && typeof response === "object") {
      text = response.text || "";
    }

    updateOutput(text, false);
    setStatus("Render complete.", "success");
    syncURLFromForm();
  }

  function clearAll() {
    elements.input.value = "";
    elements.extended.checked = false;
    if (elements.heads) {
      elements.heads.value = "pointer";
    }
    elements.paddingX.value = String(DEFAULT_PADDING);
    elements.paddingY.value = String(DEFAULT_PADDING);
    if (elements.exampleSelect) {
      elements.exampleSelect.value = "";
    }
    updateOutput("", false);
    setStatus("Cleared.");
    syncURLFromForm();
  }

  async function copyOutput() {
    const text = elements.output.textContent || "";
    if (!text) {
      return;
    }
    try {
      await navigator.clipboard.writeText(text);
      setStatus("Copied to clipboard.", "success");
    } catch (err) {
      console.error("Copy failed", err);
      setStatus("Copy failed.", "error");
    }
  }

  function downloadOutput() {
    const text = elements.output.textContent || "";
    if (!text) {
      return;
    }
    const blob = new Blob([text], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "diagram.txt";
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
    setStatus("Download started.", "success");
  }

  function attachEvents() {
    elements.renderBtn.addEventListener("click", render);
    elements.clearBtn.addEventListener("click", clearAll);
    elements.copyBtn.addEventListener("click", copyOutput);
    elements.downloadBtn.addEventListener("click", downloadOutput);
    if (elements.exampleSelect) {
      elements.exampleSelect.addEventListener("change", () => {
        const selected = elements.exampleSelect.value;
        const examples = Array.isArray(window.MERMAID_ASCII_EXAMPLES)
          ? window.MERMAID_ASCII_EXAMPLES
          : [];
        const idx = parseInt(selected, 10);
        if (!Number.isNaN(idx) && examples[idx]) {
          elements.input.value = examples[idx].mermaid;
          render();
        }
      });
    }
    [elements.extended, elements.paddingX, elements.paddingY, elements.heads]
      .filter(Boolean)
      .forEach((el) => {
        el.addEventListener("change", () => {
          syncURLFromForm();
        });
      });
  }

  document.addEventListener("DOMContentLoaded", () => {
    elements.input.value = SAMPLE_MERMAID;
    elements.paddingX.value = String(DEFAULT_PADDING);
    elements.paddingY.value = String(DEFAULT_PADDING);
    if (elements.heads) {
      elements.heads.value = "pointer";
    }
    populateExamples();
    const initialState = readStateFromURL();
    if (Object.keys(initialState).length > 0) {
      applyStateToForm(initialState);
    }
    attachEvents();
    loadWasm();
  });

  if (typeof window !== "undefined") {
    window.__MERMAID_ASCII_APP__ = {
      readStateFromURL,
      applyStateToForm,
      currentStateSnapshot,
      syncURLFromForm,
      populateExamples,
    };
  }
})();
