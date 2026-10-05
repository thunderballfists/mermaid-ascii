package cmd

import (
	"fmt"
	"testing"
)

func TestUnicodeLabelsKeepSpaces(t *testing.T) {
	for _, source := range []string{
		"A-->|a long edge label|B",
		"A-->|Café 東京|B",
		"A-.->|👩‍💻 returns|B",
		"A==>|┼ ─ ┤|B",
		"A-->|first label|B\nA-->|second label|B",
		"A<-->|loop label|A\nA-->B",
		"subgraph Outer\nA-->|inside label|B\nend\nB-->|outside label|C",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxY := range []int{0, 1} {
					for _, uniform := range []bool{false, true} {
						for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
							t.Run(fmt.Sprintf("%s/%s/gap=%d/boxY=%d/uniform=%v/heads=%s", source, direction, gap, boxY, uniform, headStyleString(heads)), func(t *testing.T) {
								rendered, output := renderLayoutGraph(t, "graph "+direction+"\n"+source, gap, 1, boxY, heads, uniform)
								if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
									t.Fatalf("label text changed during compositing: %v\n%s", defects, output)
								}
								assertNodeInteriors(t, rendered, output)
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
