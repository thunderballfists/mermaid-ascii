package cmd

func (g *graph) equalizeNodeGaps() {
	if useAscii || !g.uniformSpacing || len(g.nodes) < 2 {
		return
	}
	minX, maxX := g.nodes[0].gridCoord.x, g.nodes[0].gridCoord.x
	minY, maxY := g.nodes[0].gridCoord.y, g.nodes[0].gridCoord.y
	for _, contentNode := range g.nodes[1:] {
		minX, maxX = Min(minX, contentNode.gridCoord.x), Max(maxX, contentNode.gridCoord.x)
		minY, maxY = Min(minY, contentNode.gridCoord.y), Max(maxY, contentNode.gridCoord.y)
	}
	equalizeGapBands(g.columnWidth, minX, maxX)
	equalizeGapBands(g.rowHeight, minY, maxY)
}

func equalizeGapBands(bandSizes map[int]int, start, end int) {
	largestGap := 0
	for band := start + 3; band < end; band += 4 {
		trailingBorderSpace := Max(0, bandSizes[band-1]-1)
		largestGap = Max(largestGap, bandSizes[band]+trailingBorderSpace)
	}
	for band := start + 3; band < end; band += 4 {
		trailingBorderSpace := Max(0, bandSizes[band-1]-1)
		bandSizes[band] = largestGap - trailingBorderSpace
	}
}
