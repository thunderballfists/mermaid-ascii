package cmd

import (
	"container/heap"
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
)

type EdgeStyle int

const (
	StyleSolid EdgeStyle = iota
	StyleDotted
	StyleDashed
	StyleThick
)

type arrowDirection int

const (
	arrowDirOpen arrowDirection = iota
	arrowDirForward
	arrowDirBackward
	arrowDirBoth
)

type arrowEndCap int

const (
	arrowCapNone arrowEndCap = iota
	arrowCapCircle
	arrowCapCross
	arrowCapDiamond
)

type priorityQueueItem struct {
	coord    gridCoord
	priority int
	index    int
}

type priorityQueue []*priorityQueueItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*priorityQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func heuristic(a, b gridCoord) int {
	absX := Abs(a.x - b.x)
	absY := Abs(a.y - b.y)
	if absX == 0 || absY == 0 {
		return absX + absY
	} else {
		// Punish for taking extra corner, we prefer straight (less complex) lines
		return absX + absY + 1
	}
}

func (g *graph) getPath(from gridCoord, to gridCoord) ([]gridCoord, error) {
	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &priorityQueueItem{coord: from, priority: 0})

	costSoFar := map[gridCoord]int{from: 0}
	cameFrom := map[gridCoord]*gridCoord{from: nil}

	directions := []gridCoord{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*priorityQueueItem).coord

		if current.Equals(to) {
			path := []gridCoord{}
			for c := &current; c != nil; c = cameFrom[*c] {
				path = append([]gridCoord{*c}, path...)
			}
			log.Debugf("Found it! path: %v", path)
			return path, nil
		}

		for _, dir := range directions {
			next := gridCoord{x: current.x + dir.x, y: current.y + dir.y}
			if !g.isFreeInGrid(next) && !next.Equals(to) {
				continue
			}

			newCost := costSoFar[current] + 1
			if cost, ok := costSoFar[next]; !ok || newCost < cost {
				costSoFar[next] = newCost
				priority := newCost + heuristic(next, to)
				heap.Push(pq, &priorityQueueItem{coord: next, priority: priority})
				cameFrom[next] = &current
			}
		}
	}

	return nil, fmt.Errorf("no path found")
}

func (g *graph) isFreeInGrid(c gridCoord) bool {
	// We'll fix it later if we overshoot the grid size
	if c.x < 0 || c.y < 0 {
		return false
	}
	return g.grid[c] == nil
}

func (g *graph) drawArrow(from gridCoord, to gridCoord, e *edge) (*drawing, *drawing, *drawing, *drawing, *drawing) {
	if len(e.path) == 0 {
		return nil, nil, nil, nil, nil
	}
	log.Debugf("Drawing arrow from %v to %v with path %v", from, to, e.path)
	dLabel := g.drawArrowLabel(e)
	dPath, linesDrawn, lineDirs := g.drawPath(e.path, e)
	if !useAscii {
		g.applyArrowBodyStyle(dPath, linesDrawn, lineDirs, e, dLabel)
	}
	dBoxStart := g.drawBoxStart(e, linesDrawn[0])
	dArrowHead := g.drawArrowHead(linesDrawn, lineDirs, e)
	if !useAscii {
		g.applyArrowEndCaps(dArrowHead, linesDrawn, e)
	}
	dCorners := g.drawCorners(e)
	return dPath, dBoxStart, dArrowHead, dCorners, dLabel
}

func mergePath(path []gridCoord) []gridCoord {
	// If two steps are in the same direction, merge them to one step.
	if len(path) <= 2 {
		return path
	}
	indexToRemove := []int{}
	step0 := path[0]
	step1 := path[1]
	for idx, step2 := range path[2:] {
		prevDir := determineDirection(genericCoord(step0), genericCoord(step1))
		dir := determineDirection(genericCoord(step1), genericCoord(step2))
		if prevDir == dir {
			indexToRemove = append(indexToRemove, idx+1) // +1 because we skip the initial step
		}
		step0 = step1
		step1 = step2
	}
	newPath := []gridCoord{}
	for idx, step := range path {
		if !slices.Contains(indexToRemove, idx) {
			newPath = append(newPath, step)
		}
	}
	log.Debugf("Reduced path from %v to %v", path, newPath)
	return newPath
}

