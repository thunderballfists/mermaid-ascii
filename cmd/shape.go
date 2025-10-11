package cmd

type Shape int

const (
	ShapeRectangle Shape = iota
	ShapeRoundedRectangle
	ShapeCircle
	ShapeDoubleCircle
	ShapeTripleCircle
	ShapeDiamond
	ShapeHexagon
	ShapeParallelogramRight
	ShapeParallelogramLeft
	ShapeTrapezoidRight
	ShapeTrapezoidLeft
	ShapeSubroutine
)

func defaultShape() Shape {
	return ShapeRectangle
}
