package cmd

func (n *node) drawShapeBorders(d *drawing, width, height int) {
	switch n.shape {
	case ShapeDiamond:
		for x := 0; x <= width; x++ {
			for y := 0; y <= height; y++ {
				(*d)[x][y] = " "
				if glyph := diamondBorderGlyph(x, y, width, height); glyph != "" {
					(*d)[x][y] = glyph
				}
			}
		}
	case ShapeCylinder:
		(*d)[0][1], (*d)[width][1] = "├", "┤"
		for column := 1; column < width; column++ {
			(*d)[column][1] = "─"
		}
	case ShapeStadium:
		for row := 1; row < height; row++ {
			(*d)[0][row], (*d)[width][row] = "(", ")"
		}
	case ShapeSubroutine:
		for _, column := range []int{2, width - 2} {
			(*d)[column][0], (*d)[column][height] = "┬", "┴"
			for row := 1; row < height; row++ {
				(*d)[column][row] = "│"
			}
		}
	}
}

func (n *node) isBorderCell(x, y, width, height int) bool {
	if n.shape == ShapeDiamond {
		return diamondBorderGlyph(x, y, width, height) != ""
	}
	return x == 0 || x == width || y == 0 || y == height ||
		(n.shape == ShapeCylinder && y == 1) ||
		(n.shape == ShapeSubroutine && (x == 2 || x == width-2))
}

// Routing and labels may grow shared bands after the initial node sizing.
// Odd middle bands align the diamond's four tips with the cardinal grid ports.
func (g *graph) normalizeDiamondBands() {
	if useAscii {
		return
	}
	for _, n := range g.nodes {
		if n.shape != ShapeDiamond {
			continue
		}
		// Label measurement can expand a trailing border band. Keep ports at
		// the visible tips by transferring that excess into the following gap.
		for _, axis := range []struct {
			sizes map[int]int
			start int
		}{{g.columnWidth, n.gridCoord.x}, {g.rowHeight, n.gridCoord.y}} {
			border := axis.start + 2
			if extra := axis.sizes[border] - 1; extra > 0 {
				axis.sizes[border] = 1
				axis.sizes[border+1] += extra
			}
		}
		if y := n.gridCoord.y + 1; g.rowHeight[y]%2 == 0 {
			g.rowHeight[y]++
		}
		// A tall shared row must still leave a border cell on every slope row.
		x := n.gridCoord.x + 1
		g.columnWidth[x] = Max(g.columnWidth[x], g.rowHeight[n.gridCoord.y+1])
		if g.columnWidth[x]%2 == 0 {
			g.columnWidth[x]++
		}
	}
}

// Each sloping side uses diagonal steps joined by horizontal cells. This
// preserves a closed outline without making long labels proportionally taller.
func diamondBorderGlyph(x, y, width, height int) string {
	midX, midY := width/2, height/2
	bottom := y > midY
	if bottom {
		y = height - y
	}
	if y == 0 {
		if x != midX {
			return ""
		}
		if bottom {
			return "∨"
		}
		return "∧"
	}
	if y == midY {
		if x == 0 {
			return "<"
		}
		if x == width {
			return ">"
		}
	}
	outer := midX * (midY - y) / midY
	inner := midX*(midY-y+1)/midY - 1
	right := x > midX
	if right {
		x = width - x
	}
	if x < outer || x > inner {
		return ""
	}
	if x != outer {
		return "─"
	}
	if right != bottom {
		return "╲"
	}
	return "╱"
}
