package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type edge struct {
	from          *node
	to            *node
	text          string
	nearHead      bool
	path          []gridCoord
	labelLine     []gridCoord
	startDir      direction
	endDir        direction
	edgeStyle     EdgeStyle
	arrowDir      arrowDirection
	leftCap       arrowEndCap
	rightCap      arrowEndCap
	lenHint       int
	separateRoute bool
	labelPosition *drawingCoord
}

func (g *graph) determinePath(e *edge) {
	if !useAscii && g.routeRepeatedEdge(e) {
		return
	}
	defer func() {
		if !useAscii && len(g.subgraphs) > 0 {
			g.routeRepeatedEdge(e)
		}
	}()
	// Get both paths and use least amount of steps
	var preferredPath, alternativePath []gridCoord
	var from, to gridCoord
	var err error
	preferredDir, preferredOppositeDir, alternativeDir, alternativeOppositeDir := determineStartAndEndDir(e)

	from = e.from.gridCoord.Direction(preferredDir)
	to = e.to.gridCoord.Direction(preferredOppositeDir)
	log.Debugf("Determining preferred path from %v (direction %v) to %v (direction %v)", *e.from, preferredDir, *e.to, preferredOppositeDir)

	// Get preferred path
	preferredPath, err = g.getPath(from, to)
	if err != nil {
		fmt.Printf("Error getting path from %v to %v: %v", from, to, err)
		// This is a big assumption, but if we can't get the preferred path, we assume the alternative path is better
		e.startDir = alternativeDir
		e.endDir = alternativeOppositeDir
		e.path = alternativePath
		return
	}
	preferredPath = mergePath(preferredPath)

	// Alternative path
	from = e.from.gridCoord.Direction(alternativeDir)
	to = e.to.gridCoord.Direction(alternativeOppositeDir)
	log.Debugf("Determining alternative path from %v (direction %v) to %v (direction %v)", *e.from, alternativeDir, *e.to, alternativeOppositeDir)

	alternativePath, err = g.getPath(from, to)
	if err != nil {
		fmt.Printf("Error getting path from %v to %v: %v", from, to, err)
		e.startDir = preferredDir
		e.endDir = preferredOppositeDir
		e.path = preferredPath
	}
	alternativePath = mergePath(alternativePath)

	nrStepsPreferred := len(preferredPath)
	nrStepsAlternative := len(alternativePath)
	if nrStepsPreferred <= nrStepsAlternative {
		log.Debugf("Using preferred path with %v steps instead of alternative path with %v steps", nrStepsPreferred, nrStepsAlternative)
		e.startDir = preferredDir
		e.endDir = preferredOppositeDir
		e.path = preferredPath
	} else {
		log.Debugf("Using alternative path with %v steps instead of alternative path with %v steps", nrStepsAlternative, nrStepsPreferred)
		e.startDir = alternativeDir
		e.endDir = alternativeOppositeDir
		e.path = alternativePath
	}
}

func (e *edge) flowOptions() FlowOptions {
	opts := defaultFlowOptions()
	switch e.edgeStyle {
	case StyleThick:
		opts.Thickness = ThicknessDouble
		opts.LineStyle = LineStyleSolid
	case StyleDotted:
		opts.LineStyle = LineStyleDotted
	case StyleDashed:
		opts.LineStyle = LineStyleDashed
	default:
		opts.LineStyle = LineStyleSolid
	}
	opts.Heads = headStyle
	return opts
}

