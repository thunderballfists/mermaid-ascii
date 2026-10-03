package cmd

const (
	junctionUp uint8 = 1 << iota
	junctionRight
	junctionDown
	junctionLeft
)

type junctionStroke struct {
	directions       uint8
	horizontalWeight Thickness
	verticalWeight   Thickness
}

var junctionProfiles = [][2]Thickness{
	{ThicknessLight, ThicknessLight},
	{ThicknessHeavy, ThicknessHeavy},
	{ThicknessDouble, ThicknessDouble},
	{ThicknessDouble, ThicknessLight},
	{ThicknessLight, ThicknessDouble},
}

var junctionGlyphs = map[uint8][5]string{
	junctionUp:                                               {"╵", "╹"},
	junctionRight:                                            {"╶", "╺"},
	junctionDown:                                             {"╷", "╻"},
	junctionLeft:                                             {"╴", "╸"},
	junctionUp | junctionDown:                                {"│", "┃", "║"},
	junctionLeft | junctionRight:                             {"─", "━", "═"},
	junctionDown | junctionRight:                             {"┌", "┏", "╔", "╒", "╓"},
	junctionDown | junctionLeft:                              {"┐", "┓", "╗", "╕", "╖"},
	junctionUp | junctionRight:                               {"└", "┗", "╚", "╘", "╙"},
	junctionUp | junctionLeft:                                {"┘", "┛", "╝", "╛", "╜"},
	junctionUp | junctionDown | junctionRight:                {"├", "┣", "╠", "╞", "╟"},
	junctionUp | junctionDown | junctionLeft:                 {"┤", "┫", "╣", "╡", "╢"},
	junctionLeft | junctionRight | junctionDown:              {"┬", "┳", "╦", "╤", "╥"},
	junctionLeft | junctionRight | junctionUp:                {"┴", "┻", "╩", "╧", "╨"},
	junctionUp | junctionRight | junctionDown | junctionLeft: {"┼", "╋", "╬", "╪", "╫"},
}

var junctionStrokes = buildJunctionStrokes()

func buildJunctionStrokes() map[string]junctionStroke {
	strokes := make(map[string]junctionStroke)
	for directions, glyphs := range junctionGlyphs {
		for profileIndex, glyph := range glyphs {
			if glyph == "" {
				continue
			}
			stroke := junctionStroke{directions: directions}
			if directions&(junctionLeft|junctionRight) != 0 {
				stroke.horizontalWeight = junctionProfiles[profileIndex][0]
			}
			if directions&(junctionUp|junctionDown) != 0 {
				stroke.verticalWeight = junctionProfiles[profileIndex][1]
			}
			strokes[glyph] = stroke
		}
	}
	for alias, canonical := range map[string]string{
		"┈": "─", "┄": "─", "┊": "│", "┆": "│",
		"╭": "┌", "╮": "┐", "╯": "┘", "╰": "└",
	} {
		strokes[alias] = strokes[canonical]
	}
	return strokes
}

func isJunctionChar(glyph string) bool {
	_, known := junctionStrokes[glyph]
	return known
}

func mergeJunctions(base, overlay string) string {
	baseStroke, baseKnown := junctionStrokes[base]
	overlayStroke, overlayKnown := junctionStrokes[overlay]
	if base == overlay || !baseKnown || !overlayKnown {
		return base
	}
	directions := baseStroke.directions | overlayStroke.directions
	horizontalWeight := Thickness(Max(int(baseStroke.horizontalWeight), int(overlayStroke.horizontalWeight)))
	verticalWeight := Thickness(Max(int(baseStroke.verticalWeight), int(overlayStroke.verticalWeight)))
	if directions&(junctionLeft|junctionRight) == 0 {
		horizontalWeight = verticalWeight
	}
	if directions&(junctionUp|junctionDown) == 0 {
		verticalWeight = horizontalWeight
	}
	glyphs := junctionGlyphs[directions]
	for profileIndex, weights := range junctionProfiles {
		if weights[0] == horizontalWeight && weights[1] == verticalWeight && glyphs[profileIndex] != "" {
			return glyphs[profileIndex]
		}
	}
	return glyphs[Max(int(horizontalWeight), int(verticalWeight))]
}
