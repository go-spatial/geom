// kdtree is a two dimensional kd-tree implementation
package kdtree

import (
	"errors"

	"github.com/go-spatial/geom"
)

/*
KdTree is an index for 2 dimensional point data.

Limitations:

* Bulk inserts and balancing are not supported. If you have a large amount of data to insert it is
  best to randomize the data before inserting.
* Duplicate points are not supported and will return an error.

See the *_iterator.go files for how to query data out of the kd-tree.

*/
type KdTree struct {
	root *KdNode
}

var ErrDuplicateNode = errors.New("duplicate node")

/*
Insert the specified geometry into the kd-tree.

If a duplicate point is inserted, the currently indexed point will be returned along with an error.
*/
func (kdt *KdTree) Insert(p geom.Pointer) (*KdNode, error) {
	node := NewKdNode(p)

	if kdt.root == nil {
		kdt.root = node
		return node, nil
	}

	currentNode := kdt.root

	// toggle between dimensions 0 and 1
	for d := 0; ; d = d ^ 1 {
		cxy := currentNode.p.XY()
		currentNode.bbox.AddPoints(p.XY())

		switch {
		// if the new point is a duplicate
		case p.XY()[0] == cxy[0] && p.XY()[1] == cxy[1]:
			return currentNode, ErrDuplicateNode

		// if the new point is on the left
		case p.XY()[d] < currentNode.p.XY()[d]:
			if currentNode.Left() == nil {
				// if there is no left node, populate it
				currentNode.SetLeft(node)
				return node, nil
			}
			// if there already is a left node, traverse into it
			currentNode = currentNode.Left()

		// if the new point is the same or on the right
		default:
			if currentNode.Right() == nil {
				// if there is no right node, populate it
				currentNode.SetRight(node)
				return node, nil
			}

			// traverse into the right node
			currentNode = currentNode.Right()
		}
	}

	return node, nil
}
