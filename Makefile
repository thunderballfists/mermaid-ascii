PREFIX ?= /usr/local/bin

mermaid-ascii: cmd/*.go
	go build

.PHONY: install
install: $(PREFIX)/mermaid-ascii

$(PREFIX)/mermaid-ascii: mermaid-ascii | $(PREFIX)
	install -m 755 $< $@

.PHONY: clean
clean:
	$(RM) mermaid-ascii

.PHONY: uninstall
uninstall:
	$(RM) $(PREFIX)/mermaid-ascii

dev:
	air

# --------
# WebAssembly
# --------
wasm:
	tinygo build -o dist/main.wasm -target wasm ./wasm
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" dist/
	cp -r site/* dist/ || true
	touch dist/.nojekyll

serve:
	cd dist && python3 -m http.server 5173
test:
	go test ./...
