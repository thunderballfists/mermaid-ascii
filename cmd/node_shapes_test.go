package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestRoundedNodeRendering(t *testing.T) {
	for _, testCase := range []struct {
		label      string
		boxX, boxY int
		want       string
	}{
		{"A", 0, 0, "╭─╮\n│A│\n╰─╯"},
		{"A", 1, 0, "╭───╮\n│ A │\n╰───╯"},
		{"A", 1, 1, "╭───╮\n│   │\n│ A │\n│   │\n╰───╯"},
		{"東京", 0, 0, "╭────╮\n│東京│\n╰────╯"},
		{"e\u0301", 0, 0, "╭─╮\n│e\u0301│\n╰─╯"},
	} {
		t.Run(fmt.Sprintf("%s/box=%d,%d", testCase.label, testCase.boxX, testCase.boxY), func(t *testing.T) {
			_, actual := renderBoxPaddingGraph(t, "graph LR\nA("+testCase.label+")", 5, testCase.boxX, testCase.boxY, HeadStylePointer)
			if actual != testCase.want {
				t.Fatalf("rounded node = %q; want %q", actual, testCase.want)
			}
		})
	}
}

func TestRoundedNodeRoutes(t *testing.T) {
	for _, source := range []string{
		"A(A)-->B(B)\nB-->C(C)",
		"A(A)==>B(B)",
		"A(A)<-->B(B)\nB-->A",
		"A(A)-->|loop|A",
		"A(A)-->|one|B(B)\nA-->|two|B",
		"subgraph Outer\nA(A)-->B(B)\nend\nB-->C(C)",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxX := range []int{0, 1} {
					for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
						t.Run(fmt.Sprintf("%s/%s/gap=%d/boxX=%d/heads=%s", source, direction, gap, boxX, headStyleString(heads)), func(t *testing.T) {
							rendered, output := renderBoxPaddingGraph(t, "graph "+direction+"\n"+source, gap, boxX, 0, heads)
							if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
								t.Fatalf("rounded routes lost information: %v\n%s", defects, output)
							}
							assertNodeInteriors(t, rendered, output)
							if !strings.Contains(output, "╭") || !strings.Contains(output, "╰") {
								t.Fatalf("rounded outline missing:\n%s", output)
							}
						})
					}
				}
			}
		}
	}
}

func TestRoundedNodeBorderColors(t *testing.T) {
	_, output := renderHTMLColors(t, "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA(Start):::box ==> B\nlinkStyle 0 stroke:#ff0000")
	for _, glyph := range []string{"╭", "╮", "╰", "╯", "╞"} {
		if !strings.Contains(output, "<span style='color: #00ff00'>"+glyph+"</span>") {
			t.Errorf("rounded border lost its color at %s:\n%s", glyph, output)
		}
	}
}

func TestASCIIShapeRenderingPreserved(t *testing.T) {
	for _, testCase := range []struct{ shape, legacyLabel string }{
		{"A(Start)", "A"}, {"A([Start])", "A"}, {"A[[Start]]", "Start"},
	} {
		_, expected := renderAuditGraph(t, "graph LR\nA["+testCase.legacyLabel+"]-->B[Finish]", true, 5, HeadStylePointer)
		_, actual := renderAuditGraph(t, "graph LR\n"+testCase.shape+"-->B[Finish]", true, 5, HeadStylePointer)
		if actual != expected {
			t.Errorf("ASCII shape %s changed:\n%s", testCase.shape, actual)
		}
	}
}
