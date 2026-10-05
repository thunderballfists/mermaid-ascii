package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/lib"
)

func chainGaps(output, direction string) []int {
	var starts, ends []int
	for rowIndex, row := range strings.Split(output, "\n") {
		for columnIndex, cell := range []rune(row) {
			if direction == "LR" && rowIndex == 0 {
				if cell == '┌' {
					starts = append(starts, columnIndex)
				} else if cell == '┐' {
					ends = append(ends, columnIndex)
				}
			} else if direction == "TD" {
				if cell == '┌' {
					starts = append(starts, rowIndex)
				} else if cell == '└' {
					ends = append(ends, rowIndex)
				}
			}
		}
	}
	var gaps []int
	for index := 1; index < len(starts) && index <= len(ends); index++ {
		gaps = append(gaps, starts[index]-ends[index-1]-1)
	}
	return gaps
}

func TestUniformSpacingWithLabels(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, heads := range []string{"pointer", "open", "wide", "none"} {
			t.Run(direction+"/"+heads, func(t *testing.T) {
				var options lib.Options
				if err := json.Unmarshal([]byte(fmt.Sprintf(`{"UniformSpacing":true,"PaddingX":1,"PaddingY":1,"BoxPaddingY":0,"Heads":%q}`, heads)), &options); err != nil {
					t.Fatal(err)
				}
				output, err := lib.Render("graph "+direction+"\nA-->|long_edge_label|B\nB-->C", options)
				if err != nil {
					t.Fatal(err)
				}
				gaps := chainGaps(output, direction)
				if len(gaps) != 2 || gaps[0] != gaps[1] || gaps[0] < 1 {
					t.Fatalf("expected two equal gaps, got %v\n%s", gaps, output)
				}
				if !strings.Contains(output, "long_edge_label") {
					t.Fatalf("label lost:\n%s", output)
				}
			})
		}
	}
}

func TestUniformSpacingOptInAndASCIIGuard(t *testing.T) {
	const source = "graph LR\nA-->|long_edge_label|B\nB-->C"
	for _, ascii := range []bool{false, true} {
		options := lib.Options{UseASCII: ascii}
		baseline, err := lib.Render(source, options)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(`{"UniformSpacing":true}`), &options); err != nil {
			t.Fatal(err)
		}
		uniform, err := lib.Render(source, options)
		if err != nil {
			t.Fatal(err)
		}
		if ascii && uniform != baseline || !ascii && uniform == baseline {
			t.Fatalf("unexpected uniform spacing (ascii=%v):\n%s", ascii, uniform)
		}
		repeated, err := lib.Render(source, lib.Options{UseASCII: ascii})
		if err != nil || repeated != baseline {
			t.Fatalf("uniform spacing leaked: %v\n%s", err, repeated)
		}
		if !ascii && reflect.DeepEqual(chainGaps(baseline, "LR"), chainGaps(uniform, "LR")) {
			t.Fatal("label-expanded gaps were not equalized")
		}
	}
}

func TestUniformSpacingRoutes(t *testing.T) {
	for _, source := range []string{
		"A-->|long_edge_label|B\nB-->C\nA-->D",
		"A<-->B\nB-->C",
		"A-->|one|B\nA-->|two|B\nB-->C",
		"A-->|long_first_label|B\nA-->|long_second_label|B\nA-->|long_third_label|B\nB-->C",
		"A-->|one|B\nB-->|two|A\nB-->C",
		"subgraph Outer\nA-->B\nsubgraph Inner\nC-->D\nend\nB-->C\nend\nD-->E",
		"subgraph One\nA-->B\nend\nsubgraph Two\nC-->D\nend\nA-->|cross|D\nB-->E",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxX := range []int{0, 1} {
					for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
						t.Run(fmt.Sprintf("%s/%s/gap=%d/boxX=%d/heads=%s", source, direction, gap, boxX, headStyleString(heads)), func(t *testing.T) {
							rendered, output := renderLayoutGraph(t, "graph "+direction+"\n"+source, gap, boxX, 0, heads, true)
							if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
								t.Fatalf("uniform gaps lost information: %v\n%s", defects, output)
							}
							assertNodeInteriors(t, rendered, output)
							assertUniformNodeGaps(t, rendered, output)
						})
					}
				}
			}
		}
	}
}

func assertUniformNodeGaps(t *testing.T, rendered graph, output string) {
	t.Helper()
	for _, horizontal := range []bool{false, true} {
		bands := map[int][2]int{}
		for _, contentNode := range rendered.nodes {
			width, height := getDrawingSize(contentNode.drawing)
			if horizontal {
				bands[contentNode.gridCoord.x] = [2]int{contentNode.drawingCoord.x, contentNode.drawingCoord.x + width}
			} else {
				bands[contentNode.gridCoord.y] = [2]int{contentNode.drawingCoord.y, contentNode.drawingCoord.y + height}
			}
		}
		expectedGap := -1
		for rank, before := range bands {
			if after, exists := bands[rank+4]; exists {
				gap := after[0] - before[1] - 1
				if expectedGap >= 0 && gap != expectedGap {
					t.Fatalf("unequal visible gaps horizontal=%v: got %d want %d\n%s", horizontal, gap, expectedGap, output)
				}
				expectedGap = gap
			}
		}
	}
}

func TestUniformSpacingSingleNode(t *testing.T) {
	var options lib.Options
	if err := json.Unmarshal([]byte(`{"UniformSpacing":true}`), &options); err != nil {
		t.Fatal(err)
	}
	output, err := lib.Render("graph LR\nA", options)
	if err != nil || output != "┌───┐\n│   │\n│ A │\n│   │\n└───┘" {
		t.Fatalf("single node changed: %v\n%s", err, output)
	}
}

func TestUniformSpacingWithLabelledSelfLoop(t *testing.T) {
	_, output := renderLayoutGraph(t, "graph LR\nA-->|loop|A\nA-->B\nB-->C", 1, 0, 0, HeadStylePointer, true)
	gaps := chainGaps(output, "LR")
	if len(gaps) != 2 || gaps[0] != gaps[1] {
		t.Fatalf("expanded border band produced uneven visible gaps: %v\n%s", gaps, output)
	}
}

func TestUniformSpacingAfterLabelPlacement(t *testing.T) {
	rendered, output := renderLayoutGraph(t, "graph LR\nA-->|long1|B\nA-->|long2|B\nA-->|long3|B\nB-->|short|C", 1, 0, 0, HeadStylePointer, true)
	assertUniformNodeGaps(t, rendered, output)
}
