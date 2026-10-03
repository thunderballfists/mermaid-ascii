package cmd

func (g *graph) reserveRepeatedPortMargins() {
	if headStyle == HeadStyleNone {
		return
	}
	heads := map[*node]int{}
	repeated := map[*node]bool{}
	pairs := map[[2]int]bool{}
	for _, connection := range g.edges {
		pair := [2]int{Min(connection.from.index, connection.to.index), Max(connection.from.index, connection.to.index)}
		if pairs[pair] && connection.from != connection.to {
			repeated[connection.from], repeated[connection.to] = true, true
		}
		pairs[pair] = true
		if connection.arrowDir == arrowDirBackward || connection.arrowDir == arrowDirBoth {
			heads[connection.from]++
		}
		if connection.arrowDir == arrowDirForward || connection.arrowDir == arrowDirBoth {
			heads[connection.to]++
		}
	}
	shiftX, shiftY := 0, 0
	for _, contentNode := range g.nodes {
		capacity := 4
		if contentNode.gridCoord.x == 0 {
			capacity--
		}
		if contentNode.gridCoord.y == 0 {
			capacity--
		}
		if (repeated[contentNode] || len(g.subgraphs) > 0) && heads[contentNode] > capacity {
			if contentNode.gridCoord.x == 0 {
				shiftX = 4
				capacity++
			}
			if heads[contentNode] > capacity && contentNode.gridCoord.y == 0 {
				shiftY = 4
			}
		}
	}
	if shiftX == 0 && shiftY == 0 {
		return
	}
	grid := map[gridCoord]*node{}
	for position, contentNode := range g.grid {
		grid[gridCoord{position.x + shiftX, position.y + shiftY}] = contentNode
	}
	g.grid = grid
	for _, contentNode := range g.nodes {
		contentNode.gridCoord.x += shiftX
		contentNode.gridCoord.y += shiftY
	}
}

func (g *graph) routeRepeatedEdge(connection *edge) bool {
	if connection.from == connection.to {
		return false
	}
	previous := []*edge{}
	for _, other := range g.edges {
		if other == connection || len(other.path) == 0 {
			continue
		}
		samePair := other.from == connection.from && other.to == connection.to
		reversedPair := other.from == connection.to && other.to == connection.from
		conflictingHead := len(connection.path) > 0 && headStyle != HeadStyleNone && sharesHeadPort(connection, other)
		if samePair || reversedPair && (connection.text != "" || other.text != "" || connection.arrowDir != other.arrowDir) || conflictingHead {
			previous = append(previous, other)
		}
	}
	if len(previous) == 0 {
		return false
	}
	originalPath := connection.path
	connection.path = nil
	defer func() {
		if len(connection.path) == 0 {
			connection.path = originalPath
		}
	}()
	ports := []direction{Down, Right, Up, Left}
	if graphDirection == "TD" {
		ports = []direction{Right, Down, Left, Up}
	}
	routeCosts := map[gridCoord]int{}
	for _, other := range previous {
		for segment := 1; segment < len(other.path); segment++ {
			from, to := other.path[segment-1], other.path[segment]
			step := gridCoord{}
			if from.x != to.x {
				step.x = (to.x - from.x) / Abs(to.x-from.x)
			} else {
				step.y = (to.y - from.y) / Abs(to.y-from.y)
			}
			for point := from; point != to; point = (gridCoord{point.x + step.x, point.y + step.y}) {
				routeCosts[point] += 4
			}
		}
	}
	for _, startPort := range ports {
		for _, endPort := range ports {
			if !safeRepeatedPort(connection.from, startPort) || !safeRepeatedPort(connection.to, endPort) {
				continue
			}
			startHead := connection.arrowDir == arrowDirBackward || connection.arrowDir == arrowDirBoth
			endHead := connection.arrowDir == arrowDirForward || connection.arrowDir == arrowDirBoth
			if headStyle != HeadStyleNone && (startHead && g.occupiedHeadPort(connection.from, startPort) || endHead && g.occupiedHeadPort(connection.to, endPort)) {
				continue
			}
			available := true
			for _, other := range previous {
				otherStart, otherEnd := other.startDir, other.endDir
				if other.from != connection.from {
					otherStart, otherEnd = otherEnd, otherStart
				}
				if startPort == otherStart && endPort == otherEnd {
					available = false
					break
				}
			}
			if !available {
				continue
			}
			from := connection.from.gridCoord.Direction(startPort)
			to := connection.to.gridCoord.Direction(endPort)
			path, err := g.getPathWithCosts(from, to, routeCosts)
			if err != nil {
				continue
			}
			connection.startDir, connection.endDir = startPort, endPort
			connection.path = mergePath(path)
			connection.separateRoute = true
			return true
		}
	}
	return false
}

func sharesHeadPort(connection, other *edge) bool {
	startHead := connection.arrowDir == arrowDirBackward || connection.arrowDir == arrowDirBoth
	endHead := connection.arrowDir == arrowDirForward || connection.arrowDir == arrowDirBoth
	otherStartHead := other.arrowDir == arrowDirBackward || other.arrowDir == arrowDirBoth
	otherEndHead := other.arrowDir == arrowDirForward || other.arrowDir == arrowDirBoth
	return startHead && (otherStartHead && connection.from == other.from && connection.startDir == other.startDir || otherEndHead && connection.from == other.to && connection.startDir == other.endDir) ||
		endHead && (otherStartHead && connection.to == other.from && connection.endDir == other.startDir || otherEndHead && connection.to == other.to && connection.endDir == other.endDir)
}

func safeRepeatedPort(contentNode *node, port direction) bool {
	return (port != Up || contentNode.gridCoord.y > 0) && (port != Left || contentNode.gridCoord.x > 0)
}

func (g *graph) occupiedHeadPort(contentNode *node, port direction) bool {
	for _, connection := range g.edges {
		if len(connection.path) == 0 {
			continue
		}
		startHead := connection.arrowDir == arrowDirBackward || connection.arrowDir == arrowDirBoth || connection.text != ""
		endHead := connection.arrowDir == arrowDirForward || connection.arrowDir == arrowDirBoth || connection.text != ""
		if connection.from == contentNode && connection.startDir == port && startHead || connection.to == contentNode && connection.endDir == port && endHead {
			return true
		}
	}
	return false
}

func (g *graph) uniqueEdgesFirst() []*edge {
	seen := map[[2]int]bool{}
	unique, repeated := []*edge{}, []*edge{}
	for _, connection := range g.edges {
		pair := [2]int{Min(connection.from.index, connection.to.index), Max(connection.from.index, connection.to.index)}
		if seen[pair] {
			repeated = append(repeated, connection)
		} else {
			unique = append(unique, connection)
		}
		seen[pair] = true
	}
	return append(unique, repeated...)
}