func (g *graph) determineLabelLine(e *edge) {
	// What line on the path should the label be placed?
	lenLabel := textWidth(e.text)
	if lenLabel == 0 {
		return
	}
	prevStep := e.path[0]
	var largestLineSize int
	// Init to first line if we find nothing else
	largestLine := []gridCoord{prevStep, e.path[1]}
	largestLineSize = 0
	for _, step := range e.path[1:] {
		line := []gridCoord{gridCoord(prevStep), gridCoord(step)}
		lineWidth := g.calculateLineWidth(line)
		if e.separateRoute {
			endpoints := g.lineToDrawing(line)
			lineWidth = Abs(endpoints[0].x-endpoints[1].x) + Abs(endpoints[0].y-endpoints[1].y) - 1
		}
		if !e.separateRoute && lineWidth >= lenLabel {
			largestLine = line
			break
		} else if lineWidth > largestLineSize {
			largestLineSize = lineWidth
			largestLine = line
		}
		prevStep = step
	}

	var maxX, minX int
	if largestLine[0].x > largestLine[1].x {
		maxX = largestLine[0].x
		minX = largestLine[1].x
	} else {
		maxX = largestLine[1].x
		minX = largestLine[0].x
	}
	middleX := minX + (maxX-minX)/2
	if e.separateRoute && middleX%2 == 0 {
		if middleX > minX {
			middleX--
		} else {
			middleX++
		}
	}
	extraSpacing := 2
	switch {
	case lenLabel <= 2:
		extraSpacing = 9
	case lenLabel <= 3:
		extraSpacing = 8
	case lenLabel <= 5:
		extraSpacing = 4
	}
	desiredWidth := lenLabel + extraSpacing
	log.Debugf("Increasing column width for column %v from size %v to %v", middleX, g.columnWidth[middleX], desiredWidth)
	g.columnWidth[middleX] = Max(g.columnWidth[middleX], desiredWidth) // Wrap with dashes + arrowhead
	log.Debugf("New column sizes: %v", g.columnWidth)
	e.labelLine = largestLine
	if e.separateRoute && largestLine[0].y == largestLine[1].y {
		g.reserveLineClearance(largestLine, lenLabel+2)
	}
	if !useAscii && largestLine[0].x == largestLine[1].x && headStyle != HeadStyleNone && e.arrowDir != arrowDirOpen {
		g.reserveLineClearance(largestLine, 3)
	}
}

func (g *graph) reserveArrowHeadClearance(connection *edge) {
	if useAscii || headStyle == HeadStyleNone || connection.arrowDir == arrowDirOpen || len(connection.path) != 2 {
		return
	}
	if connection.arrowDir == arrowDirBoth {
		g.reserveLineClearance(connection.path, 2)
	}
	for _, other := range g.edges {
		hasOpposedHeads := other.arrowDir == connection.arrowDir || other.arrowDir == arrowDirBoth || connection.arrowDir == arrowDirBoth
		if other != connection && hasOpposedHeads && other.arrowDir != arrowDirOpen && len(other.path) == 2 && connection.path[0] == other.path[1] && connection.path[1] == other.path[0] {
			g.reserveLineClearance(connection.path, 2)
		}
	}
}

func (g *graph) reserveLineClearance(line []gridCoord, minimumCells int) {
	if useAscii || len(line) != 2 || line[0].Equals(line[1]) {
		return
	}
	endpoints := g.lineToDrawing(line)
	bandSizes := g.columnWidth
	startIndex, endIndex := line[0].x, line[1].x
	availableCells := Abs(endpoints[0].x-endpoints[1].x) - 1
	if line[0].x == line[1].x {
		bandSizes = g.rowHeight
		startIndex, endIndex = line[0].y, line[1].y
		availableCells = Abs(endpoints[0].y-endpoints[1].y) - 1
	} else if line[0].y != line[1].y {
		return
	}
	if availableCells >= minimumCells {
		return
	}
	middleBand := Min(startIndex, endIndex) + Abs(startIndex-endIndex)/2
	if middleBand%2 == 0 {
		middleBand++
		if middleBand > Max(startIndex, endIndex) {
			middleBand -= 2
		}
	}
	extraCells := minimumCells - availableCells
	if middleBand == startIndex || middleBand == endIndex {
		extraCells *= 2
	}
	bandSizes[middleBand] += extraCells
}

func (g graph) calculateLineWidth(line []gridCoord) int {
	totalSize := 0
	for _, c := range line {
		totalSize += g.columnWidth[c.x]
	}
	return totalSize
}
