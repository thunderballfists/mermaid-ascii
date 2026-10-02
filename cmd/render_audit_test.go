package cmd

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func renderAuditGraph(t *testing.T, source string, ascii bool, padding int, heads HeadStyle) (graph, string) {
	t.Helper()
	originalASCII, originalHeads, originalDirection := useAscii, headStyle, graphDirection
	defer func() {
		useAscii, headStyle, graphDirection = originalASCII, originalHeads, originalDirection
	}()
	useAscii, headStyle = ascii, heads
	properties, err := mermaidFileToMap(source, "cli")
	if err != nil {
		t.Fatal(err)
	}
	properties.paddingX, properties.paddingY = padding, padding
	rendered := mkGraph(properties)
	rendered.setStyleClasses(properties)
	rendered.setSubgraphs(properties.subgraphs)
	rendered.createMapping()
	output := drawingToString(rendered.draw())
	return rendered, output
}

func captureAuditRender(t *testing.T, source string, ascii bool, padding int, heads HeadStyle) (rendered graph, output string, panicValue any) {
	t.Helper()
	defer func() {
		panicValue = recover()
	}()
	rendered, output = renderAuditGraph(t, source, ascii, padding, heads)
	return
}

func TestRenderAuditMatrix(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{"horizontal", "graph LR\nA-->B"},
		{"vertical", "graph TD\nA-->B"},
		{"undirected", "graph LR\nA---B"},
		{"bidirectional", "graph LR\nA<-->B"},
		{"incoming-branches", "graph LR\nA-->C\nB-->C"},
		{"mixed-branches", "graph LR\nA-->B\nA-.->C\nA==>D"},
		{"cycle", "graph LR\nA-->B\nB-->A"},
		{"self-loop", "graph LR\nA-->A"},
		{"parallel", "graph LR\nA-->|one|B\nA-->|two|B"},
		{"unicode-labels", "graph LR\nA[Café]-->|é|B[東京]"},
		{"long-label", "graph TD\nA-->|a long edge label|B"},
		{"nested", "graph LR\nsubgraph Outer\nA\nsubgraph Inner\nB-->C\nend\nA-->B\nend\nC-->D"},
	}
	for _, testCase := range cases {
		for _, padding := range []int{1, 5} {
			for _, ascii := range []bool{true, false} {
				for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
					name := fmt.Sprintf("%s/ascii=%v/padding=%d/heads=%s", testCase.name, ascii, padding, headStyleString(heads))
					t.Run(name, func(t *testing.T) {
						rendered, output, panicValue := captureAuditRender(t, testCase.source, ascii, padding, heads)
						actual := auditRenderDefects(rendered, output, panicValue, ascii, heads)
						expected := knownAuditDefects(testCase.name, ascii, padding, heads)
						if !reflect.DeepEqual(actual, expected) {
							t.Fatalf("characterization changed: defects=%v expected=%v panic=%v\n%s", actual, expected, panicValue, output)
						}
						t.Logf("defects=%v\n%s", actual, output)
					})
				}
			}
		}
	}
}

func auditRenderDefects(rendered graph, output string, panicValue any, ascii bool, heads HeadStyle) []string {
	if panicValue != nil {
		return []string{"render-panic"}
	}
	defects := map[string]bool{}
	expectedHeads := 0
	for _, connection := range rendered.edges {
		if ascii || heads != HeadStyleNone {
			if connection.arrowDir == arrowDirForward || connection.arrowDir == arrowDirBoth {
				expectedHeads++
			}
			if connection.arrowDir == arrowDirBackward || connection.arrowDir == arrowDirBoth {
				expectedHeads++
			}
		}
		if connection.text != "" && !strings.Contains(output, connection.text) {
			defects["edge-label-lost"] = true
		}
	}
	actualHeads := strings.Count(output, ">") + strings.Count(output, "<") + strings.Count(output, "^") + strings.Count(output, "v")
	if !ascii {
		actualHeads = strings.Count(output, "►") + strings.Count(output, "◄") + strings.Count(output, "▲") + strings.Count(output, "▼") + strings.Count(output, "▶") + strings.Count(output, "◀") + strings.Count(output, "→") + strings.Count(output, "←") + strings.Count(output, "↑") + strings.Count(output, "↓")
	}
	if actualHeads != expectedHeads {
		defects["arrowhead-count"] = true
	}
	for _, contentNode := range rendered.nodes {
		if !strings.Contains(output, contentNode.label) {
			defects["node-label-lost"] = true
		}
		width, height := getDrawingSize(contentNode.drawing)
		left, top := contentNode.drawingCoord.x, contentNode.drawingCoord.y
		right, bottom := left+width, top+height
		for _, group := range rendered.subgraphs {
			verticalOverlap := top <= group.maxY && bottom >= group.minY
			horizontalOverlap := left <= group.maxX && right >= group.minX
			if verticalOverlap && ((group.minX >= left && group.minX <= right) || (group.maxX >= left && group.maxX <= right)) || horizontalOverlap && ((group.minY >= top && group.minY <= bottom) || (group.maxY >= top && group.maxY <= bottom)) {
				defects["group-border-overlap"] = true
			}
		}
	}
	rows := strings.Split(output, "\n")
	for rowIndex, row := range rows {
		cells := []rune(row)
		for columnIndex, cell := range cells {
			if cell != '└' || rowIndex == 0 || rowIndex+1 >= len(rows) || columnIndex+1 >= len(cells) || cells[columnIndex+1] != '┈' {
				continue
			}
			above, below := []rune(rows[rowIndex-1]), []rune(rows[rowIndex+1])
			if columnIndex < len(above) && columnIndex < len(below) && above[columnIndex] == '║' && below[columnIndex] == '║' {
				defects["mixed-junction"] = true
			}
		}
	}
	result := []string{}
	for defect := range defects {
		result = append(result, defect)
	}
	sort.Strings(result)
	return result
}

func knownAuditDefects(name string, ascii bool, padding int, heads HeadStyle) []string {
	if name == "self-loop" && padding == 1 {
		return []string{"render-panic"}
	}
	defects := []string{}
	hasHeads := ascii || heads != HeadStyleNone
	if name == "undirected" && ascii || name == "bidirectional" && (ascii || padding == 1 && hasHeads) || (name == "cycle" && padding == 1 || name == "parallel" || name == "long-label" && padding == 1) && hasHeads {
		defects = append(defects, "arrowhead-count")
	}
	if name == "parallel" || name == "unicode-labels" {
		defects = append(defects, "edge-label-lost")
	}
	if name == "nested" && padding == 1 {
		defects = append(defects, "group-border-overlap")
	}
	if name == "mixed-branches" && !ascii {
		defects = append(defects, "mixed-junction")
	}
	if name == "unicode-labels" {
		defects = append(defects, "node-label-lost")
	}
	return defects
}