func (g *graph) drawPath(path []gridCoord, e *edge) (*drawing, [][]drawingCoord, []direction) {
	d := copyCanvas(g.drawing)
	previousCoord := path[0]
	linesDrawn := make([][]drawingCoord, 0)
	lineDirs := make([]direction, 0)
	var previousDrawingCoord drawingCoord
	for _, nextCoord := range path[1:] {
		previousDrawingCoord = g.gridToDrawingCoord(previousCoord, nil)
		nextDrawingCoord := g.gridToDrawingCoord(nextCoord, nil)
		if previousDrawingCoord.Equals(nextDrawingCoord) {
			log.Debugf("Skipping drawing identical line on %v", nextCoord)
			continue
		}
		dir := determineDirection(genericCoord(previousCoord), genericCoord(nextCoord))
		s := d.drawLine(previousDrawingCoord, nextDrawingCoord, 1, -1, e)
		if len(s) == 0 {
			// drawLine may return no coords if offsets collapse the line. Use at least one point so arrow and junction logic
			// can still infer a direction.
			s = append(s, previousDrawingCoord)
		}
		linesDrawn = append(linesDrawn, s)
		lineDirs = append(lineDirs, dir)
		previousCoord = nextCoord
	}
	return d, linesDrawn, lineDirs
}

func (g *graph) drawBoxStart(e *edge, firstLine []drawingCoord) *drawing {
	d := *(copyCanvas(g.drawing))
	if len(e.path) < 2 {
		return &d
	}
	from := firstLine[0]
	dir := determineDirection(genericCoord(e.path[0]), genericCoord(e.path[1]))
	log.Debugf("Drawing box start at %v with direction %v for line %v", from, dir, e.path)

	if useAscii {
		return &d
	}

	pal := buildPalette(e.flowOptions())
	switch dir {
	case Up:
		d[from.x][from.y+1] = pal.TeeDown
	case Down:
		d[from.x][from.y-1] = pal.TeeUp
	case Left:
		d[from.x+1][from.y] = pal.TeeRight
	case Right:
		d[from.x-1][from.y] = pal.TeeLeft
	}
	return &d
}

func (g *graph) drawArrowHead(lines [][]drawingCoord, lineDirs []direction, e *edge) *drawing {
	d := *(copyCanvas(g.drawing))
	if len(lines) == 0 {
		return &d
	}

	lastLine := lines[len(lines)-1]
	lastFallback := Middle
	if len(lineDirs) > 0 {
		lastFallback = lineDirs[len(lineDirs)-1]
	}

	if useAscii {
		if len(lastLine) == 0 {
			return &d
		}
		from := lastLine[0]
		lastPos := lastLine[len(lastLine)-1]
		dir := determineDirection(genericCoord(from), genericCoord(lastPos))
		if len(lastLine) == 1 || dir == Middle {
			dir = lastFallback
		}
		char := asciiArrowHead(dir, lastFallback)
		d[lastPos.x][lastPos.y] = char
		return &d
	}

	pal := buildPalette(e.flowOptions())
	endHasHead := e.arrowDir == arrowDirForward || e.arrowDir == arrowDirBoth
	startHasHead := e.arrowDir == arrowDirBackward || e.arrowDir == arrowDirBoth

	if endHasHead && len(lastLine) > 0 {
		dir := resolveHeadDirection(lastLine, lastFallback, false)
		char := pal.arrowHead(dir, lastFallback)
		if char != "" {
			target := lastLine[len(lastLine)-1]
			d[target.x][target.y] = char
		}
	}

	if startHasHead && len(lines) > 0 {
		firstLine := lines[0]
		if len(firstLine) > 0 {
			firstFallback := Middle
			if len(lineDirs) > 0 {
				firstFallback = lineDirs[0]
			}
			startFallback := firstFallback.getOpposite()
			dir := resolveHeadDirection(firstLine, startFallback, true)
			char := pal.arrowHead(dir, startFallback)
			if char != "" {
				start := firstLine[0]
				d[start.x][start.y] = char
			}
		}
	}

	return &d
}

func asciiArrowHead(dir direction, fallback direction) string {
	switch dir {
	case Up:
		return "^"
	case Down:
		return "v"
	case Left:
		return "<"
	case Right:
		return ">"
	default:
		switch fallback {
		case Up:
			return "^"
		case Down:
			return "v"
		case Left:
			return "<"
		case Right:
			return ">"
		default:
			return "*"
		}
	}
}

func resolveHeadDirection(line []drawingCoord, fallback direction, atStart bool) direction {
	if len(line) >= 2 {
		from := line[0]
		to := line[len(line)-1]
		if atStart {
			from, to = to, from
		}
		dir := determineDirection(genericCoord(from), genericCoord(to))
		if dir != Middle {
			return dir
		}
	}
	return fallback
}

