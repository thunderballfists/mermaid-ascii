package cmd

import (
	"fmt"
	"testing"
)

func TestUnicodeIncidentHeads(t *testing.T) {
	for _, source := range []string{
		"A<-->B\nB-->C\nD-->C",
		"A & B --> C & D",
		"A-->B\nB-->C\nA-->C\nB-->D\nD-->C",
		"A\nB\nB-->A\nA-->A\nB-->C\nC-->A",
	} {
		for _, direction := range []string{"LR", "TD"} {
			for _, gap := range []int{1, 5} {
				for _, boxX := range []int{0, 1} {
					for _, boxY := range []int{0, 1} {
						for _, uniform := range []bool{false, true} {
							for _, heads := range []HeadStyle{HeadStylePointer, HeadStyleOpen, HeadStyleWide, HeadStyleNone} {
								t.Run(fmt.Sprintf("%s/%s/gap=%d/box=%d,%d/uniform=%v/heads=%s", source, direction, gap, boxX, boxY, uniform, headStyleString(heads)), func(t *testing.T) {
									rendered, output := renderLayoutGraph(t, "graph "+direction+"\n"+source, gap, boxX, boxY, heads, uniform)
									if defects := auditRenderDefects(rendered, output, nil, false, heads); len(defects) != 0 {
										t.Fatalf("incident heads lost information: %v\n%s", defects, output)
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
}
