package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestMergeJunctionsConnectivity(t *testing.T) {
	cases := []struct {
		base, overlay, expected string
	}{
		{"║", "└", "╟"},
		{"║", "┘", "╢"},
		{"═", "┐", "╤"},
		{"═", "┘", "╧"},
		{"╚", "║", "╠"},
		{"╔", "╚", "╠"},
		{"╟", "─", "╫"},
		{"╤", "│", "╪"},
		{"┈", "│", "┼"},
		{"┊", "─", "┼"},
		{"┄", "│", "┼"},
		{"┆", "─", "┼"},
		{"╭", "│", "├"},
		{"╴", "╷", "┐"},
		{"┃", "┗", "┣"},
		{"┃", "└", "┣"},
		{"┃", "═", "╬"},
		{"┈", "┈", "┈"},
		{"╭", "╭", "╭"},
	}
	for _, testCase := range cases {
		t.Run(testCase.base+"+"+testCase.overlay, func(t *testing.T) {
			for _, pair := range [][2]string{{testCase.base, testCase.overlay}, {testCase.overlay, testCase.base}} {
				if actual := mergeJunctions(pair[0], pair[1]); actual != testCase.expected {
					t.Errorf("mergeJunctions(%q, %q) = %q; want %q", pair[0], pair[1], actual, testCase.expected)
				}
			}
		})
	}
}

func TestMergeDrawingsJunctionModeBoundary(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	cases := []struct {
		ascii                   bool
		base, overlay, expected string
	}{
		{false, "║", "└", "╟"},
		{false, "└", "║", "╟"},
		{false, "║", "►", "►"},
		{false, "║", "label", "label"},
		{false, "║", "\x1b[31m└\x1b[0m", "\x1b[31m└\x1b[0m"},
		{true, "║", "└", "└"},
		{true, "|", "+", "+"},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("ascii=%v/%s+%s", testCase.ascii, testCase.base, testCase.overlay), func(t *testing.T) {
			useAscii = testCase.ascii
			base, overlay := mkDrawing(0, 0), mkDrawing(0, 0)
			(*base)[0][0], (*overlay)[0][0] = testCase.base, testCase.overlay
			actual := drawingToString(mergeDrawings(base, drawingCoord{}, overlay))
			if actual != testCase.expected {
				t.Fatalf("merged drawing = %q; want %q", actual, testCase.expected)
			}
		})
	}
}

func TestMergeJunctionsPreservesEveryDirection(t *testing.T) {
	shapes := []string{
		"", "╵╹", "╶╺", "└┗╚╘╙╰", "╷╻", "│┃║┊┆",
		"┌┏╔╒╓╭", "├┣╠╞╟", "╴╸", "┘┛╝╛╜╯", "─━═┈┄",
		"┴┻╩╧╨", "┐┓╗╕╖╮", "┤┫╣╡╢", "┬┳╦╤╥", "┼╋╬╪╫",
	}
	directionsByGlyph := map[string]int{}
	for directions, glyphs := range shapes {
		for _, glyph := range glyphs {
			directionsByGlyph[string(glyph)] = directions
		}
	}
	for base, baseDirections := range directionsByGlyph {
		for overlay, overlayDirections := range directionsByGlyph {
			merged := mergeJunctions(base, overlay)
			if actual, known := directionsByGlyph[merged]; !known || actual != baseDirections|overlayDirections {
				t.Errorf("mergeJunctions(%q, %q) = %q: directions=%04b; want %04b", base, overlay, merged, actual, baseDirections|overlayDirections)
			}
			if reversed := mergeJunctions(overlay, base); merged != reversed {
				t.Errorf("merge order changed %q + %q: %q != %q", base, overlay, merged, reversed)
			}
		}
	}
}

func TestRenderMixedBranchConnectivity(t *testing.T) {
	for _, padding := range []int{1, 5} {
		t.Run(fmt.Sprintf("padding=%d", padding), func(t *testing.T) {
			_, output := renderAuditGraph(t, "graph LR\nA-->B\nA-.->C\nA==>D", false, padding, HeadStylePointer)
			if strings.Count(output, "╟┈") != 1 {
				t.Fatalf("want a light dotted branch on a continuous double trunk:\n%s", output)
			}
			if strings.Count(output, "►") != 3 {
				t.Fatalf("junction merge must preserve all three heads:\n%s", output)
			}
		})
	}
}

func TestRenderBoxDrawingEdgeLabels(t *testing.T) {
	for _, label := range []string{"║", "╔", "╚", "╟", "┈", "┊", "╭", "┃", "┏"} {
		t.Run(label, func(t *testing.T) {
			_, output := renderAuditGraph(t, "graph LR\nA-->|"+label+"|B", false, 5, HeadStylePointer)
			if !strings.Contains(output, "─"+label+"►") {
				t.Fatalf("box-drawing label must overlay the route unchanged:\n%s", output)
			}
		})
	}
}

func TestBoxDrawingGroupTitleOverlay(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	rendered := graph{
		drawing:   mkDrawing(8, 4),
		subgraphs: []*subgraph{{name: "║", nodes: []*node{{}}, minX: 0, maxX: 8, minY: 0, maxY: 4}},
	}
	for column := range *rendered.drawing {
		(*rendered.drawing)[column][1] = "─"
	}
	rendered.drawSubgraphLabels()
	if actual := strings.Split(drawingToString(rendered.drawing), "\n")[1]; actual != "────║────" {
		t.Fatalf("group title must overlay strokes unchanged: %q", actual)
	}
}
