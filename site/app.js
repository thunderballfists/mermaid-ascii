(function () {
  const elements = {
    input: document.getElementById("mermaid-input"),
    extended: document.getElementById("extended-toggle"),
    paddingX: document.getElementById("padding-x"),
    paddingY: document.getElementById("padding-y"),
    heads: document.getElementById("heads"),
    renderBtn: document.getElementById("render-btn"),
    clearBtn: document.getElementById("clear-btn"),
    copyBtn: document.getElementById("copy-btn"),
    downloadBtn: document.getElementById("download-btn"),
    output: document.getElementById("render-output"),
    status: document.getElementById("status-message"),
  };

  const SAMPLE_MERMAID = "graph LR\nA-->B";
  const state = {
    go: null,
    renderFn: null,
    ready: false,
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
  }

  function clearAll() {
    elements.input.value = "";
    updateOutput("", false);
    setStatus("Cleared.");
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
  }

  document.addEventListener("DOMContentLoaded", () => {
    elements.input.value = SAMPLE_MERMAID;
    elements.paddingX.value = "5";
    elements.paddingY.value = "5";
    if (elements.heads) {
      elements.heads.value = "pointer";
    }
    attachEvents();
    loadWasm();
  });
})();