func (g *graph) applyArrowBodyStyle(d *drawing, lines [][]drawingCoord, dirs []direction, e *edge, labelDrawing *drawing) {
	if e.edgeStyle != StyleDotted {
		return
	}

	labelCoords := map[drawingCoord]struct{}{}
	if e.text != "" && labelDrawing != nil {
		maxX, maxY := getDrawingSize(labelDrawing)
		for x := 0; x <= maxX; x++ {
			for y := 0; y <= maxY; y++ {
				if (*labelDrawing)[x][y] != " " {
					labelCoords[drawingCoord{x, y}] = struct{}{}
				}
			}
		}
	}

	junctionCounts := make(map[drawingCoord]int)
	for _, line := range lines {
		for _, coord := range line {
			junctionCounts[coord]++
		}
	}

	for idx, line := range lines {
		if idx >= len(dirs) {
			break
		}
		dir := dirs[idx]
		if dir != Left && dir != Right {
			continue
		}
		shouldDot := false
		for _, coord := range line {
			if (*d)[coord.x][coord.y] != "─" {
				continue
			}
			if _, ok := labelCoords[coord]; ok {
				continue
			}
			if junctionCounts[coord] > 1 {
				continue
			}
			if shouldDot {
				(*d)[coord.x][coord.y] = "·"
			}
			shouldDot = !shouldDot
		}
	}
}

func (g *graph) applyArrowEndCaps(d *drawing, lines [][]drawingCoord, e *edge) {
	if (e.leftCap == arrowCapNone && e.rightCap == arrowCapNone) || len(lines) == 0 {
		return
	}

	if e.leftCap != arrowCapNone {
		hasHead := e.arrowDir == arrowDirBackward || e.arrowDir == arrowDirBoth
		firstLine := lines[0]
		g.placeEndCap(d, firstLine, e.leftCap, true, hasHead)
	}

	if e.rightCap != arrowCapNone {
		hasHead := e.arrowDir == arrowDirForward || e.arrowDir == arrowDirBoth
		lastLine := lines[len(lines)-1]
		g.placeEndCap(d, lastLine, e.rightCap, false, hasHead)
	}
}

func (g *graph) placeEndCap(d *drawing, line []drawingCoord, cap arrowEndCap, atStart bool, hasHead bool) {
	if len(line) == 0 {
		return
	}

	char := capRune(cap)
	if char == "" {
		return
	}

	target, ok := g.endCapTarget(line, atStart, hasHead)
	if !ok {
		return
	}

	if target.x < 0 || target.y < 0 {
		return
	}

	d.increaseSize(target.x, target.y)
	(*d)[target.x][target.y] = char
}

func (g *graph) endCapTarget(line []drawingCoord, atStart bool, hasHead bool) (drawingCoord, bool) {
	if len(line) == 0 {
		return drawingCoord{}, false
	}

	idx := 0
	var dir direction
	if atStart {
		idx = 0
		if hasHead && len(line) > 1 {
			idx = 1
		}
		base := line[idx]
		if idx+1 < len(line) {
			dir = determineDirection(genericCoord(line[idx]), genericCoord(line[idx+1]))
		} else if idx > 0 {
			dir = determineDirection(genericCoord(line[idx-1]), genericCoord(line[idx]))
		} else {
			dir = Middle
		}
		moveDir := dir.getOpposite()
		dx, dy, ok := directionStep(moveDir)
		if !ok {
			return drawingCoord{}, false
		}
		return drawingCoord{base.x + dx, base.y + dy}, true
	}

	idx = len(line) - 1
	if hasHead && len(line) > 1 {
		idx = len(line) - 2
	}
	if idx < 0 {
		idx = 0
	}
	base := line[idx]
	if idx > 0 {
		dir = determineDirection(genericCoord(line[idx-1]), genericCoord(line[idx]))
	} else if idx+1 < len(line) {
		dir = determineDirection(genericCoord(line[idx]), genericCoord(line[idx+1]))
	} else {
		dir = Middle
	}
	dx, dy, ok := directionStep(dir)
	if !ok {
		return drawingCoord{}, false
	}
	return drawingCoord{base.x + dx, base.y + dy}, true
}

func directionStep(dir direction) (int, int, bool) {
	switch dir {
	case Up:
		return 0, -1, true
	case Down:
		return 0, 1, true
	case Left:
		return -1, 0, true
	case Right:
		return 1, 0, true
	default:
		return 0, 0, false
	}
}

