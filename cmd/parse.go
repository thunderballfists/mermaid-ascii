package cmd

import (
	"errors"
	"fmt"
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
	boxPaddingX     *int
	boxPaddingY     *int
	uniformSpacing  bool
	subgraphs       []*textSubgraph
	nodeDefinitions map[string]nodeDefinition
	linkStyles      []linkStyleRule
}

type textNode struct {
	id         string
	label      string
	shape      Shape
	styleClass string
}

type textEdge struct {
	sourceIndex   int
	parent        textNode
	child         textNode
	label         string
	labelNearHead bool
	edgeStyle     EdgeStyle
	arrowDir      arrowDirection
	leftCap       arrowEndCap
	rightCap      arrowEndCap
	lenHint       int
}

type nodeDefinition struct {
	styleClass    string
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
var classAssignmentPattern = regexp.MustCompile(`^class\s+([^\s,;]+(?:\s*,\s*[^\s,;]+)*)\s+([^\s,;]+)\s*;?\s*$`)

var edgeOperatorPattern = `<?[ox)]?[-.=]+(?:[ox)]?[-.=]+)*[ox)]?>?`

var edgeTailPipeLabelRegex = regexp.MustCompile(`^(.+?)\s*(` + edgeOperatorPattern + `)\s*\|(.+?)\|\s*(.+)$`)
var edgeMidPipeLabelRegex = regexp.MustCompile(`^(.+?)\s*--\s*\|(.+?)\|\s*(` + edgeOperatorPattern + `)\s*(.+)$`)
var edgeMidTextLabelRegex = regexp.MustCompile(`^(.+?)\s*--\s+(.+?)\s+(` + edgeOperatorPattern + `)\s*(.+)$`)
var edgeOpenLabelRegex = regexp.MustCompile(`^(.+?)\s*--\s+(.+?)\s+--\s*(.+)$`)
var edgeOperatorRegex = regexp.MustCompile(edgeOperatorPattern)

type edgeSpec struct {
	lhs       string
	rhs       string
	operator  string
	label     string
	nearHead  bool
	edgeStyle EdgeStyle
	direction arrowDirection
	leftCap   arrowEndCap
	rightCap  arrowEndCap
	lenHint   int
}

func classifyEdgeOperator(raw string) (EdgeStyle, arrowDirection, arrowEndCap, arrowEndCap, int) {
	operator := strings.TrimSpace(raw)
	if operator == "" {
		return StyleSolid, arrowDirOpen, arrowCapNone, arrowCapNone, 0
	}

	style := StyleSolid
	switch {
	case strings.Contains(operator, "="):
		style = StyleThick
	case strings.Contains(operator, "."):
		style = StyleDotted
	default:
		style = StyleSolid
	}

	hasLeft := strings.Contains(operator, "<")
	hasRight := strings.Contains(operator, ">")
	direction := arrowDirOpen
	switch {
	case hasLeft && hasRight:
		direction = arrowDirBoth
	case hasRight:
		direction = arrowDirForward
	case hasLeft:
		direction = arrowDirBackward
	default:
		direction = arrowDirOpen
	}

	lenHint := 0
	runes := []rune(operator)
	for _, r := range runes {
		if r == '-' || r == '.' || r == '=' {
			lenHint++
		}
	}

	firstBody := -1
	lastBody := -1
	for i, r := range runes {
		if r == '-' || r == '.' || r == '=' {
			firstBody = i
			break
		}
	}
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == '-' || runes[i] == '.' || runes[i] == '=' {
			lastBody = i
			break
		}
	}

	leftCap := arrowCapNone
	rightCap := arrowCapNone

	if firstBody > 0 {
		for i := 0; i < firstBody; i++ {
			switch runes[i] {
			case 'o':
				leftCap = arrowCapCircle
			case 'x':
				leftCap = arrowCapCross
			case ')':
				leftCap = arrowCapDiamond
			}
		}
	}

	if lastBody != -1 && lastBody < len(runes)-1 {
		for i := len(runes) - 1; i > lastBody; i-- {
			switch runes[i] {
			case 'o':
				rightCap = arrowCapCircle
			case 'x':
				rightCap = arrowCapCross
			case ')':
				rightCap = arrowCapDiamond
			}
		}
	}

	if rightCap == arrowCapNone {
		for i := len(runes) - 1; i >= 0; i-- {
			if runes[i] != 'o' && runes[i] != 'x' && runes[i] != ')' {
				continue
			}
			if firstBody != -1 && i > firstBody {
				switch runes[i] {
				case 'o':
					rightCap = arrowCapCircle
				case 'x':
					rightCap = arrowCapCross
				case ')':
					rightCap = arrowCapDiamond
				}
				break
			}
			if firstBody != -1 && i < firstBody {
				switch runes[i] {
				case 'o':
					leftCap = arrowCapCircle
				case 'x':
					leftCap = arrowCapCross
				case ')':
					leftCap = arrowCapDiamond
				}
				break
			}
		}
	}

	if lenHint == 0 {
		lenHint = len(runes)
	}

	return style, direction, leftCap, rightCap, lenHint
}

