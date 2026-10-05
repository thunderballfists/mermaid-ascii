package cmd

func (g *graph) placeRepeatedRouteLabels() {
	hasRepeated := false
	for _, connection := range g.edges {
		hasRepeated = hasRepeated || connection.separateRoute
	}
	if !hasRepeated && len(g.subgraphs) == 0 {
		return
	}
	for attempt := 0; attempt <= len(g.edges)*4; attempt++ {
		occupied := map[drawingCoord]bool{}
		if len(g.subgraphs) > 0 {
			g.calculateSubgraphBoundingBoxes()
			for _, group := range g.subgraphs {
				if len(group.nodes) == 0 {
					continue
				}
				width := textWidth(group.name)
				left := group.minX + Max(1, (group.maxX-group.minX)/2-width/2)
				for column := 0; column < width; column++ {
					occupied[drawingCoord{left + column, group.minY + 1}] = true
				}
			}
		}
		for _, contentNode := range g.nodes {
			width, height := getDrawingSize(contentNode.drawing)
			for column := 0; column <= width; column++ {
				for row := 0; row <= height; row++ {
					occupied[drawingCoord{contentNode.drawingCoord.x + column, contentNode.drawingCoord.y + row}] = true
				}
			}
		}
		for _, connection := range g.edges {
			connection.labelPosition = nil
			_, lines, directions := g.drawPath(connection.path, connection)
			heads := g.drawArrowHead(lines, directions, connection)
			g.applyArrowEndCaps(heads, lines, connection)
			for column, cells := range *heads {
				for row, glyph := range cells {
					if glyph != " " {
						occupied[drawingCoord{column, row}] = true
					}
				}
			}
		}
		resized := false
		for _, connection := range g.edges {
			if connection.text == "" {
				continue
			}
			labelWidth := Max(1, textWidth(connection.text))
			candidates := []drawingCoord{}
			if position, available := labelPositionOnLine(g.lineToDrawing(connection.labelLine), connection.text, connection.nearHead); available {
				candidates = append(candidates, position)
			}
			for segment := 1; segment < len(connection.path); segment++ {
				endpoints := g.lineToDrawing(connection.path[segment-1 : segment+1])
				if endpoints[0].y == endpoints[1].y {
					for column := Min(endpoints[0].x, endpoints[1].x) + 1; column+labelWidth <= Max(endpoints[0].x, endpoints[1].x); column++ {
						candidates = append(candidates, drawingCoord{column, endpoints[0].y})
					}
				} else {
					for row := Min(endpoints[0].y, endpoints[1].y) + 1; row < Max(endpoints[0].y, endpoints[1].y); row++ {
						candidates = append(candidates, drawingCoord{Max(0, endpoints[0].x-labelWidth/2), row})
					}
				}
			}
			for _, position := range candidates {
				free := position.x >= 0 && position.y >= 0
				for offset := 0; offset < labelWidth && free; offset++ {
					free = !occupied[drawingCoord{position.x + offset, position.y}]
				}
				if free {
					chosen := position
					connection.labelPosition = &chosen
					for offset := 0; offset < labelWidth; offset++ {
						occupied[drawingCoord{position.x + offset, position.y}] = true
					}
					break
				}
			}
			if connection.labelPosition == nil {
				endpoints := g.lineToDrawing(connection.labelLine)
				span := Abs(endpoints[0].x-endpoints[1].x) + Abs(endpoints[0].y-endpoints[1].y) - 1
				extra := 1
				if endpoints[0].y == endpoints[1].y {
					extra = labelWidth + 2
				}
				g.reserveLineClearance(connection.labelLine, span+extra)
				resized = true
				break
			}
		}
		if !resized {
			return
		}
		g.equalizeNodeGaps()
		for _, contentNode := range g.nodes {
			position := g.gridToDrawingCoord(*contentNode.gridCoord, nil)
			contentNode.setCoord(&position)
			contentNode.setDrawing(*g)
		}
		g.setDrawingSizeToGridConstraints()
	}
}
