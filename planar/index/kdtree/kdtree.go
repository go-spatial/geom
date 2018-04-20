/*

A two dimensional kd-tree implementation

*/
package kdtree

import (
	"errors"

	"github.com/go-spatial/geom"
)

type KdTree struct {
	root *KdNode
}

/*

Creates a empty kd-tree to be populated.

Limitations:

* Bulk inserts and balancing are not supported. If you have a large amount of data to insert it is
  best to randomize the data before inserting.
* Duplicate points are not supported and will return an error.

See the *_iterator.go files for how to query data out of the kd-tree.

*/
func NewKdTree() *KdTree {
	return &KdTree{}
}

/*
Insert the specified geometry into the kd-tree.

If a duplicate point is inserted, the currently indexed point will be returned along with an error.
*/
func (this *KdTree) Insert(p geom.Pointer) (*KdNode, error) {
	node := NewKdNode(p)

	if this.root == nil {
		this.root = node
	} else {
		currentNode := this.root

		// dimension being evaluated in the loop
		d := 0

		for {
			currentNode.bbox.AddPoints(p.XY())
			if p.XY()[0] == currentNode.p.XY()[0] &&
				p.XY()[1] == currentNode.p.XY()[1] {
				// if the new point is on the left
				return currentNode, errors.New("duplicate node")
			} else if p.XY()[d] < currentNode.p.XY()[d] {
				if currentNode.Left() == nil {
					// if there is no left node, populate it
					currentNode.SetLeft(node)
					break
				} else {
					// if there already is a left node, traverse into it
					currentNode = currentNode.Left()
				}
			} else {
				// if the new point is the same or on the right
				currentNode.bbox.AddPoints(p.XY())
				if currentNode.Right() == nil {
					// if there is no right node, populate it
					currentNode.SetRight(node)
					break
				} else {
					// if there already is a right node, traverse into it
					currentNode = currentNode.Right()
				}
			}

			// switch back and forth between the zeroth and first dimensions
			d = d ^ 1
		}
	}

	return node, nil
}