func buildEdgeSpec(lhs, rhs, operator, label string, nearHead bool) edgeSpec {
	lhs = strings.TrimSpace(lhs)
	rhs = strings.TrimSpace(rhs)
	operator = strings.TrimSpace(operator)
	if !useAscii && strings.HasPrefix(operator, ")") {
		_, remainder := splitNodeIdentifier(lhs)
		if _, _, rest, ok := parseShapeAndLabel(remainder + ")"); ok && rest == "" {
			lhs += ")"
			operator = operator[1:]
		}
	}
	style, direction, leftCap, rightCap, lenHint := classifyEdgeOperator(operator)
	return edgeSpec{
		lhs:       lhs,
		rhs:       rhs,
		operator:  operator,
		label:     label,
		nearHead:  nearHead,
		edgeStyle: style,
		direction: direction,
		leftCap:   leftCap,
		rightCap:  rightCap,
		lenHint:   lenHint,
	}
}

func splitEdgeComponents(line string) (string, string, string, bool) {
	matches := edgeOperatorRegex.FindAllStringIndex(line, -1)
	var opRange []int
	var operator string
	for i := len(matches) - 1; i >= 0; i-- {
		rng := matches[i]
		op := line[rng[0]:rng[1]]
		count := 0
		for _, r := range op {
			if r == '-' || r == '.' || r == '=' {
				count++
			}
		}
		if count >= 2 {
			opRange = rng
			operator = op
			break
		}
	}
	if opRange == nil {
		return "", "", "", false
	}
	lhs := strings.TrimSpace(line[:opRange[0]])
	rhs := strings.TrimSpace(line[opRange[1]:])
	if lhs == "" || rhs == "" {
		return "", "", "", false
	}
	return lhs, operator, rhs, true
}

func sanitizeEdgeLabel(raw string) string {
	label, ok := processRawLabel(raw)
	if !ok {
		return ""
	}
	return label
}

func parseEdgeSpecification(line string) (edgeSpec, bool) {
	if match := edgeMidPipeLabelRegex.FindStringSubmatch(line); match != nil {
		label := sanitizeEdgeLabel(match[2])
		return buildEdgeSpec(match[1], match[4], match[3], label, false), true
	}
	if match := edgeMidTextLabelRegex.FindStringSubmatch(line); match != nil {
		label := sanitizeEdgeLabel(match[2])
		return buildEdgeSpec(match[1], match[4], match[3], label, false), true
	}
	if match := edgeTailPipeLabelRegex.FindStringSubmatch(line); match != nil {
		label := sanitizeEdgeLabel(match[3])
		return buildEdgeSpec(match[1], match[4], match[2], label, label != ""), true
	}
	if match := edgeOpenLabelRegex.FindStringSubmatch(line); match != nil {
		label := sanitizeEdgeLabel(match[2])
		return buildEdgeSpec(match[1], match[3], "---", label, false), true
	}
	if lhs, operator, rhs, ok := splitEdgeComponents(line); ok {
		return buildEdgeSpec(lhs, rhs, operator, "", false), true
	}
	return edgeSpec{}, false
}

