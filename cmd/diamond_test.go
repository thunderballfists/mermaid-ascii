package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestDiamondNodeRendering(t *testing.T) {
	for _, tc := range []struct {
		label string
		x, y  int
		want  string
	}{
		{"A", 0, 0, "  ∧  \n ╱ ╲ \n< A >\n ╲ ╱ \n  ∨  "},
		{"Yes", 0, 0, "    ∧    \n  ╱─ ─╲  \n<─ Yes ─>\n  ╲─ ─╱  \n    ∨    "},
		{"A", 1, 1, "      ∧      \n    ╱─ ─╲    \n  ╱─     ─╲  \n<─    A    ─>\n  ╲─     ─╱  \n    ╲─ ─╱    \n      ∨      "},
		{"東京", 0, 0, "     ∧     \n  ╱── ──╲  \n<─ 東京  ─>\n  ╲── ──╱  \n     ∨     "},
	} {
		t.Run(fmt.Sprintf("%s/padding=%d,%d", tc.label, tc.x, tc.y), func(t *testing.T) {
			_, got := renderBoxPaddingGraph(t, "graph LR\nA{"+tc.label+"}", 5, tc.x, tc.y, HeadStylePointer)
			if got != tc.want {
				t.Fatalf("diamond = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestDiamondNodeRoutes(t *testing.T) {
	for _, source := range []string{
		"A{A}-->B[[B]]\nB-->C(C)",
		"A{A}==>B{B}",
		"A{A}<-->B{B}\nB-->A",
		"A{A}-->|loop|A",
		"A{A}-->|one|B{B}\nA-->|two|B",
		"subgraph Outer\nA{A}-->B{B}\nend\nB-->C",
		"A{A}-->B[(B)]\nA-->C[long rectangle label]",
		"A{A}-->B\nB-->C{C}\nC-->|x|A",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxY := range []int{0, 1} {
					for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
						t.Run(fmt.Sprintf("%s/%s/gap=%d/y=%d/heads=%s", source, direction, gap, boxY, headStyleString(heads)), func(t *testing.T) {
							g, output := renderBoxPaddingGraph(t, "graph "+direction+"\n"+source, gap, 0, boxY, heads)
							if defects := auditRenderDefects(g, output, nil, false, heads); len(defects) != 0 {
								t.Fatalf("diamond routes: %v\n%s", defects, output)
							}
							assertNodeInteriors(t, g, output)
							assertPortsOnBorders(t, g, output)
							assertDiamondTips(t, g)
						})
					}
				}
			}
		}
	}
}

func assertDiamondTips(t *testing.T, g graph) {
	t.Helper()
	for _, n := range g.nodes {
		if n.shape != ShapeDiamond {
			continue
		}
		width, height := getDrawingSize(n.drawing)
		for _, tip := range []struct {
			x, y  int
			glyph string
		}{
			{width / 2, 0, "∧"}, {width / 2, height, "∨"}, {0, height / 2, "<"}, {width, height / 2, ">"},
		} {
			if got := (*g.drawing)[n.drawingCoord.x+tip.x][n.drawingCoord.y+tip.y]; got != tip.glyph {
				t.Errorf("diamond %s tip (%d,%d) = %q; want %q", n.name, tip.x, tip.y, got, tip.glyph)
			}
		}
	}
}

func TestDiamondBorderColors(t *testing.T) {
	_, output := renderHTMLColors(t, "graph LR\nclassDef box color:#112233,stroke:#00ff00\nA{╱}:::box ==> B\nlinkStyle 0 stroke:#ff0000")
	for _, glyph := range []string{"∧", "∨", "<", ">", "╱", "╲"} {
		if !strings.Contains(output, "<span style='color: #00ff00'>"+glyph+"</span>") {
			t.Errorf("diamond border lost stroke at %s:\n%s", glyph, output)
		}
	}
	if !strings.Contains(output, "<span style='color: #112233'>╱</span>") {
		t.Fatalf("border-shaped label lost label color:\n%s", output)
	}
}

func TestDiamondUniformSpacing(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		g, output := renderLayoutGraph(t, "graph "+direction+"\nA{Ready}-->|a long label|B[(DB)]\nB-->C(Done)", 1, 0, 0, HeadStylePointer, true)
		if defects := auditRenderDefects(g, output, nil, false, HeadStylePointer); len(defects) != 0 {
			t.Fatalf("diamond spacing: %v\n%s", defects, output)
		}
		assertNodeInteriors(t, g, output)
		assertUniformNodeGaps(t, g, output)
		assertDiamondTips(t, g)
	}
}

func TestDiamondLabelPadding(t *testing.T) {
	for _, x := range []int{0, 1, 3} {
		for _, y := range []int{0, 1, 3} {
			g, _ := renderBoxPaddingGraph(t, "graph LR\nA{A}", 1, x, y, HeadStylePointer)
			d := *g.nodes[0].drawing
			labelX, labelY := -1, -1
			for column := range d {
				for row, glyph := range d[column] {
					if glyph == "A" {
						labelX, labelY = column, row
					}
				}
			}
			if labelX < 0 {
				t.Fatal("diamond label missing")
			}
			for column := labelX - x; column <= labelX+x; column++ {
				for row := labelY - y; row <= labelY+y; row++ {
					if column == labelX && row == labelY {
						continue
					}
					if d[column][row] != " " {
						t.Fatalf("padding=%d,%d intersects diamond at %d,%d: %q", x, y, column, row, d[column][row])
					}
				}
			}
		}
	}
}

func TestDiamondSharedBandsKeepClosedOutline(t *testing.T) {
	for _, bands := range [][2]int{{4, 3}, {20, 6}, {6, 20}, {42, 31}} {
		g, _ := renderBoxPaddingGraph(t, "graph LR\nA{A}", 1, 0, 0, HeadStylePointer)
		n := g.nodes[0]
		g.columnWidth[n.gridCoord.x+1], g.rowHeight[n.gridCoord.y+1] = bands[0], bands[1]
		g.normalizeDiamondBands()
		n.setDrawing(g)
		width, height := getDrawingSize(n.drawing)
		if width%2 != 0 || height%2 != 0 || width < height {
			t.Fatalf("bands=%v: invalid diamond bounds %d,%d", bands, width, height)
		}
		border := make(map[drawingCoord]bool)
		for x := 0; x <= width; x++ {
			for y := 0; y <= height; y++ {
				if glyph := (*n.drawing)[x][y]; glyph != " " && glyph != "A" {
					border[drawingCoord{x, y}] = true
				}
			}
		}
		// A text outline must remain connected across every diagonal step.
		queue := []drawingCoord{{width / 2, 0}}
		delete(border, queue[0])
		for len(queue) > 0 {
			position := queue[0]
			queue = queue[1:]
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					next := drawingCoord{position.x + dx, position.y + dy}
					if border[next] {
						delete(border, next)
						queue = append(queue, next)
					}
				}
			}
		}
		if len(border) != 0 {
			t.Fatalf("bands=%v: disconnected diamond outline: %v", bands, border)
		}
	}
}

func TestASCIIDiamondPreserved(t *testing.T) {
	for _, source := range []string{"A{Yes}", "A{Yes} --> B", "A{Yes}-->B", "A{Yes}-->|loop|A"} {
		_, want := renderAuditGraph(t, "graph LR\n"+strings.ReplaceAll(source, "{Yes}", "[Yes]"), true, 5, HeadStylePointer)
		_, got := renderAuditGraph(t, "graph LR\n"+source, true, 5, HeadStylePointer)
		if got != want {
			t.Fatalf("ASCII diamond changed for %s:\n%s\nwant:\n%s", source, got, want)
		}
	}
}
