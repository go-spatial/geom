package kdtree

import (
	"container/heap"
	"math"

	"github.com/go-spatial/geom"
)

/*
heapEntry is an entry in the heap for storing nodes so we can get back the nearest neighbors
in order.
*/
type heapEntry struct {
	node *KdNode
	d    float64
}

/*
kdNodeHeap is an array of heap entries.

We are not using a pointer in this array for simplicity and to avoid the dereference. This is
likely ok due to the small size of heapEntry. If for some reason heapEntry gets larger then this
should be changed to a pointer. Also, benchmarks may reveal that a pointer is faster. Dunno.

https://stackoverflow.com/questions/27622083/performance-slices-of-structs-vs-slices-of-pointers-to-structs
*/
type kdNodeHeap []heapEntry

// Implements the container/heap interface.
func (knh kdNodeHeap) Len() int           { return len(knh) }
func (knh kdNodeHeap) Less(i, j int) bool { return knh[i].d < knh[j].d }
func (knh kdNodeHeap) Swap(i, j int)      { knh[i], knh[j] = knh[j], knh[i] }
func (knh *kdNodeHeap) Push(x interface{}) {
	// this will simply panic if the wrong type is passed.
	he, ok := x.(*heapEntry)
	if !ok {
		panic("the wrong interface type was passed to kdNodeHeap.")
	}
	*knh = append(*knh, *he)
}
func (knh *kdNodeHeap) Pop() interface{} {
	old := *knh
	n := len(old)
	x := old[n-1]
	*knh = old[0 : n-1]
	return x
}

/*
DistanceFunc specifies how to calculate the distance from a point to the extent.

In most cases the default EuclideanDistance function will do just fine.
*/
type DistanceFunc func(p geom.Pointer, e *geom.Extent) float64

func distance(c1 [2]float64, c2 [2]float64) float64 {
	v1 := c2[0] - c1[0]
	v2 := c2[1] - c1[1]
	return math.Sqrt(v1*v1 + v2*v2)
}

func EuclideanDistance(p geom.Pointer, e *geom.Extent) float64 {
	if e.ContainsPoint(p.XY()) {
		return 0
	}

	x := p.XY()[0]
	y := p.XY()[1]

	if x < e.MinX() {
		if y < e.MinY() {
			return distance(p.XY(), e.Min())
		}
		if y > e.MaxY() {
			return distance(p.XY(), [2]float64{e.MinX(), e.MaxY()})
		}
		return e.MinX() - x
	}

	if x > e.MaxX() {
		if y < e.MinY() {
			return distance(p.XY(), [2]float64{e.MaxX(), e.MinY()})
		}
		if y > e.MaxY() {
			return distance(p.XY(), e.Max())
		}
		return x - e.MaxX()
	}

	if y < e.MinY() {
		return e.MinY() - y
	}

	return y - e.MaxY()
}

type NearestNeighborIterator struct {
	p         geom.Pointer
	kdTree    *KdTree
	df        DistanceFunc
	currentIt *heapEntry
	nodeHeap  kdNodeHeap
	bboxHeap  kdNodeHeap
}

/*
NewNearestNeighborIterator creates an iterator that returns the neighboring points in descending
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

To use this iterator:

	nnit := NewNearestNeighborIterator(geom.Point{0,0}, myKdTree, EuclideanDistance)

	for nnit.Next() {
		n, d := nnit.Value()
		// do stuff
	}

*/
func NewNearestNeighborIterator(p geom.Pointer, kdTree *KdTree, df DistanceFunc) *NearestNeighborIterator {
	result := NearestNeighborIterator{
		p:      p,
		kdTree: kdTree,
		df:     df,
	}

	result.pushNode(kdTree.root)

	return &result
}

/*
Next iterates to the next nearest neighbor. True is returned if there is another nearest neighbor,
otherwise false is returned.
*/
func (nni *NearestNeighborIterator) Next() bool {
	for {
		if nni.nodeHeap.Len() == 0 && nni.bboxHeap.Len() == 0 {
			nni.currentIt = nil
			break
		}
		if nni.nodeHeap.Len() > 0 && ((nni.bboxHeap.Len() > 0 && nni.nodeHeap[0].d <= nni.bboxHeap[0].d) ||
			nni.bboxHeap.Len() == 0) {
			he := (heap.Pop(&nni.nodeHeap).(heapEntry))
			nni.currentIt = &he
			break
		}

		parent := heap.Pop(&nni.bboxHeap).(heapEntry).node

		nni.pushNode(parent.Left())
		nni.pushNode(parent.Right())
	}

	return nni.currentIt != nil
}

// pushNode pushes the specified node onto both the bboxHeap and nodeHeap.
func (nni *NearestNeighborIterator) pushNode(n *KdNode) {
	if n == nil {
		return
	}

	// push the bounding box distance onto the bbox heap
	d := nni.df(nni.p, &n.bbox)
	heap.Push(&nni.bboxHeap, &heapEntry{n, d})
	// push the point distance onto the node heap
	d = nni.df(nni.p, geom.NewExtent(n.p.XY()))
	heap.Push(&nni.nodeHeap, &heapEntry{n, d})
}

// Value returns the geometry pointer and distance for the current neighbor.
func (nni *NearestNeighborIterator) Value() (geom.Pointer, float64) {
	return nni.currentIt.node.P(), nni.currentIt.d
}