func (gp *graphProperties) tryParseEdge(line string) ([]textNode, bool, error) {
	spec, ok := parseEdgeSpecification(line)
	if !ok {
		return nil, false, nil
	}

	var (
		lhs []textNode
		rhs []textNode
		err error
	)

	if lhs, err = gp.parseString(spec.lhs); err != nil {
		lhs = []textNode{gp.parseNode(spec.lhs)}
	}
	if rhs, err = gp.parseString(spec.rhs); err != nil {
		rhs = []textNode{gp.parseNode(spec.rhs)}
	}

	return setArrowWithSpec(lhs, rhs, spec, gp.data), true, nil
}

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
	{open: "([", close: "])", shape: ShapeStadium},
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
	if !useAscii && styleClass != "" {
		def.styleClass = styleClass
		gp.nodeDefinitions[def.id] = def
	}
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
			shape := spec.shape
			if useAscii && shape == ShapeStadium {
				shape = ShapeRoundedRectangle
			}
			content := trimmed[len(spec.open):]
			raw, rest, ok := extractLabel(content, spec.close)
			if !ok {
				return shape, "", "", false
			}
			return shape, raw, rest, true
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

func setArrowWithSpec(lhs, rhs []textNode, spec edgeSpec, data *orderedmap.OrderedMap[string, []textEdge]) []textNode {
	log.Debug("Setting arrow from ", lhs, " to ", rhs, " with operator ", spec.operator, " and label ", spec.label)
	for _, l := range lhs {
		for _, r := range rhs {
			setData(l, textEdge{
				parent:        l,
				child:         r,
				label:         spec.label,
				labelNearHead: spec.nearHead,
				edgeStyle:     spec.edgeStyle,
				arrowDir:      spec.direction,
				leftCap:       spec.leftCap,
				rightCap:      spec.rightCap,
				lenHint:       spec.lenHint,
			}, data)
		}
	}
	return rhs
}

func addNode(node textNode, data *orderedmap.OrderedMap[string, []textEdge]) {
	if _, ok := data.Get(node.id); !ok {
		data.Set(node.id, []textEdge{})
	}
}

func setData(parent textNode, edge textEdge, data *orderedmap.OrderedMap[string, []textEdge]) {
	for element := data.Front(); element != nil; element = element.Next() {
		edge.sourceIndex += len(element.Value)
	}
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
	if strings.TrimSpace(line) == "" {
		// Ignore empty lines
		return []textNode{}, nil
	}

	if nodes, handled, err := gp.tryParseEdge(line); handled {
		return nodes, err
	}

	// Patterns are matched in order
	patterns := []struct {
		regex   *regexp.Regexp
		handler func([]string) ([]textNode, error)
	}{
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
		fields := strings.Fields(trimmedLine)
		if !useAscii && len(fields) > 0 && fields[0] == "class" {
			match := classAssignmentPattern.FindStringSubmatch(trimmedLine)
			if match == nil {
				return &properties, fmt.Errorf("class requires nonempty node IDs and one class name")
			}
			for _, identifier := range strings.Split(match[1], ",") {
				identifier = strings.TrimSpace(identifier)
				definition := properties.getOrCreateNodeDefinition(identifier)
				definition.styleClass = match[2]
				properties.nodeDefinitions[identifier] = definition
			}
			continue
		}
		if !useAscii && len(fields) > 0 && fields[0] == "classDef" {
			if len(fields) < 3 {
				return &properties, fmt.Errorf("classDef requires a class name and styles")
			}
			colors, err := parseColorDeclarations(strings.Join(fields[2:], ""))
			if err != nil {
				return &properties, err
			}
			styleClasses[fields[1]] = styleClass{name: fields[1], styles: colors}
			continue
		}
		if !useAscii && len(fields) > 0 && fields[0] == "linkStyle" {
			rule, err := parseLinkStyle(trimmedLine)
			if err != nil {
				return &properties, err
			}
			properties.linkStyles = append(properties.linkStyles, rule)
			continue
		}

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
	if !useAscii {
		edgeCount := 0
		for element := data.Front(); element != nil; element = element.Next() {
			edgeCount += len(element.Value)
		}
		for _, rule := range properties.linkStyles {
			for _, index := range rule.indices {
				if index >= edgeCount {
					return &properties, fmt.Errorf("linkStyle edge index %d is out of range", index)
				}
			}
		}
	}
	return &properties, nil
}
