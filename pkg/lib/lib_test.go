package lib

import (
	"errors"
	"testing"
)

func TestRenderRequiresRegistration(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		// Ensure tests remain isolated in case the default renderer changes.
		RegisterRenderer(func(string, Options) (string, error) {
			return "", ErrRendererNotRegistered
		})
	})

	if _, err := Render("graph LR\nA-->B\n", Options{}); !errors.Is(err, ErrRendererNotRegistered) {
		t.Fatalf("expected ErrRendererNotRegistered, got %v", err)
	}
}

func TestRegisterRenderer(t *testing.T) {
	t.Helper()

	stubErr := errors.New("stub")
	px, py := 2, 3
	stub := func(mermaid string, opts Options) (string, error) {
		if mermaid != "graph LR\nA-->B\n" {
			t.Fatalf("unexpected mermaid content: %q", mermaid)
		}
		if !opts.UseASCII {
			t.Fatalf("expected UseASCII to be true")
		}
		if opts.PaddingX == nil || *opts.PaddingX != px {
			t.Fatalf("unexpected PaddingX: %+v", opts.PaddingX)
		}
		if opts.PaddingY == nil || *opts.PaddingY != py {
			t.Fatalf("unexpected PaddingY: %+v", opts.PaddingY)
		}
		if opts.Heads != "pointer" {
			t.Fatalf("unexpected Heads: %q", opts.Heads)
		}
		return "ok", stubErr
	}

	prev := RegisterRenderer(stub)
	t.Cleanup(func() {
		RegisterRenderer(prev)
	})

	got, err := Render("graph LR\nA-->B\n", Options{
		UseASCII: true,
		PaddingX: &px,
		PaddingY: &py,
		Heads:    "pointer",
	})
	if got != "ok" {
		t.Fatalf("expected ok, got %q", got)
	}
	if !errors.Is(err, stubErr) {
		t.Fatalf("expected stub error, got %v", err)
	}
}
