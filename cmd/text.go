package cmd

import "github.com/rivo/uniseg"

func textWidth(text string) int {
	if useAscii {
		return len(text)
	}
	return uniseg.StringWidth(text)
}

func (canvas *drawing) drawUnicodeText(start drawingCoord, text string, maxWidth int, textColor, styleType string) {
	graphemes := uniseg.NewGraphemes(text)
	column := start.x
	lastColumn := start.x
	zeroWidthText := ""
	for graphemes.Next() {
		width := graphemes.Width()
		if width == 0 {
			zeroWidthText += graphemes.Str()
			continue
		}
		if column+width > start.x+maxWidth {
			break
		}
		(*canvas)[column][start.y] = wrapTextInColor(zeroWidthText+graphemes.Str(), textColor, styleType)
		zeroWidthText = ""
		for offset := 1; offset < width; offset++ {
			(*canvas)[column+offset][start.y] = ""
		}
		lastColumn = column
		column += width
	}
	if zeroWidthText != "" {
		if column == start.x {
			(*canvas)[start.x][start.y] = zeroWidthText + (*canvas)[start.x][start.y]
		} else {
			(*canvas)[lastColumn][start.y] += zeroWidthText
		}
	}
}
