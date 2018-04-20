package kdtree

import (
	"encoding/json"

	"github.com/go-spatial/geom"
)

/*
The KdNode structure contains the nodes within the tree. Both leaf and non-leaf nodes store the
point data.

If a node has a nil left & right child it is a leaf node.
*/
type KdNode struct {
	p     geom.Pointer
	left  *KdNode
	right *KdNode
	// the extent of this node and all its children.
	bbox geom.Extent
}

/*
Create a new node with a properly initialized bbox.
*/
func NewKdNode(p geom.Pointer) *KdNode {
	var result KdNode

	result.p = p
	result.bbox = *geom.NewExtent(p.XY())

	return &result
}

/*
The node's left child.
*/
func (this *KdNode) Left() *KdNode {
	return this.left
}

/*
A marshalling function that is useful when testing/debugging.
*/
func (this *KdNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		P     geom.Pointer
		Left  *KdNode `json:",omitempty"`
		Right *KdNode `json:",omitempty"`
	}{
		this.p,
		this.left,
		this.right,
	})
}

/*
Returns the node's right child.
*/
func (this *KdNode) Right() *KdNode {
	return this.right
}

/*
Returns the point geometry associated with the node.
*/
func (this *KdNode) P() geom.Pointer {
	return this.p
}

/*
Set the left child.
*/
func (this *KdNode) SetLeft(left *KdNode) {
	this.left = left
}

/*
Set the right child.
*/
func (this *KdNode) SetRight(right *KdNode) {
	this.right = right
}