func capRune(cap arrowEndCap) string {
	switch cap {
	case arrowCapCircle:
		return "○"
	case arrowCapCross:
		return "×"
	case arrowCapDiamond:
		return "◇"
	default:
		return ""
	}
}

func (g *graph) drawCorners(e *edge) *drawing {
	d := copyCanvas(g.drawing)
	path := e.path
	var pal palette
	if !useAscii {
		pal = buildPalette(e.flowOptions())
	}
	for idx, coord := range path {
		// Skip the first and last step
		if idx == 0 || idx == len(path)-1 {
			continue
		}
		drawingCoord := g.gridToDrawingCoord(coord, nil)

		prevDir := determineDirection(genericCoord(path[idx-1]), genericCoord(coord))
		nextDir := determineDirection(genericCoord(coord), genericCoord(path[idx+1]))

		var corner string
		if useAscii {
			corner = "+"
		} else {
			switch {
			case (prevDir == Right && nextDir == Down) || (prevDir == Up && nextDir == Left):
				corner = pal.CornerTopRight
			case (prevDir == Right && nextDir == Up) || (prevDir == Down && nextDir == Left):
				corner = pal.CornerBottomRight
			case (prevDir == Left && nextDir == Down) || (prevDir == Up && nextDir == Right):
				corner = pal.CornerTopLeft
			case (prevDir == Left && nextDir == Up) || (prevDir == Down && nextDir == Right):
				corner = pal.CornerBottomLeft
			default:
				corner = pal.TeeCross
			}
		}

		(*d)[drawingCoord.x][drawingCoord.y] = corner
	}
	return d
}

func (g *graph) drawArrowLabel(e *edge) *drawing {
	d := copyCanvas(g.drawing)
	lenLabel := textWidth(e.text)
	if lenLabel == 0 {
		return d
	}

	log.Debugf("Drawing text '%s' on gridline %v", e.text, e.labelLine)
	d.drawTextOnLine(g.lineToDrawing(e.labelLine), e.text, e.nearHead)
	return d
}

func (d *drawing) drawTextOnLine(line []drawingCoord, label string, nearHead bool) {
	// Write text in middle of the line
	//  123456789
	// |---------|
	//     123
	log.Debugf("Drawing text '%s' on drawingline %v", label, line)
	labelWidth := textWidth(label)
	if len(line) == 0 {
		return
	}
	var minX, maxX, minY, maxY int
	if line[0].x > line[1].x {
		minX = line[1].x
		maxX = line[0].x
	} else {
		minX = line[0].x
		maxX = line[1].x
	}
	if line[0].y > line[1].y {
		minY = line[1].y
		maxY = line[0].y
	} else {
		minY = line[0].y
		maxY = line[1].y
	}
	middleX := minX + (maxX-minX)/2
	middleY := minY + (maxY-minY)/2
	if minY == maxY {
		span := maxX - minX
		if span < labelWidth+2 {
			// Not enough room to place the label with padding on both sides
			return
		}

		available := span - labelWidth
		leftLen := available / 2
		rightLen := available - leftLen

		if nearHead {
			if line[1].x != line[0].x {
				biased := available * 4 / 5
				if biased >= available {
					biased = available - 1
				}
				if biased <= 0 {
					biased = 1
				}
				if line[1].x > line[0].x {
					leftLen = biased
					rightLen = available - leftLen
				} else {
					rightLen = biased
					leftLen = available - rightLen
				}
				if leftLen < 1 {
					leftLen = 1
					rightLen = available - leftLen
				}
				if rightLen < 1 {
					rightLen = 1
					leftLen = available - rightLen
				}
			}
		} else {
			if line[1].x > line[0].x {
				if rightLen <= leftLen && leftLen > 1 {
					leftLen--
					rightLen++
				}
			} else if line[1].x < line[0].x {
				if leftLen <= rightLen && rightLen > 1 {
					leftLen++
					rightLen--
				}
			}
		}

		if leftLen <= 0 || rightLen <= 0 {
			return
		}

		startX := minX + leftLen + 1
		if startX <= minX {
			startX = minX + 1
		}
		if end := startX + labelWidth; end >= maxX {
			startX = maxX - labelWidth - 1
		}
		startLabelCoord := drawingCoord{x: startX, y: middleY}
		d.drawText(startLabelCoord, label)
		return
	}

	startLabelCoord := drawingCoord{x: middleX - labelWidth/2, y: middleY}
	d.drawText(startLabelCoord, label)
}
