package kdtree

import (
	"container/heap"
	"math"

	"github.com/go-spatial/geom"
)

/*
Create a heap for storing nodes so we can get back the nearest neighbors in order.
*/
type HeapEntry struct {
	node *KdNode
	d    float64
}

type KdNodeHeap []*HeapEntry

func (this KdNodeHeap) Len() int            { return len(this) }
func (this KdNodeHeap) Less(i, j int) bool  { return this[i].d < this[j].d }
func (this KdNodeHeap) Swap(i, j int)       { this[i], this[j] = this[j], this[i] }
func (this *KdNodeHeap) Push(x interface{}) { *this = append(*this, x.(*HeapEntry)) }
func (this *KdNodeHeap) Pop() interface{} {
	old := *this
	n := len(old)
	x := old[n-1]
	*this = old[0 : n-1]
	return x
}

/*
The distance function specifies how to calculate the distance from a point to the extent.

In most cases the default euclidean distance function will do just fine.
*/
type DistanceFunc func(p geom.Pointer, e *geom.Extent) float64

func distance(c1 [2]float64, c2 [2]float64) float64 {
	v1 := c2[0] - c1[0]
	v2 := c2[1] - c1[1]
	return math.Sqrt(v1*v1 + v2*v2)
}

func EuclideanDistance(p geom.Pointer, e *geom.Extent) float64 {
	x := p.XY()[0]
	y := p.XY()[1]
	var result float64
	if e.ContainsPoint(p.XY()) {
		result = 0
	} else if x < e.MinX() {
		if y < e.MinY() {
			result = distance(p.XY(), e.Min())
		} else if y > e.MaxY() {
			result = distance(p.XY(), [2]float64{e.MinX(), e.MaxY()})
		} else {
			result = e.MinX() - x
		}
	} else if x > e.MaxX() {
		if y < e.MinY() {
			result = distance(p.XY(), [2]float64{e.MaxX(), e.MinY()})
		} else if y > e.MaxY() {
			result = distance(p.XY(), e.Max())
		} else {
			result = x - e.MaxX()
		}
	} else if y < e.MinY() {
		result = e.MinY() - y
	} else {
		result = y - e.MaxY()
	}

	return result
}

type NearestNeighborIterator struct {
	p         geom.Pointer
	kdTree    *KdTree
	df        DistanceFunc
	currentIt *HeapEntry
	nodeHeap  KdNodeHeap
	bboxHeap  KdNodeHeap
}

/*
The nearest neighbor iterator creates an iterator that returns the neighboring points in descending
order. Retrieving all the results will occur in O(n log(n)) time. Results are calculated in a lazy
fashion so retrieving the nearest point or the nearest handful should still be quite efficient.

Algorithm:

* Initialize by calculating the distance from the source point to the root node and the root node's
  bounding box. Push both these values onto their respective heaps (nodeHeap & bboxHeap)
* While there is still data to return:
	* If the distance on the top of the nodeHeap is < the distance on the top of bboxHeap we know
	  that node is closer than all the remaining nodes and can be returned to the user.
	* Otherwise, the node might not be closest, so pop of the next bounding box and push its
	  children onto the nodeHeap.

The iterator design here was taken from: https://ewencp.org/blog/golang-iterators/index.html

To use this:

	nnit := NewNearestNeighborIterator(from, kdt, EuclideanDistance)

	for nnit.Next() {
		n, d := nnit.Value()
		// do stuff
	}

*/
func NewNearestNeighborIterator(p geom.Pointer, kdTree *KdTree, df DistanceFunc) *NearestNeighborIterator {
	var result NearestNeighborIterator

	result.p = p
	result.kdTree = kdTree
	result.df = df

	result.pushNode(kdTree.root)

	return &result
}

/*
Iterates to the next nearest neighbor. True is returned if there is another nearest neighbor,
otherwise false is returned.
*/
func (this *NearestNeighborIterator) Next() bool {
	done := false

	for !done {
		if this.nodeHeap.Len() == 0 && this.bboxHeap.Len() == 0 {
			done = true
			this.currentIt = nil
		} else if this.nodeHeap.Len() > 0 && ((this.bboxHeap.Len() > 0 && this.nodeHeap[0].d <= this.bboxHeap[0].d) ||
			this.bboxHeap.Len() == 0) {
			this.currentIt = heap.Pop(&this.nodeHeap).(*HeapEntry)
			done = true
		} else {
			parent := heap.Pop(&this.bboxHeap).(*HeapEntry).node

			this.pushNode(parent.Left())
			this.pushNode(parent.Right())
		}
	}

	return this.currentIt != nil
}

/*
Pushes the specified node onto both the bboxHeap and nodeHeap.
*/
func (this *NearestNeighborIterator) pushNode(n *KdNode) {
	if n != nil {
		// push the bounding box distance onto the bbox heap
		d := this.df(this.p, &n.bbox)
		heap.Push(&this.bboxHeap, &HeapEntry{n, d})
		// push the point distance onto the node heap
		d = this.df(this.p, geom.NewExtent(n.p.XY()))
		heap.Push(&this.nodeHeap, &HeapEntry{n, d})
	}
}

/*
Returns the geometry pointer and distance for the current neighbor.
*/
func (this *NearestNeighborIterator) Value() (geom.Pointer, float64) {
	return this.currentIt.node.P(), this.currentIt.d
}
