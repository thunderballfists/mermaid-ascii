package cmd

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/lib"
)

func TestUnicodeColorModes(t *testing.T) {
	const source = "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA[東京]:::box -->|two words| B:::box\nlinkStyle 0 stroke:#ff0000,color:#0000ff"
	strip := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	for _, boxY := range []int{0, 1} {
		for _, uniform := range []bool{false, true} {
			var options lib.Options
			if err := json.Unmarshal([]byte(`{"ColorMode":"always","BoxPaddingY":0}`), &options); err != nil {
				t.Fatal(err)
			}
			options.BoxPaddingY, options.UniformSpacing = &boxY, uniform
			colored, err := lib.Render(source, options)
			if err != nil {
				t.Fatal(err)
			}
			for _, fragment := range []string{
				"\x1b[38;2;0;255;0m├\x1b[0m",
				"\x1b[38;2;255;0;0m►\x1b[0m",
				"\x1b[38;2;0;0;255mt\x1b[0m",
				"\x1b[38;2;17;34;51m東\x1b[0m",
			} {
				if !strings.Contains(colored, fragment) {
					t.Errorf("missing colored cell %q", fragment)
				}
			}
			if err := json.Unmarshal([]byte(`{"ColorMode":"never"}`), &options); err != nil {
				t.Fatal(err)
			}
			plain, err := lib.Render(source, options)
			if err != nil || strings.Contains(plain, "\x1b") || strip.ReplaceAllString(colored, "") != plain {
				t.Fatalf("plain output changed: %v\n%s", err, plain)
			}
		}
	}
}

func TestColorModeValidationAndASCIIGuard(t *testing.T) {
	var options lib.Options
	if err := json.Unmarshal([]byte(`{"ColorMode":"invalid"}`), &options); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Render("graph LR\nA-->B", options); err == nil {
		t.Fatal("invalid color mode accepted")
	}
	options.UseASCII = true
	actual, err := lib.Render("graph LR\nA-->B", options)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := lib.Render("graph LR\nA-->B", lib.Options{UseASCII: true})
	if err != nil || actual != expected {
		t.Fatalf("color mode changed ASCII: %v\n%s", err, actual)
	}
}
