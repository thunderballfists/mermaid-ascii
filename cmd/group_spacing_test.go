package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestUnicodeGroupClearance(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, padding := range []int{1, 2, 5} {
			for _, source := range []string{
				"subgraph Outer\nA\nsubgraph Inner\nB-->C\nend\nA-->B\nend\nC-->D",
				"subgraph One\nA\nend\nsubgraph Two\nB\nend\nA-->B\nB-->C",
				"subgraph 東京東京東京東京\nA\nend\nA-->B",
				"subgraph Outer\nsubgraph Inner\nA\nend\nend\nA-->B",
				"subgraph One\nA\nend\nsubgraph Two\nB\nend",
				"subgraph First\nsubgraph InnerOne\nA\nend\nend\nsubgraph Second\nsubgraph InnerTwo\nB\nend\nend\nA-->|label|B",
				"subgraph Group0\nsubgraph Nested0\nD\nend\nE\nend\nB<--|label3|E\nE<--B",
				"A\nC\nD\nsubgraph Group2\nB\nend\nB<--|label0|C\nD<-->B\nA<--D",
			} {
				t.Run(fmt.Sprintf("%s/padding=%d/%s", direction, padding, source), func(t *testing.T) {
					rendered, output, panicValue := captureAuditRender(t, "graph "+direction+"\n"+source, false, padding, HeadStylePointer)
					if defects := auditRenderDefects(rendered, output, panicValue, false, HeadStylePointer); len(defects) != 0 {
						t.Fatalf("group layout lost information: %v panic=%v\n%s", defects, panicValue, output)
					}
					assertNodeInteriors(t, rendered, output)
					for _, group := range rendered.subgraphs {
						if !strings.Contains(output, group.name) {
							t.Errorf("group title missing %q\n%s", group.name, output)
						}
						for _, contentNode := range rendered.nodes {
							member := false
							for _, groupNode := range group.nodes {
								member = member || groupNode == contentNode
							}
							width, height := getDrawingSize(contentNode.drawing)
							left, top := contentNode.drawingCoord.x, contentNode.drawingCoord.y
							contained := left > group.minX && left+width < group.maxX && top > group.minY && top+height < group.maxY
							if contained != member {
								t.Errorf("node %s group %s: contained=%v member=%v\n%s", contentNode, group.name, contained, member, output)
							}
						}
						for _, child := range group.children {
							if child.minX <= group.minX || child.maxX >= group.maxX || child.minY <= group.minY || child.maxY >= group.maxY {
								t.Errorf("child %s is not contained by %s\n%s", child.name, group.name, output)
							}
						}
					}
				})
			}
		}
	}
}

func TestGroupClearanceLeavesUngroupedBandsUnchanged(t *testing.T) {
	for _, emptyGroup := range []bool{false, true} {
		rendered := graph{
			nodes:       []*node{{gridCoord: &gridCoord{0, 0}}, {gridCoord: &gridCoord{4, 0}}},
			columnWidth: map[int]int{3: 0},
			rowHeight:   map[int]int{},
		}
		if emptyGroup {
			rendered.subgraphs = []*subgraph{{}}
		}
		rendered.reserveGroupClearance()
		if rendered.columnWidth[3] != 0 {
			t.Fatalf("ungrouped gap grew to %d with emptyGroup=%v", rendered.columnWidth[3], emptyGroup)
		}
	}
}
