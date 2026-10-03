package cmd

import (
	"fmt"
	"testing"
)

func TestUnicodeOpposedEdgeClearance(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, padding := range []int{1, 2, 5} {
			for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
				t.Run(fmt.Sprintf("%s/padding=%d/heads=%s", direction, padding, headStyleString(heads)), func(t *testing.T) {
					rendered, output := renderAuditGraph(t, "graph "+direction+"\nA-->B\nB-->A", false, padding, heads)
					if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
						t.Fatalf("opposed edges lost a head: %v\n%s", defects, output)
					}
				})
			}
		}
	}
}

func TestOpposedClearanceUsesPhysicalHeads(t *testing.T) {
	originalASCII, originalHeads := useAscii, headStyle
	t.Cleanup(func() { useAscii, headStyle = originalASCII, originalHeads })
	useAscii, headStyle = false, HeadStylePointer
	for _, vertical := range []bool{false, true} {
		for _, pair := range []struct {
			first, second arrowDirection
			gap           int
		}{
			{arrowDirForward, arrowDirForward, 2},
			{arrowDirBackward, arrowDirBackward, 2},
			{arrowDirForward, arrowDirBackward, 1},
			{arrowDirBackward, arrowDirForward, 1},
			{arrowDirForward, arrowDirOpen, 1},
			{arrowDirBoth, arrowDirForward, 2},
		} {
			t.Run(fmt.Sprintf("vertical=%v/%d/%d", vertical, pair.first, pair.second), func(t *testing.T) {
				first := &edge{arrowDir: pair.first, path: []gridCoord{{0, 0}, {2, 0}}}
				if vertical {
					first.path[1] = gridCoord{0, 2}
				}
				second := &edge{arrowDir: pair.second, path: []gridCoord{first.path[1], first.path[0]}}
				rendered := graph{edges: []*edge{first, second}, columnWidth: map[int]int{0: 1, 1: 1, 2: 1}, rowHeight: map[int]int{0: 1, 1: 1, 2: 1}}
				rendered.reserveArrowHeadClearance(second)
				gap := rendered.columnWidth[1]
				if vertical {
					gap = rendered.rowHeight[1]
				}
				if gap != pair.gap {
					t.Fatalf("gap = %d; want %d", gap, pair.gap)
				}
			})
		}
	}
}
