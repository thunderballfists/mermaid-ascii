package cmd

import "sort"

type groupLane struct {
	nodes    []*node
	children []*groupLane
}

func (rendered *graph) separateGroupLanes() {
	if useAscii || !rendered.hasInterleavedGroups() {
		return
	}
	root := &groupLane{}
	lanes := map[*subgraph]*groupLane{nil: root}
	owners := map[*node]*subgraph{}
	for _, group := range rendered.sortSubgraphsByDepth() {
		lane := &groupLane{}
		lanes[group] = lane
		lanes[group.parent].children = append(lanes[group.parent].children, lane)
		for _, contentNode := range group.nodes {
			owners[contentNode] = group
		}
	}
	for _, contentNode := range rendered.nodes {
		lane := lanes[owners[contentNode]]
		lane.nodes = append(lane.nodes, contentNode)
	}
	root.assign(0)
	rendered.grid = map[gridCoord]*node{}
	for _, contentNode := range rendered.nodes {
		rendered.reserveSpotInGrid(contentNode, contentNode.gridCoord)
	}
	rendered.groupLanes = true
}

func (lane *groupLane) assign(start int) int {
	sort.SliceStable(lane.nodes, func(first, second int) bool {
		if graphDirection == "LR" {
			return lane.nodes[first].gridCoord.y < lane.nodes[second].gridCoord.y
		}
		return lane.nodes[first].gridCoord.x < lane.nodes[second].gridCoord.x
	})
	perRank := map[int]int{}
	width := 0
	for _, contentNode := range lane.nodes {
		primary, secondary := contentNode.gridCoord.x, &contentNode.gridCoord.y
		if graphDirection == "TD" {
			primary, secondary = contentNode.gridCoord.y, &contentNode.gridCoord.x
		}
		*secondary = 4 * (start + perRank[primary])
		perRank[primary]++
		width = Max(width, perRank[primary])
	}
	for _, child := range lane.children {
		width += child.assign(start + width)
	}
	return width
}

func (rendered *graph) hasInterleavedGroups() bool {
	bounds := map[*subgraph][4]int{}
	for _, group := range rendered.subgraphs {
		if len(group.nodes) == 0 {
			continue
		}
		first := group.nodes[0].gridCoord
		box := [4]int{first.x, first.y, first.x, first.y}
		members := map[*node]bool{}
		for _, contentNode := range group.nodes {
			position := contentNode.gridCoord
			box[0], box[1] = Min(box[0], position.x), Min(box[1], position.y)
			box[2], box[3] = Max(box[2], position.x), Max(box[3], position.y)
			members[contentNode] = true
		}
		for _, contentNode := range rendered.nodes {
			position := contentNode.gridCoord
			if !members[contentNode] && position.x >= box[0] && position.x <= box[2] && position.y >= box[1] && position.y <= box[3] {
				return true
			}
		}
		for other, otherBox := range bounds {
			if members[other.nodes[0]] {
				continue
			}
			ancestor := false
			for parent := group.parent; parent != nil; parent = parent.parent {
				ancestor = ancestor || parent == other
			}
			if !ancestor && box[0] <= otherBox[2] && box[2] >= otherBox[0] && box[1] <= otherBox[3] && box[3] >= otherBox[1] {
				return true
			}
		}
		bounds[group] = box
	}
	return false
}
