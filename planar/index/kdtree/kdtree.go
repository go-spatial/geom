/*

A two dimensional kd-tree implementation

Use the associated iterator class to query the tree.
*/
package kdtree

import (
	"errors"

	"github.com/go-spatial/geom"
)

type KdTree struct {
	root *KdNode
}

func NewKdTree() *KdTree {
	return &KdTree{}
}

/*
Insert the specified geometry into the kd-tree.

An error will be raised if a duplicate point is inserted.
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
				return nil, errors.New("duplicate node")
				// if the new point is on the left
			} else if p.XY()[d] < currentNode.p.XY()[d] {
				// if there is no left node, populate it
				if currentNode.Left() == nil {
					currentNode.SetLeft(node)
					break
					// if there already is a left node, traverse into it
				} else {
					currentNode = currentNode.Left()
				}
				// if the new point is the same or on the right
			} else {
				currentNode.bbox.AddPoints(p.XY())
				// if there is no right node, populate it
				if currentNode.Right() == nil {
					currentNode.SetRight(node)
					break
					// if there already is a right node, traverse into it
				} else {
					currentNode = currentNode.Right()
				}
			}

			// switch back and forth between the zeroth and first dimensions
			d = d ^ 1
		}
	}

	return node, nil
}
