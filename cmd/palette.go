package cmd

// Thickness represents the weight of lines in extended mode.
type Thickness int

const (
	ThicknessLight Thickness = iota
	ThicknessHeavy
	ThicknessDouble
)

// LineStyle represents the visual style of stems in extended mode.
type LineStyle int

const (
	LineStyleSolid LineStyle = iota
	LineStyleDotted
	LineStyleDashed
)

// HeadStyle controls which glyph is used for arrow heads.
type HeadStyle int

const (
	HeadStylePointer HeadStyle = iota
	HeadStyleOpen
	HeadStyleWide
	HeadStyleNone
)

// RoundedFallback determines how non-light thicknesses handle curved corners.
type RoundedFallback int

const (
	RoundedFallbackSquare RoundedFallback = iota
	RoundedFallbackRoundedLight
)

// FlowOptions describes how a flow element (edge/box) should be rendered.
type FlowOptions struct {
	Thickness       Thickness
	LineStyle       LineStyle
	CurvedCorners   bool
	RoundedFallback RoundedFallback
	Heads           HeadStyle
}

type palette struct {
	Horizontal string
	Vertical   string

	CornerTopLeft     string
	CornerTopRight    string
	CornerBottomRight string
	CornerBottomLeft  string

	TeeLeft  string
	TeeRight string
	TeeUp    string
	TeeDown  string
	TeeCross string

	HeadRight string
	HeadLeft  string
	HeadUp    string
	HeadDown  string
}

func defaultFlowOptions() FlowOptions {
	return FlowOptions{
		Thickness:       ThicknessLight,
		LineStyle:       LineStyleSolid,
		CurvedCorners:   false,
		RoundedFallback: RoundedFallbackSquare,
		Heads:           HeadStylePointer,
	}
}

func buildPalette(opts FlowOptions) palette {
	pal := palette{}

	// Determine stems first as they impact junction appearance.
	switch opts.LineStyle {
	case LineStyleDotted:
		pal.Horizontal = "┈"
		pal.Vertical = "┊"
	case LineStyleDashed:
		pal.Horizontal = "┄"
		pal.Vertical = "┆"
	default:
		pal.Horizontal = "─"
		pal.Vertical = "│"
		switch opts.Thickness {
		case ThicknessHeavy:
			pal.Horizontal = "━"
			pal.Vertical = "┃"
		case ThicknessDouble:
			pal.Horizontal = "═"
			pal.Vertical = "║"
		}
	}

	squareCorners := map[Thickness][4]string{
		ThicknessLight:  {"┌", "┐", "┘", "└"},
		ThicknessHeavy:  {"┏", "┓", "┛", "┗"},
		ThicknessDouble: {"╔", "╗", "╝", "╚"},
	}
	roundedCorners := [4]string{"╭", "╮", "╯", "╰"}

	useCorners := squareCorners[ThicknessLight]
	if v, ok := squareCorners[opts.Thickness]; ok {
		useCorners = v
	}

	if opts.CurvedCorners {
		switch opts.Thickness {
		case ThicknessLight:
			useCorners = roundedCorners
		default:
			if opts.RoundedFallback == RoundedFallbackRoundedLight {
				useCorners = roundedCorners
			}
		}
	}

	pal.CornerTopLeft = useCorners[0]
	pal.CornerTopRight = useCorners[1]
	pal.CornerBottomRight = useCorners[2]
	pal.CornerBottomLeft = useCorners[3]

	teeChars := map[Thickness][5]string{
		ThicknessLight:  {"├", "┤", "┬", "┴", "┼"},
		ThicknessHeavy:  {"┣", "┫", "┳", "┻", "╋"},
		ThicknessDouble: {"╠", "╣", "╦", "╩", "╬"},
	}
	useTees := teeChars[ThicknessLight]
	if v, ok := teeChars[opts.Thickness]; ok {
		useTees = v
	}
	pal.TeeLeft = useTees[0]
	pal.TeeRight = useTees[1]
	pal.TeeUp = useTees[2]
	pal.TeeDown = useTees[3]
	pal.TeeCross = useTees[4]

	switch opts.Heads {
	case HeadStyleOpen:
		pal.HeadRight, pal.HeadLeft, pal.HeadUp, pal.HeadDown = "→", "←", "↑", "↓"
	case HeadStyleWide:
		pal.HeadRight, pal.HeadLeft, pal.HeadUp, pal.HeadDown = "▶", "◀", "▲", "▼"
	case HeadStyleNone:
		pal.HeadRight, pal.HeadLeft, pal.HeadUp, pal.HeadDown = "", "", "", ""
	default:
		pal.HeadRight, pal.HeadLeft, pal.HeadUp, pal.HeadDown = "►", "◄", "▲", "▼"
	}

	return pal
}

func (p palette) arrowHead(dir direction, fallback direction) string {
	switch dir {
	case Right:
		return p.HeadRight
	case Left:
		return p.HeadLeft
	case Up:
		return p.HeadUp
	case Down:
		return p.HeadDown
	}

	switch fallback {
	case Right:
		return p.HeadRight
	case Left:
		return p.HeadLeft
	case Up:
		return p.HeadUp
	case Down:
		return p.HeadDown
	}
	return ""
}
