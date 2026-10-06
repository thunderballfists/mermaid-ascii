package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestDoubleEdgeBoxAttachments(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	for _, testCase := range []struct {
		name              string
		to                gridCoord
		first, attachment drawingCoord
		border, expected  string
	}{
		{"right", gridCoord{2, 1}, drawingCoord{4, 3}, drawingCoord{3, 3}, "│", "╞"},
		{"left", gridCoord{0, 1}, drawingCoord{2, 3}, drawingCoord{3, 3}, "│", "╡"},
		{"down", gridCoord{1, 2}, drawingCoord{3, 4}, drawingCoord{3, 3}, "─", "╥"},
		{"up", gridCoord{1, 0}, drawingCoord{3, 2}, drawingCoord{3, 3}, "─", "╨"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			rendered := graph{drawing: mkDrawing(6, 6)}
			port := testCase.attachment
			(*rendered.drawing)[port.x][port.y] = testCase.border
			connection := edge{edgeStyle: StyleThick, path: []gridCoord{{1, 1}, testCase.to}}
			start := rendered.drawBoxStart(&connection, []drawingCoord{testCase.first})
			merged := mergeDrawings(rendered.drawing, drawingCoord{}, start)
			if actual := (*merged)[port.x][port.y]; actual != testCase.expected {
				t.Fatalf("box attachment = %q; want %q", actual, testCase.expected)
			}
		})
	}
}

func TestRenderDoubleEdgeThinBoxBorders(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
			for _, boxY := range []int{0, 1} {
				t.Run(fmt.Sprintf("%s/heads=%s/boxY=%d", direction, headStyleString(heads), boxY), func(t *testing.T) {
					rendered, output := renderBoxPaddingGraph(t, "graph "+direction+"\nA==>B", 5, 1, boxY, heads)
					want := "╞"
					if direction == "TD" {
						want = "╥"
					}
					if strings.Count(output, want) != 1 || strings.ContainsAny(output, "╠╣╦╩") {
						t.Fatalf("double edge must retain thin box borders:\n%s", output)
					}
					if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
						t.Fatalf("double edge lost information: %v\n%s", defects, output)
					}
					assertNodeInteriors(t, rendered, output)
				})
			}
		}
	}
}
