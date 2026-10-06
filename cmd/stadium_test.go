package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestStadiumNodeRendering(t *testing.T) {
	for _, testCase := range []struct {
		label      string
		boxX, boxY int
		want       string
	}{
		{"A", 0, 0, "╭─╮\n(A)\n╰─╯"},
		{"Start", 1, 0, "╭───────╮\n( Start )\n╰───────╯"},
		{"A", 1, 1, "╭───╮\n(   )\n( A )\n(   )\n╰───╯"},
		{"東京", 0, 0, "╭────╮\n(東京)\n╰────╯"},
	} {
		t.Run(fmt.Sprintf("%s/box=%d,%d", testCase.label, testCase.boxX, testCase.boxY), func(t *testing.T) {
			rendered, actual := renderBoxPaddingGraph(t, "graph LR\nA(["+testCase.label+"])", 5, testCase.boxX, testCase.boxY, HeadStylePointer)
			if rendered.nodes[0].shape == ShapeRoundedRectangle {
				t.Fatal("stadium must be distinct from rounded rectangle")
			}
			if actual != testCase.want {
				t.Fatalf("stadium = %q; want %q", actual, testCase.want)
			}
		})
	}
}

func TestStadiumNodeRoutes(t *testing.T) {
	for _, source := range []string{
		"A([A])-->B[[B]]\nB-->C(C)",
		"A([A])==>B([B])",
		"A([A])<-->B([B])\nB-->A",
		"A([A])-->|loop|A",
		"A([A])-->|one|B([B])\nA-->|two|B",
		"subgraph Outer\nA([A])-->B([B])\nend\nB-->C",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxX := range []int{0, 1} {
					for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
						t.Run(fmt.Sprintf("%s/%s/gap=%d/boxX=%d/heads=%s", source, direction, gap, boxX, headStyleString(heads)), func(t *testing.T) {
							rendered, output := renderBoxPaddingGraph(t, "graph "+direction+"\n"+source, gap, boxX, 0, heads)
							if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
								t.Fatalf("stadium routes lost information: %v\n%s", defects, output)
							}
							assertNodeInteriors(t, rendered, output)
							if !strings.ContainsAny(output, "()") {
								t.Fatalf("stadium end caps missing:\n%s", output)
							}
						})
					}
				}
			}
		}
	}
}

func TestStadiumEndCapColors(t *testing.T) {
	_, output := renderHTMLColors(t, "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA([Start]):::box ==> B\nlinkStyle 0 stroke:#ff0000")
	for _, glyph := range []string{"(", "╭", "╞"} {
		if !strings.Contains(output, "<span style='color: #00ff00'>"+glyph+"</span>") {
			t.Errorf("stadium border lost its color at %s:\n%s", glyph, output)
		}
	}
}
