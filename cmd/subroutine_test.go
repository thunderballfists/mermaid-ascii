package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestSubroutineNodeRendering(t *testing.T) {
	for _, testCase := range []struct {
		label      string
		boxX, boxY int
		want       string
	}{
		{"A", 0, 0, "┌─┬─┬─┐\n│ │A│ │\n└─┴─┴─┘"},
		{"Run", 1, 0, "┌─┬─────┬─┐\n│ │ Run │ │\n└─┴─────┴─┘"},
		{"A", 1, 1, "┌─┬───┬─┐\n│ │   │ │\n│ │ A │ │\n│ │   │ │\n└─┴───┴─┘"},
		{"東京", 0, 0, "┌─┬────┬─┐\n│ │東京│ │\n└─┴────┴─┘"},
	} {
		t.Run(fmt.Sprintf("%s/box=%d,%d", testCase.label, testCase.boxX, testCase.boxY), func(t *testing.T) {
			_, actual := renderBoxPaddingGraph(t, "graph LR\nA[["+testCase.label+"]]", 5, testCase.boxX, testCase.boxY, HeadStylePointer)
			if actual != testCase.want {
				t.Fatalf("subroutine = %q; want %q", actual, testCase.want)
			}
		})
	}
}

func TestSubroutineNodeRoutes(t *testing.T) {
	for _, source := range []string{
		"A[[A]]-->B(B)\nB-->C",
		"A[[A]]==>B[[B]]",
		"A[[A]]<-->B[[B]]\nB-->A",
		"A[[A]]-->|loop|A",
		"A[[A]]-->|one|B[[B]]\nA-->|two|B",
		"subgraph Outer\nA[[A]]-->B[[B]]\nend\nB-->C",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxX := range []int{0, 1} {
					for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
						t.Run(fmt.Sprintf("%s/%s/gap=%d/boxX=%d/heads=%s", source, direction, gap, boxX, headStyleString(heads)), func(t *testing.T) {
							rendered, output := renderBoxPaddingGraph(t, "graph "+direction+"\n"+source, gap, boxX, 0, heads)
							if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
								t.Fatalf("subroutine routes lost information: %v\n%s", defects, output)
							}
							assertNodeInteriors(t, rendered, output)
							if !strings.Contains(output, "┬") || !strings.Contains(output, "┴") {
								t.Fatalf("subroutine dividers missing:\n%s", output)
							}
						})
					}
				}
			}
		}
	}
}

func TestSubroutineDividerColors(t *testing.T) {
	rendered, output := renderHTMLColors(t, "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA[[東京]]:::box ==> B\nlinkStyle 0 stroke:#ff0000")
	contentNode, err := rendered.getNode("A")
	if err != nil {
		t.Fatal(err)
	}
	width, height := getDrawingSize(contentNode.drawing)
	for _, column := range []int{2, width - 2} {
		for row := 0; row <= height; row++ {
			cell := (*rendered.drawing)[contentNode.drawingCoord.x+column][contentNode.drawingCoord.y+row]
			if !strings.HasPrefix(cell, "<span style='color: #00ff00'>") {
				t.Errorf("divider lost border color: %q\n%s", cell, output)
			}
		}
	}
}
