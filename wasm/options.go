package main

import (
	"encoding/json"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/lib"
)

type wasmOptions struct {
	UseASCII bool   `json:"useAscii"`
	PaddingX *int   `json:"paddingX"`
	PaddingY *int   `json:"paddingY"`
	Heads    string `json:"heads"`
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
		UseASCII: intermediate.UseASCII,
		PaddingX: intermediate.PaddingX,
		PaddingY: intermediate.PaddingY,
		Heads:    intermediate.Heads,
	}, nil
}
