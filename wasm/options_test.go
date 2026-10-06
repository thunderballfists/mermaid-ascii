package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDecodeOptionsEmpty(t *testing.T) {
	opts, err := decodeOptions("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.UseASCII {
		t.Fatalf("expected UseASCII to be false")
	}
	if opts.PaddingX != nil || opts.PaddingY != nil || opts.BoxPaddingX != nil || opts.BoxPaddingY != nil || opts.UniformSpacing {
		t.Fatalf("expected padding pointers to be nil, got %+v", opts)
	}
	if opts.Heads != "" {
		t.Fatalf("expected empty Heads, got %q", opts.Heads)
	}
}

func TestDecodeOptionsBoxLayout(t *testing.T) {
	opts, err := decodeOptions(`{"boxPaddingX":3,"boxPaddingY":0,"uniformSpacing":true}`)
	if err != nil {
		t.Fatal(err)
	}
	assertPtrEqual(t, opts.BoxPaddingX, 3)
	assertPtrEqual(t, opts.BoxPaddingY, 0)
	if !opts.UniformSpacing {
		t.Fatal("uniform spacing option lost")
	}
	opts, err = decodeOptions(`{"boxPaddingX":null,"boxPaddingY":null}`)
	if err != nil || opts.BoxPaddingX != nil || opts.BoxPaddingY != nil || opts.UniformSpacing {
		t.Fatalf("null must preserve defaults: %+v, %v", opts, err)
	}
}

func TestDecodeOptionsValues(t *testing.T) {
	opts, err := decodeOptions(`{"useAscii":true,"paddingX":4,"paddingY":6,"heads":"pointer"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.UseASCII {
		t.Fatalf("expected UseASCII to be true")
	}
	assertPtrEqual(t, opts.PaddingX, 4)
	assertPtrEqual(t, opts.PaddingY, 6)
	if opts.Heads != "pointer" {
		t.Fatalf("unexpected Heads: %q", opts.Heads)
	}
}

func TestDecodeOptionsHeadsVariants(t *testing.T) {
	opts, err := decodeOptions(`{"heads":"none"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Heads != "none" {
		t.Fatalf("expected heads to be none, got %q", opts.Heads)
	}
}

func TestDecodeOptionsColorMode(t *testing.T) {
	opts, err := decodeOptions(`{"colorMode":"always"}`)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(opts)
	if err != nil || !strings.Contains(string(encoded), `"ColorMode":"always"`) {
		t.Fatalf("color mode lost: %s, %v", encoded, err)
	}
}

func TestDecodeOptionsInvalidJSON(t *testing.T) {
	_, err := decodeOptions("{")
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("expected json.SyntaxError, got %T", err)
	}
}

func assertPtrEqual(t *testing.T, got *int, want int) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected pointer value %d, got nil", want)
	}
	if *got != want {
		t.Fatalf("expected %d, got %d", want, *got)
	}
}
