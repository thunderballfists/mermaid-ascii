package cmd

import (
	"fmt"
	"testing"
)

func TestGroupLanesPreserveRanksAndGrid(t *testing.T) {
	for _, direction := range []string{"LR", "TD"} {
		for _, ascii := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/ascii=%v", direction, ascii), func(t *testing.T) {
				originalASCII, originalDirection := useAscii, graphDirection
				t.Cleanup(func() { useAscii, graphDirection = originalASCII, originalDirection })
				useAscii, graphDirection = ascii, direction
				rendered := graph{grid: map[gridCoord]*node{}}
				for index, originalPosition := range []gridCoord{{0, 0}, {4, 0}, {0, 4}, {0, 8}, {4, 4}} {
					position := originalPosition
					if direction == "TD" {
						position.x, position.y = position.y, position.x
					}
					contentNode := &node{name: fmt.Sprint(index), gridCoord: &position}
					rendered.nodes = append(rendered.nodes, contentNode)
					rendered.reserveSpotInGrid(contentNode, contentNode.gridCoord)
				}
				rendered.subgraphs = []*subgraph{
					{nodes: rendered.nodes[:3]},
					{nodes: rendered.nodes[3:]},
				}
				before := map[*node]gridCoord{}
				for _, contentNode := range rendered.nodes {
					before[contentNode] = *contentNode.gridCoord
				}
				rendered.separateGroupLanes()
				if rendered.groupLanes == ascii {
					t.Fatalf("lane separation=%v for ascii=%v", rendered.groupLanes, ascii)
				}
				for _, contentNode := range rendered.nodes {
					previous, position := before[contentNode], *contentNode.gridCoord
					if ascii && previous != position || direction == "LR" && previous.x != position.x || direction == "TD" && previous.y != position.y {
						t.Errorf("node %s changed primary rank or ASCII position: %v -> %v", contentNode, previous, position)
					}
					for column := 0; column < 3; column++ {
						for row := 0; row < 3; row++ {
							if rendered.grid[gridCoord{position.x + column, position.y + row}] != contentNode {
								t.Errorf("node %s has stale grid occupancy", contentNode)
							}
						}
					}
					before[contentNode] = position
				}
				if !ascii && rendered.hasInterleavedGroups() {
					t.Fatal("separated groups still interleave")
				}
				rendered.separateGroupLanes()
				for _, contentNode := range rendered.nodes {
					if *contentNode.gridCoord != before[contentNode] {
						t.Errorf("second separation changed node %s", contentNode)
					}
				}
			})
		}
	}
}

func TestUnicodeNoninterleavedGroupLanesUnchanged(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	for _, emptyGroup := range []bool{false, true} {
		rendered := graph{nodes: []*node{
			{gridCoord: &gridCoord{0, 0}},
			{gridCoord: &gridCoord{4, 0}},
			{gridCoord: &gridCoord{0, 4}},
		}}
		rendered.subgraphs = []*subgraph{{nodes: rendered.nodes[:2]}}
		if emptyGroup {
			rendered.subgraphs = append(rendered.subgraphs, &subgraph{})
		}
		rendered.separateGroupLanes()
		if rendered.groupLanes || *rendered.nodes[2].gridCoord != (gridCoord{0, 4}) {
			t.Fatal("valid group placement was changed")
		}
	}
}
