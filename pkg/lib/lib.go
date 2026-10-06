package lib

import (
	"errors"
	"sync/atomic"
)

// Options configures how a Mermaid diagram should be rendered.
// Nil padding fields leave the renderer's existing defaults unchanged.
type Options struct {
	UseASCII       bool
	PaddingX       *int
	PaddingY       *int
	BoxPaddingX    *int
	BoxPaddingY    *int
	UniformSpacing bool
	Heads          string
	ColorMode      string
}

// Renderer produces a textual representation of a Mermaid diagram.
type Renderer func(string, Options) (string, error)

var (
	errRendererNotRegistered = errors.New("lib: renderer not registered")
	renderer                 atomic.Value
)

func init() {
	renderer.Store(Renderer(func(string, Options) (string, error) {
		return "", errRendererNotRegistered
	}))
}

// ErrRendererNotRegistered is returned when Render is invoked before a renderer
// has been registered.
var ErrRendererNotRegistered = errRendererNotRegistered

// RegisterRenderer installs the renderer that Render delegates to. It returns
// the previously registered renderer.
func RegisterRenderer(fn Renderer) Renderer {
	if fn == nil {
		panic("lib: RegisterRenderer called with nil Renderer")
	}
	previous := renderer.Load().(Renderer)
	renderer.Store(fn)
	return previous
}

// Render converts Mermaid syntax into text by delegating to the registered
// renderer.
func Render(mermaid string, opts Options) (string, error) {
	return renderer.Load().(Renderer)(mermaid, opts)
}
