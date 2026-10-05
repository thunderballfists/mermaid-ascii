package cmd

import (
	"fmt"
	"testing"
)

func TestUnicodeSelfLoopLabelsWithOutgoingEdges(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, gap := range []int{1, 5} {
			for _, boxX := range []int{0, 1} {
				for _, boxY := range []int{0, 1} {
					for _, uniform := range []bool{false, true} {
						for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
							for _, label := range []string{"loop", "東京", "👩‍💻"} {
								for _, operator := range []string{"-->", "<-->", "---"} {
									t.Run(fmt.Sprintf("%s/gap=%d/box=%d,%d/uniform=%v/heads=%s/%s/%s", direction, gap, boxX, boxY, uniform, headStyleString(heads), label, operator), func(t *testing.T) {
										source := "graph " + direction + "\nA" + operator + "|" + label + "|A\nA-->B\nB-->C"
										rendered, output := renderLayoutGraph(t, source, gap, boxX, boxY, heads, uniform)
										if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
											t.Fatalf("labelled self-loop lost information: %v\n%s", defects, output)
										}
										assertNodeInteriors(t, rendered, output)
										assertPortsOnBorders(t, rendered, output)
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
	}
}

func assertPortsOnBorders(t *testing.T, rendered graph, output string) {
	t.Helper()
	for _, connection := range rendered.edges {
		for _, endpoint := range []struct {
			contentNode *node
			port        direction
		}{{connection.from, connection.startDir}, {connection.to, connection.endDir}} {
			width, height := getDrawingSize(endpoint.contentNode.drawing)
			left, top := endpoint.contentNode.drawingCoord.x, endpoint.contentNode.drawingCoord.y
			expected := drawingCoord{left + CeilDiv(width, 2), top + CeilDiv(height, 2)}
			switch endpoint.port {
			case Left:
				expected.x = left
			case Right:
				expected.x = left + width
			case Up:
				expected.y = top
			case Down:
				expected.y = top + height
			}
			actual := rendered.gridToDrawingCoord(*endpoint.contentNode.gridCoord, &endpoint.port)
			if actual != expected {
				t.Errorf("node %s port=%v is at %v; want border %v\n%s", endpoint.contentNode.name, endpoint.port, actual, expected, output)
			}
		}
	}
}
