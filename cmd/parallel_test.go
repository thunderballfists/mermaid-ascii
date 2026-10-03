package cmd

import (
	"fmt"
	"reflect"
	"testing"
)

func TestUnicodeParallelRoutes(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, padding := range []int{1, 5} {
			for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
				for _, second := range []string{"A-->|two|B", "B-->|two|A"} {
					t.Run(fmt.Sprintf("%s/padding=%d/heads=%s/%s", direction, padding, headStyleString(heads), second), func(t *testing.T) {
						rendered, output, panicValue := captureAuditRender(t, "graph "+direction+"\nA-->|one|B\n"+second, false, padding, heads)
						if defects := auditRenderDefects(rendered, output, panicValue, false, heads); len(defects) != 0 {
							t.Fatalf("repeated edges must retain labels and heads: %v panic=%v\n%s", defects, panicValue, output)
						}
						firstPath, secondPath := rendered.edges[0].path, rendered.edges[1].path
						reversed := append([]gridCoord(nil), secondPath...)
						for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
							reversed[left], reversed[right] = reversed[right], reversed[left]
						}
						if reflect.DeepEqual(firstPath, secondPath) || reflect.DeepEqual(firstPath, reversed) {
							t.Fatalf("repeated edges share the same route: %v", firstPath)
						}
						assertNodeInteriors(t, rendered, output)
					})
				}
			}
		}
	}
}

func assertNodeInteriors(t *testing.T, rendered graph, output string) {
	t.Helper()
	for _, contentNode := range rendered.nodes {
		width, height := getDrawingSize(contentNode.drawing)
		for column := 1; column < width; column++ {
			for row := 1; row < height; row++ {
				actual := (*rendered.drawing)[contentNode.drawingCoord.x+column][contentNode.drawingCoord.y+row]
				if expected := (*contentNode.drawing)[column][row]; actual != expected {
					t.Errorf("route entered node %s at (%d,%d): %q; want %q\n%s", contentNode.name, column, row, actual, expected, output)
				}
			}
		}
	}
}

func TestUnicodeThreeParallelEdges(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, padding := range []int{1, 5} {
			for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
				t.Run(fmt.Sprintf("%s/padding=%d/heads=%s", direction, padding, headStyleString(heads)), func(t *testing.T) {
					rendered, output, panicValue := captureAuditRender(t, "graph "+direction+"\nA-->|one|B\nA-->|two|B\nA-->|three|B", false, padding, heads)
					if defects := auditRenderDefects(rendered, output, panicValue, false, heads); len(defects) != 0 {
						t.Fatalf("three parallel edges lost information: %v panic=%v\n%s", defects, panicValue, output)
					}
					assertNodeInteriors(t, rendered, output)
				})
			}
		}
	}
}

func TestUnicodeLabelledSelfLoopKeepsNodeInterior(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, label := range []string{"loop", "東京", "👩‍💻"} {
			t.Run(direction+"/"+label, func(t *testing.T) {
				rendered, output := renderAuditGraph(t, "graph "+direction+"\nA-->|"+label+"|A", false, 1, HeadStylePointer)
				if defects := auditRenderDefects(rendered, output, nil, false, HeadStylePointer); len(defects) != 0 {
					t.Fatalf("labelled self-loop lost information: %v\n%s", defects, output)
				}
				assertNodeInteriors(t, rendered, output)
			})
		}
	}
}

func TestUnicodeRepeatedEdgesWithBranches(t *testing.T) {
	for _, testCase := range []struct {
		source  string
		padding int
	}{
		{"graph TD\nA-->|x|C\nA<-->|long_edge_label|C\nB<-->C", 1},
		{"graph LR\nB-->C\nA-->C\nA---|x|B\nA<-->|long_edge_label|B", 2},
		{"graph TD\nB<-->C\nA<-->|x1|B\nA<--|x2|B\nA---|long_edge_label_3|C", 2},
		{"graph LR\nsubgraph Outer\nA-->|one|B\nA-->|two|B\nend", 5},
		{"graph LR\nB<--C\nA-->B\nB---A\nB<--|long_edge_label_4|A", 5},
		{"graph TD\nA-->C\nA---|x1|B\nA<--B", 1},
	} {
		t.Run(testCase.source, func(t *testing.T) {
			rendered, output, panicValue := captureAuditRender(t, testCase.source, false, testCase.padding, HeadStylePointer)
			if defects := auditRenderDefects(rendered, output, panicValue, false, HeadStylePointer); len(defects) != 0 {
				for _, connection := range rendered.edges {
					t.Logf("%s -> %s: direction=%v ports=%v/%v path=%v label=%v", connection.from, connection.to, connection.arrowDir, connection.startDir, connection.endDir, connection.path, connection.labelPosition)
				}
				t.Fatalf("branching repeated edges lost information: %v panic=%v\n%s", defects, panicValue, output)
			}
			assertNodeInteriors(t, rendered, output)
		})
	}
}
