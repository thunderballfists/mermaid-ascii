package cmd

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

type graphProperties struct {
	data            *orderedmap.OrderedMap[string, []textEdge]
	styleClasses    *map[string]styleClass
	graphDirection  string
	styleType       string
	paddingX        int
	paddingY        int
	subgraphs       []*textSubgraph
	nodeDefinitions map[string]nodeDefinition
}

type textNode struct {
	id         string
	label      string
	shape      Shape
	styleClass string
}

type textEdge struct {
	parent textNode
	child  textNode
	label  string
}

type nodeDefinition struct {
	id            string
	label         string
	shape         Shape
	explicitLabel bool
	shapeProvided bool
}

type textSubgraph struct {
	name     string
	nodes    []string
	parent   *textSubgraph
	children []*textSubgraph
}

var nodeWithClassPattern = regexp.MustCompile(`^(.+):::(.+)$`)

type shapeDelimiter struct {
	open  string
	close string
	shape Shape
}

var shapeDelimiters = []shapeDelimiter{
	{open: "((((", close: "))))", shape: ShapeTripleCircle},
	{open: "(((", close: ")))", shape: ShapeDoubleCircle},
	{open: "((", close: "))", shape: ShapeCircle},
	{open: "[[", close: "]]", shape: ShapeSubroutine},
	{open: "{{", close: "}}", shape: ShapeHexagon},
	{open: "[/", close: "/]", shape: ShapeParallelogramRight},
	{open: "[\\", close: "\\]", shape: ShapeParallelogramLeft},
	{open: "[<", close: "]", shape: ShapeTrapezoidLeft},
	{open: "([", close: "])", shape: ShapeRoundedRectangle},
	{open: "(", close: ")", shape: ShapeRoundedRectangle},
	{open: ">", close: "]", shape: ShapeTrapezoidRight},
	{open: "{", close: "}", shape: ShapeDiamond},
	{open: "[", close: "]", shape: ShapeRectangle},
}

func (gp *graphProperties) parseNode(line string) textNode {
	trimmedLine := strings.TrimSpace(line)
	base := trimmedLine
	styleClass := ""
	if match := nodeWithClassPattern.FindStringSubmatch(trimmedLine); match != nil {
		base = strings.TrimSpace(match[1])
		styleClass = strings.TrimSpace(match[2])
	}

	def := gp.parseAndStoreNodeDefinition(base)
	return textNode{id: def.id, label: def.label, shape: def.shape, styleClass: styleClass}
}

func (gp *graphProperties) parseAndStoreNodeDefinition(token string) nodeDefinition {
	token = strings.TrimSpace(token)
	if token == "" {
		return nodeDefinition{}
	}

	id, remainder := splitNodeIdentifier(token)
	if id == "" {
		id = token
	}

	existing := gp.getOrCreateNodeDefinition(id)

	updated := existing
	updated.id = id

	if strings.TrimSpace(remainder) != "" {
		if shape, rawLabel, _, ok := parseShapeAndLabel(remainder); ok {
			updated.shape = shape
			updated.shapeProvided = true
			if label, explicit := processRawLabel(rawLabel); explicit {
				updated.label = label
				updated.explicitLabel = true
			} else if !updated.explicitLabel {
				updated.label = id
				updated.explicitLabel = false
			}
		}
	}

	if updated.label == "" {
		updated.label = id
		updated.explicitLabel = false
	}

	gp.nodeDefinitions[id] = updated
	return updated
}

func (gp *graphProperties) getOrCreateNodeDefinition(id string) nodeDefinition {
	if def, ok := gp.nodeDefinitions[id]; ok {
		return def
	}
	def := nodeDefinition{
		id:            id,
		label:         id,
		shape:         defaultShape(),
		explicitLabel: false,
		shapeProvided: false,
	}
	gp.nodeDefinitions[id] = def
	return def
}

