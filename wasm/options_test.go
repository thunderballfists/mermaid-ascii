package main

import (
	"encoding/json"
	"errors"
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
	if opts.PaddingX != nil || opts.PaddingY != nil {
		t.Fatalf("expected padding pointers to be nil, got %+v", opts)
	}
	if opts.Heads != "" {
		t.Fatalf("expected empty Heads, got %q", opts.Heads)
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
