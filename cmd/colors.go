package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

type linkStyleRule struct {
	indices []int
	colors  map[string]string
}

func normalizeColor(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 4 && value[0] == '#' {
		value = fmt.Sprintf("#%c%c%c%c%c%c", value[1], value[1], value[2], value[2], value[3], value[3])
	}
	if len(value) != 7 || value[0] != '#' {
		return ""
	}
	if _, err := strconv.ParseUint(value[1:], 16, 24); err != nil {
		return ""
	}
	return strings.ToLower(value)
}

func parseLinkStyle(line string) (linkStyleRule, error) {
	rule := linkStyleRule{colors: make(map[string]string)}
	fields := strings.Fields(strings.TrimSuffix(line, ";"))
	if len(fields) < 3 {
		return rule, fmt.Errorf("linkStyle requires edge indices and styles")
	}
	if fields[1] != "default" {
		for _, selector := range strings.Split(fields[1], ",") {
			index, err := strconv.Atoi(selector)
			if err != nil || index < 0 {
				return rule, fmt.Errorf("invalid linkStyle edge index %q", selector)
			}
			rule.indices = append(rule.indices, index)
		}
	}
	var err error
	rule.colors, err = parseColorDeclarations(strings.Join(fields[2:], ""))
	return rule, err
}

func parseColorDeclarations(declarations string) (map[string]string, error) {
	colors := make(map[string]string)
	for _, declaration := range strings.Split(strings.TrimSuffix(declarations, ";"), ",") {
		name, value, found := strings.Cut(declaration, ":")
		if !found {
			return nil, fmt.Errorf("invalid color declaration %q", declaration)
		}
		if name == "stroke" || name == "color" {
			color := normalizeColor(value)
			if color == "" {
				return nil, fmt.Errorf("%s requires a hex color", name)
			}
			colors[name] = color
		}
	}
	return colors, nil
}

func (g *graph) colorizeDrawing(lines, corners, heads, labels []*drawing) {
	styled := false
	for _, connection := range g.edges {
		styled = styled || len(connection.colors) != 0
	}
	for _, contentNode := range g.nodes {
		styled = styled || normalizeColor(contentNode.styleClass.styles["stroke"]) != ""
	}
	if !styled {
		return
	}
	colors := make(map[drawingCoord]string)
	for _, layers := range [][]*drawing{lines, corners, heads} {
		for index, layer := range layers {
			if layer == nil {
				continue
			}
			for x, column := range *layer {
				for y, cell := range column {
					if cell != " " && cell != "" {
						colors[drawingCoord{x, y}] = g.edges[index].colors["stroke"]
					}
				}
			}
		}
	}
	for index, connection := range g.edges {
		if labels[index] == nil || textWidth(connection.text) == 0 {
			continue
		}
		position := connection.labelPosition
		if position == nil {
			start, available := labelPositionOnLine(g.lineToDrawing(connection.labelLine), connection.text, connection.nearHead)
			if !available {
				continue
			}
			position = &start
		}
		labelColor := connection.colors["color"]
		if labelColor == "" {
			labelColor = connection.colors["stroke"]
		}
		for offset := 0; offset < textWidth(connection.text); offset++ {
			colors[drawingCoord{position.x + offset, position.y}] = labelColor
		}
	}
	for _, contentNode := range g.nodes {
		width, height := getDrawingSize(contentNode.drawing)
		borderColor := normalizeColor(contentNode.styleClass.styles["stroke"])
		for x := 0; x <= width; x++ {
			for y := 0; y <= height; y++ {
				position := drawingCoord{contentNode.drawingCoord.x + x, contentNode.drawingCoord.y + y}
				delete(colors, position)
				if contentNode.isBorderCell(x, y, width, height) {
					colors[position] = borderColor
				}
			}
		}
	}
	for _, group := range g.subgraphs {
		if len(group.nodes) == 0 {
			continue
		}
		for x := group.minX; x <= group.maxX; x++ {
			delete(colors, drawingCoord{x, group.minY})
			delete(colors, drawingCoord{x, group.maxY})
		}
		for y := group.minY; y <= group.maxY; y++ {
			delete(colors, drawingCoord{group.minX, y})
			delete(colors, drawingCoord{group.maxX, y})
		}
		titleX := Max(group.minX+1, group.minX+(group.maxX-group.minX)/2-textWidth(group.name)/2)
		for offset := 0; offset < textWidth(group.name); offset++ {
			delete(colors, drawingCoord{titleX + offset, group.minY + 1})
		}
	}
	for position, color := range colors {
		if color != "" && (*g.drawing)[position.x][position.y] != "" {
			(*g.drawing)[position.x][position.y] = wrapTextInColor((*g.drawing)[position.x][position.y], color, g.styleType)
		}
	}
}
