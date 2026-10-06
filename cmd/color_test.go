package cmd

import (
	"regexp"
	"strings"
	"testing"
)

func renderHTMLColors(t *testing.T, source string) (graph, string) {
	t.Helper()
	originalASCII, originalDirection := useAscii, graphDirection
	t.Cleanup(func() { useAscii, graphDirection = originalASCII, originalDirection })
	useAscii = false
	properties, err := mermaidFileToMap(source, "html")
	if err != nil {
		t.Fatal(err)
	}
	rendered := mkGraph(properties)
	rendered.setStyleClasses(properties)
	rendered.setSubgraphs(properties.subgraphs)
	rendered.createMapping()
	return rendered, drawingToString(rendered.draw())
}

func TestUnicodeColorOwnership(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		t.Run(direction, func(t *testing.T) {
			source := "graph " + direction + "\nclassDef box color:#112233,stroke:#00ff00\nA[東京]:::box -->|two words| B:::box"
			_, plain := renderHTMLColors(t, source)
			rendered, styled := renderHTMLColors(t, source+"\nlinkStyle 0 stroke:#ff0000,color:#0000ff")
			strip := regexp.MustCompile(`<span style='color: #[0-9a-fA-F]+'>|</span>`)
			if strip.ReplaceAllString(styled, "") != strip.ReplaceAllString(plain, "") {
				t.Fatalf("colors changed glyphs or spacing:\n%s", styled)
			}
			for _, fragment := range []string{
				"<span style='color: #00ff00'>┌</span>",
				"<span style='color: #112233'>東</span>",
				"<span style='color: #0000ff'>t</span>",
			} {
				if !strings.Contains(styled, fragment) {
					t.Errorf("missing %s", fragment)
				}
			}
			for _, contentNode := range rendered.nodes {
				width, height := getDrawingSize(contentNode.drawing)
				for x := 0; x <= width; x++ {
					for y := 0; y <= height; y++ {
						if x != 0 && x != width && y != 0 && y != height {
							continue
						}
						cell := (*rendered.drawing)[contentNode.drawingCoord.x+x][contentNode.drawingCoord.y+y]
						if !strings.HasPrefix(cell, "<span style='color: #00ff00'>") {
							t.Errorf("border lost ownership: %s", cell)
						}
					}
				}
			}
			head := "►"
			if direction == "TD" {
				head = "▼"
			}
			if !strings.Contains(styled, "<span style='color: #ff0000'>"+head+"</span>") {
				t.Errorf("head lost edge color:\n%s", styled)
			}
		})
	}
}

func TestUnicodeLinkStyleSourceOrder(t *testing.T) {
	rendered, output := renderHTMLColors(t, "graph LR\nA-->B\nC-->D\nA-->C\nlinkStyle 1 stroke:#ff0000")
	destination, err := rendered.getNode("D")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(output, "<span style='color: #ff0000'>▲</span>") != 1 {
		t.Fatalf("expected one styled head:\n%s", output)
	}
	width, height := getDrawingSize(destination.drawing)
	head := (*rendered.drawing)[destination.drawingCoord.x+width/2][destination.drawingCoord.y+height+1]
	if head != "<span style='color: #ff0000'>▲</span>" {
		t.Fatalf("linkStyle used internal node order: %s", head)
	}
}

func TestUnicodeLinkStyleValidation(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	for _, directive := range []string{"linkStyle", "linkStyle x stroke:#ff0000", "linkStyle 1 stroke:#ff0000", "linkStyle 0 stroke:invalid"} {
		if _, err := mermaidFileToMap("graph LR\nA-->B\n"+directive, "html"); err == nil {
			t.Errorf("accepted invalid directive %q", directive)
		}
	}
}

func TestUnicodeLinkStyleWhitespace(t *testing.T) {
	_, output := renderHTMLColors(t, "graph LR\nA-->B\nlinkStyle\t0\tstroke:#f00")
	if !strings.Contains(output, "<span style='color: #ff0000'>►</span>") || strings.Contains(output, "linkStyle") {
		t.Fatalf("tab-separated directive was not parsed:\n%s", output)
	}
}

func TestUnicodeColorRoutes(t *testing.T) {
	strip := regexp.MustCompile(`<span style='color: #[0-9a-fA-F]+'>|</span>`)
	for _, source := range []string{
		"A<-->B\nB-->C\nD-->C",
		"A-->|one word|B\nA-.->|two words|B",
		"A-->|a loop|A\nA==>B",
		"subgraph Outer\nA-->B\nend\nB-->|cross group|C",
		"A-->B\nA-.->C\nA==>D",
	} {
		for _, direction := range []string{"LR", "TD"} {
			t.Run(direction+"/"+source, func(t *testing.T) {
				_, baseline := renderHTMLColors(t, "graph "+direction+"\n"+source)
				_, colored := renderHTMLColors(t, "graph "+direction+"\n"+source+"\nlinkStyle default stroke:#f00,color:#00f;")
				if strip.ReplaceAllString(colored, "") != baseline || !strings.Contains(colored, "color: #ff0000") {
					t.Fatalf("color changed route:\n%s", colored)
				}
			})
		}
	}
}

func TestUnicodeLinkStyleSelectors(t *testing.T) {
	rendered, _ := renderHTMLColors(t, "graph LR\nA-->B\nB-->C\nC-->D\nlinkStyle default stroke:#f00\nlinkStyle 0,2 color:#00f,stroke:#0f0")
	for _, connection := range rendered.edges {
		expectedStroke := "#00ff00"
		if connection.sourceIndex == 1 {
			expectedStroke = "#ff0000"
		}
		if connection.colors["stroke"] != expectedStroke {
			t.Errorf("edge %d: %v", connection.sourceIndex, connection.colors)
		}
	}
}

func TestUnicodeStandaloneNodeColors(t *testing.T) {
	for _, declaration := range []string{
		"classDef box color:#00f,stroke:#0f0",
		"classDef box color:#00f,stroke:#0f0;",
		"classDef\tbox color:#00f, stroke:#0f0;",
	} {
		_, output := renderHTMLColors(t, "graph LR\n"+declaration+"\nA:::box")
		if !strings.Contains(output, "<span style='color: #00ff00'>┌</span>") || !strings.Contains(output, "<span style='color: #0000ff'>A</span>") {
			t.Fatalf("standalone class lost with %q:\n%s", declaration, output)
		}
	}
}

func TestASCIIColorDirectivesKeepLegacyParsing(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = true
	properties, err := mermaidFileToMap("graph LR\nA-->B\nlinkStyle 0 stroke:#f00", "cli")
	if err != nil {
		t.Fatal(err)
	}
	_, exists := properties.data.Get("linkStyle 0 stroke:#f00")
	if len(properties.linkStyles) != 0 || !exists {
		t.Fatal("new Unicode directive changed legacy ASCII parsing")
	}
}
