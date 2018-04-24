package kdtree

import (
	"encoding/json"

	"github.com/go-spatial/geom"
)

/*
KdNode structure contains the nodes within the tree. Both leaf and non-leaf nodes store the
point data.

If a node has a nil left & right child it is a leaf node.
*/
type KdNode struct {
	p     geom.Pointer
	left  *KdNode
	right *KdNode
	// bbox is the extent of this node and all its children.
	bbox geom.Extent
}

// NewKdNode creates a new node with a properly initialized bbox.
func NewKdNode(p geom.Pointer) *KdNode {
	var result KdNode

	result.p = p
	result.bbox = *geom.NewExtent(p.XY())

	return &result
}

// Left is the node's left child
func (node *KdNode) Left() *KdNode {
	return node.left
}

// Right is the node's right child.
func (node *KdNode) Right() *KdNode {
	return node.right
}

// MarshalJSON is the marshalling function for JSON.
func (node *KdNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		P     geom.Pointer
		Left  *KdNode `json:",omitempty"`
		Right *KdNode `json:",omitempty"`
	}{
		node.p,
		node.left,
		node.right,
	})
}

// P returns the associated point geometry.
func (node *KdNode) P() geom.Pointer {
	return node.p
}

// SetLeft sets the left child.
func (node *KdNode) SetLeft(left *KdNode) {
	node.left = left
}

// SetRight sets the right child.
func (node *KdNode) SetRight(right *KdNode) {
	node.right = right
}
