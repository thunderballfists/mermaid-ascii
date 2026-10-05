PREFIX ?= /usr/local/bin
GOCACHE ?= $(PWD)/.cache/go-build
GOMODCACHE ?= $(HOME)/go/pkg/mod
TINYGO_CACHE ?= $(PWD)/.cache/tinygo
TINYGOROOT := $(shell tinygo env TINYGOROOT)

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
.PHONY: wasm
wasm:
	mkdir -p dist $(GOCACHE) $(TINYGO_CACHE)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) TINYGO_CACHE=$(TINYGO_CACHE) tinygo build -o dist/main.wasm -target wasm ./wasm
	cp "$(TINYGOROOT)/targets/wasm_exec.js" dist/
	cp -r site/* dist/ || true
	touch dist/.nojekyll

.PHONY: serve
serve: wasm
	cd dist && python3 -m http.server 5173

.PHONY: test
test:
	mkdir -p $(GOCACHE)
	GOCACHE=$(GOCACHE) go test ./...
	node --test automation/site.test.cjs
