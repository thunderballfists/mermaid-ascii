package main

import (
	"encoding/json"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/lib"
)

type wasmOptions struct {
	UseASCII       bool   `json:"useAscii"`
	PaddingX       *int   `json:"paddingX"`
	PaddingY       *int   `json:"paddingY"`
	BoxPaddingX    *int   `json:"boxPaddingX"`
	BoxPaddingY    *int   `json:"boxPaddingY"`
	UniformSpacing bool   `json:"uniformSpacing"`
	Heads          string `json:"heads"`
	ColorMode      string `json:"colorMode"`
}

func decodeOptions(optsJSON string) (lib.Options, error) {
	if strings.TrimSpace(optsJSON) == "" {
		return lib.Options{}, nil
	}

	var intermediate wasmOptions
	if err := json.Unmarshal([]byte(optsJSON), &intermediate); err != nil {
		return lib.Options{}, err
	}

	return lib.Options{
		UseASCII:       intermediate.UseASCII,
		PaddingX:       intermediate.PaddingX,
		PaddingY:       intermediate.PaddingY,
		BoxPaddingX:    intermediate.BoxPaddingX,
		BoxPaddingY:    intermediate.BoxPaddingY,
		UniformSpacing: intermediate.UniformSpacing,
		Heads:          intermediate.Heads,
		ColorMode:      intermediate.ColorMode,
	}, nil
}
