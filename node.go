package bspc

type (
	// Node contains all the info regarding a given node.
	Node struct {
		ID          ID             `json:"id"`
		SplitType   SplitType      `json:"SplitType"`
		SplitRatio  float64        `json:"splitRatio"`
		Vacant      bool           `json:"vacant"`
		Hidden      bool           `json:"hidden"`
		Sticky      bool           `json:"sticky"`
		Private     bool           `json:"private"`
		Locked      bool           `json:"locked"`
		Marked      bool           `json:"marked"`
		Preselect   *NodePreselect `json:"presel"`
		Rectangle   rectangle      `json:"rectangle"`
		Constraints constraints    `json:"constraints"`
		FirstChild  *Node          `json:"firstChild"`
		SecondChild *Node          `json:"secondChild"`
		Client      *NodeClient    `json:"client"`
	}

	// NodePreselect contains all the infor regarding a node's preselection state.
	NodePreselect struct {
		SplitDirection DirectionType `json:"splitDir"`
		SplitRatio     float64       `json:"splitRatio"`
	}

	// NodeClient contains all the info regarding a node's client. The program it contains.
	NodeClient struct {
		ClassName         string    `json:"className"`
		InstanceName      string    `json:"instanceName"`
		BorderWidth       int       `json:"borderWidth"`
		State             StateType `json:"state"`     // TODO: Add validation for this in the GetState method
		LastState         StateType `json:"lastState"` // TODO: Add validation for this in the GetState method
		Layer             LayerType `json:"layer"`     // TODO: Add validation for this in the GetState method
		LastLayer         LayerType `json:"lastLayer"` // TODO: Add validation for this in the GetState method
		Urgent            bool      `json:"urgent"`
		Shown             bool      `json:"shown"`
		TiledRectangle    rectangle `json:"tiledRectangle"`
		FloatingRectangle rectangle `json:"floatingRectangle"`
	}
)

// IsLeaf returns true if the current node is a leaf node.
func (n Node) IsLeaf() bool {
	return n.FirstChild == nil && n.SecondChild == nil && n.Client != nil
}

// LeafNodes returns the leaf nodes inside this node.
// If this node is a leaf node, it returns it inside a slice.
func (n Node) LeafNodes() []Node {
	if n.IsLeaf() {
		return []Node{n}
	}

	leafNodes := make([]Node, 0)

	if n.FirstChild != nil {
		leafNodes = append(leafNodes, (*n.FirstChild).LeafNodes()...)
	}

	if n.SecondChild != nil {
		leafNodes = append(leafNodes, (*n.SecondChild).LeafNodes()...)
	}

	return leafNodes
}
