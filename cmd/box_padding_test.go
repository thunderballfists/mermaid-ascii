package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/lib"
)

func TestUnicodeBoxPadding(t *testing.T) {
	for _, testCase := range []struct {
		name, options, label, expected string
	}{
		{"defaults", `{}`, "A", "┌───┐\n│   │\n│ A │\n│   │\n└───┘"},
		{"compact", `{"BoxPaddingX":1,"BoxPaddingY":0}`, "A", "┌───┐\n│ A │\n└───┘"},
		{"no padding", `{"BoxPaddingX":0,"BoxPaddingY":0}`, "A", "┌─┐\n│A│\n└─┘"},
		{"horizontal only", `{"BoxPaddingX":0}`, "A", "┌─┐\n│ │\n│A│\n│ │\n└─┘"},
		{"vertical only", `{"BoxPaddingY":0}`, "A", "┌───┐\n│ A │\n└───┘"},
		{"wide", `{"BoxPaddingX":3,"BoxPaddingY":0}`, "A", "┌───────┐\n│   A   │\n└───────┘"},
		{"tall", `{"BoxPaddingX":0,"BoxPaddingY":2}`, "A", "┌─┐\n│ │\n│ │\n│A│\n│ │\n│ │\n└─┘"},
		{"wide text", `{"BoxPaddingX":0,"BoxPaddingY":0}`, "東京", "┌────┐\n│東京│\n└────┘"},
		{"grapheme", `{"BoxPaddingX":0,"BoxPaddingY":0}`, "👩‍💻", "┌──┐\n│👩‍💻│\n└──┘"},
		{"combining", `{"BoxPaddingX":0,"BoxPaddingY":0}`, "e\u0301", "┌─┐\n│e\u0301│\n└─┘"},
		{"zero width", `{"BoxPaddingX":0,"BoxPaddingY":0}`, "\u0301", "┌─┐\n│\u0301 │\n└─┘"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var options lib.Options
			if err := json.Unmarshal([]byte(testCase.options), &options); err != nil {
				t.Fatal(err)
			}
			output, err := lib.Render("graph LR\nA["+testCase.label+"]", options)
			if err != nil {
				t.Fatal(err)
			}
			if output != testCase.expected {
				t.Fatalf("unexpected box:\n%s\nwant:\n%s", output, testCase.expected)
			}
		})
	}
}

func TestBoxPaddingIgnoredInASCII(t *testing.T) {
	const source = "graph LR\nA-->|label|B\nB-->C"
	expected, err := lib.Render(source, lib.Options{UseASCII: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, values := range []string{
		`{"UseASCII":true,"BoxPaddingX":0,"BoxPaddingY":0}`,
		`{"UseASCII":true,"BoxPaddingX":-1,"BoxPaddingY":-1}`,
	} {
		var options lib.Options
		if err := json.Unmarshal([]byte(values), &options); err != nil {
			t.Fatal(err)
		}
		output, err := lib.Render(source, options)
		if err != nil || output != expected {
			t.Fatalf("ASCII changed with %s: error=%v\n%s", values, err, output)
		}
	}
}

func TestNegativeUnicodeBoxPadding(t *testing.T) {
	for _, field := range []string{"BoxPaddingX", "BoxPaddingY"} {
		t.Run(field, func(t *testing.T) {
			var options lib.Options
			if err := json.Unmarshal([]byte(`{"`+field+`":-1}`), &options); err != nil {
				t.Fatal(err)
			}
			output, err := lib.Render("graph LR\nA", options)
			if err == nil || output != "" || !strings.Contains(err.Error(), "non-negative") {
				t.Fatalf("expected validation error and no output, got %q, %v", output, err)
			}
		})
	}
}

func TestBoxPaddingPreservesLegacyDefault(t *testing.T) {
	originalPadding := boxBorderPadding
	t.Cleanup(func() { boxBorderPadding = originalPadding })
	boxBorderPadding = 2
	for _, ascii := range []bool{false, true} {
		output, err := lib.Render("graph LR\nA", lib.Options{UseASCII: ascii})
		if err != nil {
			t.Fatal(err)
		}
		if rows := strings.Split(output, "\n"); len(rows) != 7 || len([]rune(rows[0])) != 7 {
			t.Fatalf("legacy border padding ignored (ascii=%v):\n%s", ascii, output)
		}
	}
	zero := 0
	if _, err := lib.Render("graph LR\nA", lib.Options{BoxPaddingY: &zero}); err != nil {
		t.Fatal(err)
	}
	output, err := lib.Render("graph LR\nA", lib.Options{})
	if err != nil || len(strings.Split(output, "\n")) != 7 {
		t.Fatalf("box padding leaked between renders: %v\n%s", err, output)
	}
}

func renderBoxPaddingGraph(t *testing.T, source string, gap, boxX, boxY int, heads HeadStyle) (graph, string) {
	t.Helper()
	originalASCII, originalHeads, originalDirection := useAscii, headStyle, graphDirection
	defer func() { useAscii, headStyle, graphDirection = originalASCII, originalHeads, originalDirection }()
	useAscii, headStyle = false, heads
	properties, err := mermaidFileToMap(source, "cli")
	if err != nil {
		t.Fatal(err)
	}
	properties.paddingX, properties.paddingY = gap, gap
	properties.boxPaddingX, properties.boxPaddingY = &boxX, &boxY
	rendered := mkGraph(properties)
	rendered.setStyleClasses(properties)
	rendered.setSubgraphs(properties.subgraphs)
	rendered.createMapping()
	return rendered, drawingToString(rendered.draw())
}

func TestCompactBoxRoutes(t *testing.T) {
	for _, source := range []string{
		"A-->B\nB-->C",
		"A<-->B",
		"A-->C\nB-->C",
		"A-->|loop|A",
		"A-->|one|B\nA-->|two|B",
		"A-->|one|B\nB-->|two|A",
		"A[Café]-->|é|B[東京]\nB-->C[👩‍💻]",
		"A[\u0301]-->B[\u0300]",
		"A-->|a_long_edge_label|B\nB-->C",
		"subgraph Outer\nA-->B\nend\nB-->C",
		"subgraph One\nA-->B\nend\nsubgraph Two\nC-->D\nend\nA-->|cross|D",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxX := range []int{0, 1, 3} {
					for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
						t.Run(fmt.Sprintf("%s/%s/gap=%d/boxX=%d/heads=%s", source, direction, gap, boxX, headStyleString(heads)), func(t *testing.T) {
							rendered, output := renderBoxPaddingGraph(t, "graph "+direction+"\n"+source, gap, boxX, 0, heads)
							if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
								t.Fatalf("compact boxes lost information: %v\n%s", defects, output)
							}
							assertNodeInteriors(t, rendered, output)
						})
					}
				}
			}
		}
	}
}