func splitNodeIdentifier(token string) (string, string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ""
	}

	for i, r := range token {
		if isDelimiterStart(r) {
			return strings.TrimSpace(token[:i]), strings.TrimSpace(token[i:])
		}
	}

	return strings.TrimSpace(token), ""
}

func isDelimiterStart(r rune) bool {
	switch r {
	case '[', '(', '{', '>':
		return true
	default:
		return false
	}
}

func parseShapeAndLabel(remainder string) (Shape, string, string, bool) {
	trimmed := strings.TrimSpace(remainder)
	for _, spec := range shapeDelimiters {
		if strings.HasPrefix(trimmed, spec.open) {
			content := trimmed[len(spec.open):]
			raw, rest, ok := extractLabel(content, spec.close)
			if !ok {
				return spec.shape, "", "", false
			}
			return spec.shape, raw, rest, true
		}
	}
	return defaultShape(), "", "", false
}

func extractLabel(content, close string) (string, string, bool) {
	var builder strings.Builder
	i := 0
	for i < len(content) {
		if strings.HasPrefix(content[i:], close) {
			return builder.String(), content[i+len(close):], true
		}
		if content[i] == '\\' {
			if i+1 < len(content) {
				builder.WriteByte(content[i])
				builder.WriteByte(content[i+1])
				i += 2
				continue
			}
			// dangling backslash, treat as literal
			builder.WriteByte(content[i])
			i++
			continue
		}
		builder.WriteByte(content[i])
		i++
	}
	return "", "", false
}

func processRawLabel(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	stripped := trimmed
	if withoutQuotes, ok := stripOuterQuotes(trimmed); ok {
		stripped = withoutQuotes
	}
	sanitized := sanitizeLabel(stripped)
	if sanitized == "" {
		return "", false
	}
	return sanitized, true
}

func stripOuterQuotes(s string) (string, bool) {
	if len(s) < 2 {
		return s, false
	}
	first := s[0]
	last := s[len(s)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		if hasOddTrailingBackslashes(s[:len(s)-1]) {
			return s, false
		}
		return s[1 : len(s)-1], true
	}
	return s, false
}

func hasOddTrailingBackslashes(s string) bool {
	count := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\\'; i-- {
		count++
	}
	return count%2 == 1
}

func parseStyleClass(matchedLine []string) styleClass {
	className := matchedLine[0]
	styles := matchedLine[1]
	// Styles are comma separated and key-values are separated by colon
	// Example: fill:#f9f,stroke:#333,stroke-width:4px
	styleMap := make(map[string]string)
	for _, style := range strings.Split(styles, ",") {
		kv := strings.Split(style, ":")
		styleMap[kv[0]] = kv[1]
	}
	return styleClass{className, styleMap}
}

func setArrowWithLabel(lhs, rhs []textNode, label string, data *orderedmap.OrderedMap[string, []textEdge]) []textNode {
	log.Debug("Setting arrow from ", lhs, " to ", rhs, " with label ", label)
	for _, l := range lhs {
		for _, r := range rhs {
			setData(l, textEdge{l, r, label}, data)
		}
	}
	return rhs
}

func setArrow(lhs, rhs []textNode, data *orderedmap.OrderedMap[string, []textEdge]) []textNode {
	return setArrowWithLabel(lhs, rhs, "", data)
}

func addNode(node textNode, data *orderedmap.OrderedMap[string, []textEdge]) {
	if _, ok := data.Get(node.id); !ok {
		data.Set(node.id, []textEdge{})
	}
}

func setData(parent textNode, edge textEdge, data *orderedmap.OrderedMap[string, []textEdge]) {
	// Check if the parent is in the map
	if children, ok := data.Get(parent.id); ok {
		// If it is, append the child to the list of children
		data.Set(parent.id, append(children, edge))
	} else {
		// If it isn't, add it to the map
		data.Set(parent.id, []textEdge{edge})
	}
	// Check if the child is in the map
	if _, ok := data.Get(edge.child.id); ok {
		// If it is, do nothing
	} else {
		// If it isn't, add it to the map
		data.Set(edge.child.id, []textEdge{})
	}
}

