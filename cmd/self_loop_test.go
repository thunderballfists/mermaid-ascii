package cmd

import (
	"fmt"
	"testing"
)

func TestUnicodeSelfLoopClearance(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, padding := range []int{1, 2, 3, 5} {
			for _, operator := range []string{"-->", "==>", "-.->", "<-->", "---"} {
				for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
					t.Run(fmt.Sprintf("%s/padding=%d/%s/heads=%s", direction, padding, operator, headStyleString(heads)), func(t *testing.T) {
						rendered, output, panicValue := captureAuditRender(t, "graph "+direction+"\nA"+operator+"A", false, padding, heads)
						if defects := auditRenderDefects(rendered, output, panicValue, false, heads); len(defects) != 0 {
							t.Fatalf("self-loop must render without losing heads: defects=%v panic=%v\n%s", defects, panicValue, output)
						}
						contentNode := rendered.nodes[0]
						width, height := getDrawingSize(contentNode.drawing)
						for column := 1; column < width; column++ {
							for row := 1; row < height; row++ {
								actual := (*rendered.drawing)[contentNode.drawingCoord.x+column][contentNode.drawingCoord.y+row]
								if expected := (*contentNode.drawing)[column][row]; actual != expected {
									t.Errorf("loop entered node interior at (%d,%d): %q; want %q\n%s", column, row, actual, expected, output)
								}
							}
						}
					})
				}
			}
		}
	}
}
