package cmd

import (
	"fmt"
	"testing"
)

func TestUnicodeBentPortClearance(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, gap := range []int{1, 5} {
			for _, boxX := range []int{0, 1} {
				for _, boxY := range []int{0, 1} {
					for _, uniform := range []bool{false, true} {
						for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
							t.Run(fmt.Sprintf("%s/gap=%d/box=%d,%d/uniform=%v/heads=%s", direction, gap, boxX, boxY, uniform, headStyleString(heads)), func(t *testing.T) {
								rendered, output := renderLayoutGraph(t, "graph "+direction+"\nA<-->B\nB-->C\nD-->C", gap, boxX, boxY, heads, uniform)
								assertNodeInteriors(t, rendered, output)
								for _, connection := range rendered.edges {
									if len(connection.path) <= 2 {
										continue
									}
									for _, segment := range [][]gridCoord{connection.path[:2], connection.path[len(connection.path)-2:]} {
										endpoints := rendered.lineToDrawing(segment)
										distance := Abs(endpoints[0].x-endpoints[1].x) + Abs(endpoints[0].y-endpoints[1].y)
										if distance < 2 {
											t.Errorf("bent port has no outside cell: %v\n%s", endpoints, output)
										}
									}
								}
								if uniform {
									assertUniformNodeGaps(t, rendered, output)
								}
							})
						}
					}
				}
			}
		}
	}
}