func (gp *graphProperties) parseString(line string) ([]textNode, error) {
	log.Debugf("Parsing line: %v", line)
	var lhs, rhs []textNode
	var err error
	// Patterns are matched in order
	patterns := []struct {
		regex   *regexp.Regexp
		handler func([]string) ([]textNode, error)
	}{
		{
			regex: regexp.MustCompile(`^\s*$`),
			handler: func(match []string) ([]textNode, error) {
				// Ignore empty lines
				return []textNode{}, nil
			},
		},
		{
			regex: regexp.MustCompile(`^(.+)\s+-->\s+(.+)$`),
			handler: func(match []string) ([]textNode, error) {
				if lhs, err = gp.parseString(match[0]); err != nil {
					lhs = []textNode{gp.parseNode(match[0])}
				}
				if rhs, err = gp.parseString(match[1]); err != nil {
					rhs = []textNode{gp.parseNode(match[1])}
				}
				return setArrow(lhs, rhs, gp.data), nil
			},
		},
		{
			regex: regexp.MustCompile(`^(.+)\s+-->\|(.+)\|\s+(.+)$`),
			handler: func(match []string) ([]textNode, error) {
				if lhs, err = gp.parseString(match[0]); err != nil {
					lhs = []textNode{gp.parseNode(match[0])}
				}
				if rhs, err = gp.parseString(match[2]); err != nil {
					rhs = []textNode{gp.parseNode(match[2])}
				}
				return setArrowWithLabel(lhs, rhs, match[1], gp.data), nil
			},
		},
		{
			regex: regexp.MustCompile(`^classDef\s+(.+)\s+(.+)$`),
			handler: func(match []string) ([]textNode, error) {
				s := parseStyleClass(match)
				(*gp.styleClasses)[s.name] = s
				return []textNode{}, nil
			},
		},
		{
			regex: regexp.MustCompile(`^(.+) & (.+)$`),
			handler: func(match []string) ([]textNode, error) {
				log.Debugf("Found & pattern node %v to %v", match[0], match[1])
				var node textNode
				if lhs, err = gp.parseString(match[0]); err != nil {
					node = gp.parseNode(match[0])
					lhs = []textNode{node}
				}
				if rhs, err = gp.parseString(match[1]); err != nil {
					node = gp.parseNode(match[1])
					rhs = []textNode{node}
				}
				return append(lhs, rhs...), nil
			},
		},
	}
	for _, pattern := range patterns {
		if match := pattern.regex.FindStringSubmatch(line); match != nil {
			nodes, err := pattern.handler(match[1:])
			if err == nil {
				return nodes, nil
			}
		}
	}
	return []textNode{}, errors.New("Could not parse line: " + line)
}

