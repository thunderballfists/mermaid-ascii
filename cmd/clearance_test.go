package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestUnicodeBidirectionalClearance(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, padding := range []int{1, 2, 5} {
			for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
				t.Run(fmt.Sprintf("%s/padding=%d/heads=%s", direction, padding, headStyleString(heads)), func(t *testing.T) {
					rendered, output := renderAuditGraph(t, "graph "+direction+"\nA<-->B", false, padding, heads)
					if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
						t.Fatalf("bidirectional heads must occupy distinct cells: %v\n%s", defects, output)
					}
					gap := rendered.columnWidth[3]
					if direction == "TD" {
						gap = rendered.rowHeight[3]
					}
					minimum := padding
					if heads != HeadStyleNone && minimum < 2 {
						minimum = 2
					}
					if gap != minimum {
						t.Errorf("gap = %d; want %d", gap, minimum)
					}
				})
			}
		}
	}
}

func TestUnicodeVerticalLabelClearance(t *testing.T) {
	for _, operator := range []string{"-->", "<--", "<-->", "---"} {
		for _, label := range []string{"a long edge label", "é", "東京", "e\u0301", "👩‍💻"} {
			for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
				for _, padding := range []int{1, 2, 5} {
					t.Run(fmt.Sprintf("%s/%s/heads=%s/padding=%d", operator, label, headStyleString(heads), padding), func(t *testing.T) {
						rendered, output := renderAuditGraph(t, "graph TD\nA"+operator+"|"+label+"|B", false, padding, heads)
						if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
							t.Fatalf("vertical label must not overwrite heads: %v\n%s", defects, output)
						}
						labelRow := -1
						for rowIndex, row := range strings.Split(output, "\n") {
							if strings.TrimSpace(row) == label {
								labelRow = rowIndex
							}
						}
						if labelRow < 0 {
							t.Fatalf("want an intact label on its own row:\n%s", output)
						}
						if gap := rendered.rowHeight[3]; heads != HeadStyleNone && operator != "---" && gap < 3 {
							t.Errorf("label/head gap = %d; want at least 3", gap)
						}
					})
				}
			}
		}
	}
}

func TestUnicodeSingleHeadSpacingUnchanged(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		t.Run(direction, func(t *testing.T) {
			rendered, output := renderAuditGraph(t, "graph "+direction+"\nA-->B", false, 1, HeadStylePointer)
			gap := rendered.columnWidth[3]
			if direction == "TD" {
				gap = rendered.rowHeight[3]
			}
			if gap != 1 {
				t.Fatalf("single-head gap must stay 1, got %d:\n%s", gap, output)
			}
		})
	}
}

func TestClearanceASCIICompatibility(t *testing.T) {
	cases := []struct {
		source, expected string
	}{
		{"graph LR\nA<-->B", "+---+ +---+\n|   | |   |\n| A |>| B |\n|   | |   |\n+---+ +---+"},
		{"graph TD\nA-->|a long edge label|B", strings.Join([]string{
			"+-------------------+", "|                   |", "|         A         |",
			"|                   |", "+-------------------+", "  a long edge label  ",
			"+-------------------+", "|                   |", "|         B         |",
			"|                   |", "+-------------------+",
		}, "\n")},
	}
	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			_, output := renderAuditGraph(t, testCase.source, true, 1, HeadStylePointer)
			if output != testCase.expected {
				t.Fatalf("legacy ASCII changed: got %q; want %q", output, testCase.expected)
			}
		})
	}
}

func TestUnicodeClearanceRespectsPaddingAxes(t *testing.T) {
	cases := []struct {
		source             string
		paddingX, paddingY int
		gapX, gapY         int
	}{
		{"graph LR\nA<-->B\nA-->C", 1, 5, 2, 5},
		{"graph TD\nA<-->|é|B\nA-->C", 5, 1, 5, 3},
	}
	for _, testCase := range cases {
		t.Run(testCase.source, func(t *testing.T) {
			rendered, output := renderAuditGraphWithPadding(t, testCase.source, false, testCase.paddingX, testCase.paddingY, HeadStylePointer)
			if defects := auditRenderDefects(rendered, output, nil, false, HeadStylePointer); len(defects) != 0 {
				t.Fatalf("asymmetric spacing corrupted render: %v\n%s", defects, output)
			}
			if actualX, actualY := rendered.columnWidth[3], rendered.rowHeight[3]; actualX != testCase.gapX || actualY != testCase.gapY {
				t.Fatalf("gaps = (%d,%d); want (%d,%d)", actualX, actualY, testCase.gapX, testCase.gapY)
			}
		})
	}
}

func TestReserveLineClearanceBandCenters(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	for _, vertical := range []bool{false, true} {
		for _, endpointBand := range []bool{false, true} {
			for _, reversed := range []bool{false, true} {
				t.Run(fmt.Sprintf("vertical=%v/endpoint=%v/reversed=%v", vertical, endpointBand, reversed), func(t *testing.T) {
					rendered := graph{columnWidth: map[int]int{0: 1, 1: 1, 2: 1}, rowHeight: map[int]int{0: 1, 1: 1, 2: 1}}
					endIndex := 2
					if endpointBand {
						endIndex = 1
					}
					line := []gridCoord{{0, 0}, {endIndex, 0}}
					if vertical {
						line[1] = gridCoord{0, endIndex}
					}
					if reversed {
						line[0], line[1] = line[1], line[0]
					}
					rendered.reserveLineClearance(line, 3)
					endpoints := rendered.lineToDrawing(line)
					if available := Abs(endpoints[0].x-endpoints[1].x) + Abs(endpoints[0].y-endpoints[1].y) - 1; available != 3 {
						t.Fatalf("drawable cells = %d; want 3", available)
					}
					rendered.reserveLineClearance(line, 3)
					if after := rendered.lineToDrawing(line); after[0] != endpoints[0] || after[1] != endpoints[1] {
						t.Fatalf("repeated reservation moved endpoints: %v -> %v", endpoints, after)
					}
				})
			}
		}
	}
}
