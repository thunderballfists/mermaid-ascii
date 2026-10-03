package cmd

import (
	"strings"
	"testing"
)

func TestDrawTextDisplayCells(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	testCases := []struct {
		name string
		text string
		want string
	}{
		{"accent", "Café", " Café    "},
		{"combining", "e\u0301", " e\u0301       "},
		{"wide", "東京!", " 東京!   "},
		{"emoji", "👩‍💻!", " 👩‍💻!     "},
		{"flag", "🇩🇪!", " 🇩🇪!     "},
		{"zero-prefix", "\u200bCafé", " \u200bCafé    "},
		{"zero-only", "\u0301", " \u0301        "},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			canvas := mkDrawing(8, 0)
			canvas.drawText(drawingCoord{1, 0}, testCase.text)
			if got := drawingToString(canvas); got != testCase.want {
				t.Fatalf("display cells: got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestUnicodeNodeMeasurement(t *testing.T) {
	originalPadding := boxBorderPadding
	t.Cleanup(func() { boxBorderPadding = originalPadding })
	boxBorderPadding = 1
	testCases := []struct {
		label string
		width int
		row   string
	}{
		{"Café", 6, "│ Café │"},
		{"東京", 6, "│ 東京 │"},
		{"e\u0301", 3, "│ e\u0301 │"},
		{"👩‍💻", 4, "│ 👩‍💻 │"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			rendered, output := renderAuditGraph(t, "graph LR\nA["+testCase.label+"]", false, 5, HeadStylePointer)
			if got := rendered.columnWidth[1]; got != testCase.width {
				t.Errorf("label column: got %d cells, want %d", got, testCase.width)
			}
			if !strings.Contains(output, testCase.row) {
				t.Errorf("missing centered label %q:\n%s", testCase.row, output)
			}
		})
	}
}

func TestUnicodeEdgeLabelPlacement(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	testCases := []struct {
		label string
		want  string
	}{
		{"é", "────é────"},
		{"東京", "──東京───"},
		{"e\u0301", "────e\u0301────"},
		{"👩‍💻", "───👩‍💻────"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			base := mkDrawing(8, 0)
			for column := range *base {
				(*base)[column][0] = "─"
			}
			label := mkDrawing(8, 0)
			label.drawTextOnLine([]drawingCoord{{0, 0}, {8, 0}}, testCase.label, false)
			if got := drawingToString(mergeDrawings(base, drawingCoord{0, 0}, label)); got != testCase.want {
				t.Fatalf("label over route: got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestUnicodeGroupLabelPlacement(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	group := &subgraph{name: "東京é", minX: 0, maxX: 12, minY: 0, maxY: 4}
	label, _ := drawSubgraphLabel(group, graph{})
	if got := strings.Split(drawingToString(label), "\n")[1]; got != "    東京é    " {
		t.Fatalf("group label: got %q, want %q", got, "    東京é    ")
	}
}

func TestUnicodeGroupTitleFits(t *testing.T) {
	title := "東京東京東京東京"
	_, output := renderAuditGraph(t, "graph LR\nsubgraph "+title+"\nA\nend", false, 5, HeadStylePointer)
	if !strings.Contains(output, title) {
		t.Fatalf("group title was clipped:\n%s", output)
	}
}

func TestUnicodeGroupLabelClipsWholeGraphemes(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	group := &subgraph{name: "東👩‍💻", minX: 0, maxX: 4, minY: 0, maxY: 4}
	label, _ := drawSubgraphLabel(group, graph{})
	if got := strings.Split(drawingToString(label), "\n")[1]; got != " 東  " {
		t.Fatalf("clipped label: got %q, want %q", got, " 東  ")
	}
}

func TestUnicodeColoredNodeKeepsGraphemes(t *testing.T) {
	rendered, _ := renderAuditGraph(t, "graph LR\nA[👩‍💻東京e\u0301]", false, 5, HeadStylePointer)
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	rendered.styleType = "html"
	rendered.nodes[0].styleClass.styles = map[string]string{"color": "#00ff00"}
	output := drawingToString(drawBox(rendered.nodes[0], rendered))
	for _, cluster := range []string{"👩‍💻", "東", "京", "e\u0301"} {
		want := "<span style='color: #00ff00'>" + cluster + "</span>"
		if !strings.Contains(output, want) {
			t.Fatalf("styled grapheme missing %q:\n%s", want, output)
		}
	}
}

func TestASCIIUnicodeLabelsCompatibility(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = true
	canvas := mkDrawing(8, 0)
	canvas.drawText(drawingCoord{1, 0}, "Café")
	if got := drawingToString(canvas); got != " CafÃ©   " {
		t.Fatalf("ASCII byte drawing changed: %q", got)
	}
	group := &subgraph{name: "東京é", minX: 0, maxX: 12, minY: 0, maxY: 4}
	label, _ := drawSubgraphLabel(group, graph{})
	if got := strings.Split(drawingToString(label), "\n")[1]; got != "  東  京  é    " {
		t.Fatalf("ASCII group label changed: %q", got)
	}
}
