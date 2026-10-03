package cmd

func (g *graph) reserveGroupClearance() {
	if len(g.subgraphs) == 0 {
		return
	}
	membership := map[*node]map[*subgraph]bool{}
	for _, contentNode := range g.nodes {
		membership[contentNode] = map[*subgraph]bool{}
	}
	for _, group := range g.subgraphs {
		for _, contentNode := range group.nodes {
			membership[contentNode][group] = true
		}
	}
	for _, before := range g.nodes {
		for _, after := range g.nodes {
			horizontal := before.gridCoord.y == after.gridCoord.y && before.gridCoord.x+4 == after.gridCoord.x
			vertical := before.gridCoord.x == after.gridCoord.x && before.gridCoord.y+4 == after.gridCoord.y
			if !horizontal && !vertical {
				continue
			}
			exited, entered := 0, 0
			for group := range membership[before] {
				if !membership[after][group] {
					exited++
				}
			}
			for group := range membership[after] {
				if !membership[before][group] {
					entered++
				}
			}
			if exited == 0 && entered == 0 {
				continue
			}
			if horizontal {
				band := after.gridCoord.x - 1
				g.columnWidth[band] = Max(g.columnWidth[band], 2*exited+2*entered+1)
			} else {
				band := after.gridCoord.y - 1
				g.rowHeight[band] = Max(g.rowHeight[band], 2*exited+4*entered+1)
			}
		}
	}
	groups := g.sortSubgraphsByDepth()
	for index := len(groups) - 1; index >= 0; index-- {
		group := groups[index]
		if len(group.nodes) == 0 {
			continue
		}
		left, right := group.nodes[0].gridCoord.x, group.nodes[0].gridCoord.x+2
		for _, contentNode := range group.nodes {
			left = Min(left, contentNode.gridCoord.x)
			right = Max(right, contentNode.gridCoord.x+2)
		}
		span := g.gridToDrawingCoord(gridCoord{x: right}, nil).x - g.gridToDrawingCoord(gridCoord{x: left}, nil).x
		extra := Max(0, textWidth(group.name)+2-(span+4))
		g.columnWidth[left+1] += extra
	}
}
