package cmd

func (n *node) drawShapeBorders(d *drawing, width, height int) {
	switch n.shape {
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
	return x == 0 || x == width || y == 0 || y == height ||
		(n.shape == ShapeCylinder && y == 1) ||
		(n.shape == ShapeSubroutine && (x == 2 || x == width-2))
}