func mermaidFileToMap(mermaid, styleType string) (*graphProperties, error) {
	// Allow split on both \n and the actual string "\n" for curl compatibility
	newlinePattern := regexp.MustCompile(`\n|\\n`)
	rawLines := newlinePattern.Split(string(mermaid), -1)

	// Process lines to remove comments
	lines := []string{}
	for _, line := range rawLines {
		// Stop processing at "---" separator (used in test files)
		if line == "---" {
			break
		}

		// Skip lines that start with %% (comment lines)
		if strings.HasPrefix(strings.TrimSpace(line), "%%") {
			continue
		}

		// Remove inline comments (anything after %%) and trim resulting whitespace
		if idx := strings.Index(line, "%%"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Skip empty lines after comment removal
		if len(strings.TrimSpace(line)) > 0 {
			lines = append(lines, line)
		}
	}

	data := orderedmap.NewOrderedMap[string, []textEdge]()
	styleClasses := make(map[string]styleClass)
	properties := graphProperties{
		data:            data,
		styleClasses:    &styleClasses,
		graphDirection:  "",
		styleType:       styleType,
		paddingX:        paddingBetweenX,
		paddingY:        paddingBetweenY,
		subgraphs:       []*textSubgraph{},
		nodeDefinitions: make(map[string]nodeDefinition),
	}

	// Pick up optional padding directives before the graph definition
	paddingRegex := regexp.MustCompile(`^(?i)padding([xy])\s*=\s*(\d+)$`)
	for len(lines) > 0 {
		trimmed := strings.TrimSpace(lines[0])
		if trimmed == "" {
			lines = lines[1:]
			continue
		}
		if match := paddingRegex.FindStringSubmatch(trimmed); match != nil {
			paddingValue, err := strconv.Atoi(match[2])
			if err != nil {
				return &properties, err
			}
			if strings.EqualFold(match[1], "x") {
				properties.paddingX = paddingValue
			} else {
				properties.paddingY = paddingValue
			}
			lines = lines[1:]
			continue
		}
		break
	}

	if len(lines) == 0 {
		return &properties, errors.New("missing graph definition")
	}

	// First line should either say "graph TD" or "graph LR"
	switch lines[0] {
	case "graph LR", "flowchart LR":
		graphDirection = "LR"
	case "graph TD", "flowchart TD":
		graphDirection = "TD"
	default:
		return &properties, errors.New("first line should define the graph")
	}
	lines = lines[1:]

	// Track subgraph context using a stack
	subgraphStack := []*textSubgraph{}
	subgraphRegex := regexp.MustCompile(`^\s*subgraph\s+(.+)$`)
	endRegex := regexp.MustCompile(`^\s*end\s*$`)

	// Iterate over the lines
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for subgraph start
		if match := subgraphRegex.FindStringSubmatch(trimmedLine); match != nil {
			subgraphName := strings.TrimSpace(match[1])
			newSubgraph := &textSubgraph{
				name:     subgraphName,
				nodes:    []string{},
				children: []*textSubgraph{},
			}

			// Set parent relationship if we're nested
			if len(subgraphStack) > 0 {
				parent := subgraphStack[len(subgraphStack)-1]
				newSubgraph.parent = parent
				parent.children = append(parent.children, newSubgraph)
			}

			subgraphStack = append(subgraphStack, newSubgraph)
			properties.subgraphs = append(properties.subgraphs, newSubgraph)
			log.Debugf("Started subgraph %s", subgraphName)
			continue
		}

		// Check for subgraph end
		if endRegex.MatchString(trimmedLine) {
			if len(subgraphStack) > 0 {
				closedSubgraph := subgraphStack[len(subgraphStack)-1]
				subgraphStack = subgraphStack[:len(subgraphStack)-1]
				log.Debugf("Ended subgraph %s", closedSubgraph.name)
			}
			continue
		}

		// Remember nodes before parsing this line
		existingNodes := make(map[string]bool)
		for el := data.Front(); el != nil; el = el.Next() {
			existingNodes[el.Key] = true
		}

		// Parse nodes and edges normally
		nodes, err := properties.parseString(line)
		if err != nil {
			log.Debugf("Parsing remaining text to node %v", line)
			node := properties.parseNode(line)
			addNode(node, properties.data)
		} else {
			// Ensure all returned nodes are in the map
			for _, node := range nodes {
				addNode(node, properties.data)
			}
		}

		// Add all new nodes to current subgraph(s)
		if len(subgraphStack) > 0 {
			for el := data.Front(); el != nil; el = el.Next() {
				nodeName := el.Key
				// If this is a new node (wasn't in existingNodes), add it to subgraph
				if !existingNodes[nodeName] {
					for _, sg := range subgraphStack {
						// Check if node is not already in the subgraph
						found := false
						for _, n := range sg.nodes {
							if n == nodeName {
								found = true
								break
							}
						}
						if !found {
							sg.nodes = append(sg.nodes, nodeName)
							log.Debugf("Added node %s to subgraph %s", nodeName, sg.name)
						}
					}
				}
			}
		}
	}
	return &properties, nil
}
