//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/lib"

	_ "github.com/AlexanderGrooff/mermaid-ascii/cmd"
)

var renderCallback js.Func

func main() {
	renderCallback = js.FuncOf(mermaidAsciiRender)
	js.Global().Set("mermaidAsciiRender", renderCallback)
	select {}
}

func mermaidAsciiRender(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return map[string]any{"error": "missing mermaid input"}
	}

	mermaid := args[0].String()
	opts, err := parseOptionsArg(args)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	result, err := lib.Render(mermaid, opts)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	return map[string]any{"text": result}
}

func parseOptionsArg(args []js.Value) (lib.Options, error) {
	if len(args) < 2 {
		return lib.Options{}, nil
	}
	raw := args[1]
	if !raw.Truthy() {
		return lib.Options{}, nil
	}

	return decodeOptions(raw.String())
}
