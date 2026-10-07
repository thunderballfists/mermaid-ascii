package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestCylinderNodeRendering(t *testing.T) {
	for _, tc := range []struct {
		label string
		x, y  int
		want  string
	}{
		{"A", 0, 0, "╭─╮\n├─┤\n│A│\n╰─╯"},
		{"DB", 1, 0, "╭────╮\n├────┤\n│ DB │\n╰────╯"},
		{"A", 1, 1, "╭───╮\n├───┤\n│   │\n│ A │\n│   │\n╰───╯"},
		{"東京", 0, 0, "╭────╮\n├────┤\n│東京│\n╰────╯"},
	} {
		t.Run(fmt.Sprintf("%s/padding=%d,%d", tc.label, tc.x, tc.y), func(t *testing.T) {
			_, got := renderBoxPaddingGraph(t, "graph LR\nA[("+tc.label+")]", 5, tc.x, tc.y, HeadStylePointer)
			if got != tc.want {
				t.Fatalf("cylinder = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestCylinderNodeRoutes(t *testing.T) {
	for _, source := range []string{
		"A[(A)]-->B[[B]]\nB-->C(C)",
		"A[(A)]==>B[(B)]",
		"A[(A)]<-->B[(B)]\nB-->A",
		"A[(A)]-->|loop|A",
		"A[(A)]-->|one|B[(B)]\nA-->|two|B",
		"subgraph Outer\nA[(A)]-->B[(B)]\nend\nB-->C",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxY := range []int{0, 1} {
					for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
						t.Run(fmt.Sprintf("%s/%s/gap=%d/y=%d/heads=%s", source, direction, gap, boxY, headStyleString(heads)), func(t *testing.T) {
							g, output := renderBoxPaddingGraph(t, "graph "+direction+"\n"+source, gap, 0, boxY, heads)
							if defects := auditRenderDefects(g, output, nil, false, heads); len(defects) != 0 {
								t.Fatalf("cylinder routes: %v\n%s", defects, output)
							}
							assertNodeInteriors(t, g, output)
							assertPortsOnBorders(t, g, output)
						})
					}
				}
			}
		}
	}
}

func TestCylinderRimColors(t *testing.T) {
	g, output := renderHTMLColors(t, "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA[(DB)]:::box ==> B\nlinkStyle 0 stroke:#ff0000")
	n := g.nodes[0]
	for x := 1; x < len(*n.drawing)-1; x++ {
		if got := (*g.drawing)[n.drawingCoord.x+x][n.drawingCoord.y+1]; got != "<span style='color: #00ff00'>─</span>" {
			t.Errorf("rim column %d = %q", x, got)
		}
	}
	if !strings.Contains(output, "<span style='color: #00ff00'>╞</span>") {
		t.Errorf("cylinder port lost node color:\n%s", output)
	}
}

func TestCylinderUniformSpacing(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		g, output := renderLayoutGraph(t, "graph "+direction+"\nA[(DB)]-->|a long label|B[[Run]]\nB-->C(Done)", 1, 0, 0, HeadStylePointer, true)
		if defects := auditRenderDefects(g, output, nil, false, HeadStylePointer); len(defects) != 0 {
			t.Fatalf("cylinder spacing: %v\n%s", defects, output)
		}
		assertNodeInteriors(t, g, output)
		assertUniformNodeGaps(t, g, output)
	}
}

func TestASCIICylinderPreserved(t *testing.T) {
	for _, source := range []string{"A[(DB)]", "A[(DB)] --> B", "A[(DB)]-->B", "A[(DB)]-->|loop|A"} {
		_, want := renderAuditGraph(t, "graph LR\n"+strings.ReplaceAll(source, "[(DB)]", "[\"(DB)\"]"), true, 5, HeadStylePointer)
		_, got := renderAuditGraph(t, "graph LR\n"+source, true, 5, HeadStylePointer)
		if got != want {
			t.Fatalf("ASCII cylinder changed for %s:\n%s\nwant:\n%s", source, got, want)
		}
	}
}
