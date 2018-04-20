package kdtree

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/go-spatial/geom"
)

/*
Test a simple predefined query
*/
func TestQuery(t *testing.T) {
	kdt := NewKdTree()

	kdt.Insert(geom.Point{0, 0})
	kdt.Insert(geom.Point{1, 0})
	kdt.Insert(geom.Point{1, 1})
	kdt.Insert(geom.Point{-1, 0})

	uut := NewNearestNeighborIterator(geom.Point{2, 2}, kdt, EuclideanDistance)

	s := ""
	for uut.Next() {
		n, d := uut.Value()
		s += fmt.Sprintf("%s:%.2g  ", toJson(t, n), d)
	}
	assertEqual(t, "[1,1]:1.4  [1,0]:2.2  [0,0]:2.8  [-1,0]:3.6  ", s)
}

/*
First time with a go-heap, lemme make sure it works.
*/
func TestHeap(t *testing.T) {
	var dummy KdNode
	h := KdNodeHeap{
		{&dummy, 2},
		{&dummy, 1},
		{&dummy, 5},
		{&dummy, 3},
		{&dummy, 8},
	}

	heap.Init(&h)
	heap.Push(&h, &HeapEntry{&dummy, 0})
	heap.Push(&h, &HeapEntry{&dummy, 10})

	var s string
	for h.Len() > 0 {
		s += fmt.Sprintf(" %f", heap.Pop(&h).(*HeapEntry).d)
	}
	assertEqual(t, " 0.000000 1.000000 2.000000 3.000000 5.000000 8.000000 10.000000", s)
}

/*
Test random data and some random queries
*/
func TestRandomQueries(t *testing.T) {
	kdt := NewKdTree()

	rng := rand.New(rand.NewSource(0))

	pointCount := 200
	for i := 0; i < pointCount; i++ {
		kdt.Insert(geom.Point{rng.Float64() * 1000, rng.Float64() * 1000})
	}

	for i := 0; i < 10; i++ {
		from := geom.Point{rng.Float64() * 1500, rng.Float64() * 1500}
		uut := NewNearestNeighborIterator(from, kdt, EuclideanDistance)

		touchedSet := make(map[string]bool)
		lastD := -1.0
		for uut.Next() {
			n, d := uut.Value()
			key := toJson(t, n)

			dx := n.XY()[0] - from.XY()[0]
			dy := n.XY()[1] - from.XY()[1]
			td := math.Sqrt(dx*dx + dy*dy)
			if td != d {
				t.Fatalf("distances was not the expected value: %f != %f", d, td)
			}

			if d < lastD {
				t.Fatalf("distances were not in ascending order: %.3g < %.3g", d, lastD)
			}
			lastD = d

			if val, ok := touchedSet[key]; ok && val {
				t.Fatalf("Value was returned multiple times in the iterator: %s", key)
			}
			touchedSet[key] = true
		}

		if len(touchedSet) != pointCount {
			t.Fatalf("Expected %d points, but received %d points.", pointCount, len(touchedSet))
		}
	}

}
