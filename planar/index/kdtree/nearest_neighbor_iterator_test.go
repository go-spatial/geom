package kdtree

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/go-spatial/geom"
)

// TestQuery tests a simple predefined query
func TestQuery(t *testing.T) {
	type tcase struct {
		points     []geom.Point
		queryPoint geom.Point
		eOrder     []string
		err        error
	}

	fn := func(t *testing.T, tc tcase) {
		var err error

		kdt := new(KdTree)

		for _, pt := range tc.points {
			if _, err = kdt.Insert(pt); err != nil {
				break
			}
		}

		if tc.err == nil && err != nil {
			t.Errorf("insert points error, expected nil, got %v", err)
			return
		}

		if tc.err != nil {
			if err == nil || err.Error() != tc.err.Error() {
				t.Errorf("error, expected %v got %v", tc.err, err)
			}
			return
		}

		uut := NewNearestNeighborIterator(tc.queryPoint, kdt, EuclideanDistance)

		i := 0
		for uut.Next() {
			n, d := uut.Value()
			gJSON, err := json.Marshal(n)
			if err != nil {
				t.Fatalf("converting to json error, expected nil, got %v", err)
				return
			}

			s := fmt.Sprintf("%s:%.2g", string(gJSON), d)

			if tc.eOrder[i] != s {
				t.Errorf("nearest neighbor iterator, expected %v got %v", tc.eOrder[i], s)
			}
			i++
		}
	}

	tests := map[string]tcase{
		"good": {
			points: []geom.Point{
				{0, 0},
				{1, 0},
				{1, 1},
				{-1, 0},
			},
			queryPoint: geom.Point{2, 2},
			eOrder: []string{
				"[1,1]:1.4",
				"[1,0]:2.2",
				"[0,0]:2.8",
				"[-1,0]:3.6",
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

// TestHeap verifies the basic functionality of the heapEntry & container/heap
func TestHeap(t *testing.T) {
	type tcase struct {
		distances []float64
	}

	fn := func(t *testing.T, tc tcase) {
		// dummy is a node placeholder that isn't relevant to the test
		var dummy KdNode
		var h kdNodeHeap

		for i := 0; i < len(tc.distances)/2; i++ {
			h = append(h, heapEntry{&dummy, tc.distances[i]})
		}

		heap.Init(&h)

		for i := len(tc.distances) / 2; i < len(tc.distances); i++ {
			heap.Push(&h, &heapEntry{&dummy, tc.distances[i]})
		}

		var gDistances []float64
		for h.Len() > 0 {
			gDistances = append(gDistances, heap.Pop(&h).(heapEntry).d)
		}

		sort.Float64s(tc.distances)

		if reflect.DeepEqual(tc.distances, gDistances) == false {
			t.Errorf("heap results are in the wrong order. expected %v got %v", tc.distances, gDistances)
		}
	}

	tests := map[string]tcase{
		"simple": {
			distances: []float64{2, 1, 5, 3, 8},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

/*
TestRandomQueries creates a kdtree full of random pointers and then executes multiple
NewNearestNeighborIterator against the tree to validate the results.

Due to the random nature of this test the typical tcase table structure is not being utilized.
*/
func TestRandomQueries(t *testing.T) {
	kdt := new(KdTree)

	// make the tests consistent with seed of zero.
	rng := rand.New(rand.NewSource(0))

	// create a kd-tree filled with random points
	pointCount := 200
	for i := 0; i < pointCount; i++ {
		kdt.Insert(geom.Point{rng.Float64() * 1000, rng.Float64() * 1000})
	}

	// iterate over the tree several times and verify the results.
	for i := 0; i < 10; i++ {
		from := geom.Point{rng.Float64() * 1500, rng.Float64() * 1500}
		uut := NewNearestNeighborIterator(from, kdt, EuclideanDistance)

		// verify that every point is touched exactly once.
		touchedSet := make(map[string]bool)

		// keep track of the last distance so we can verify it goes in ascending order.
		lastD := -1.0
		for uut.Next() {
			n, d := uut.Value()
			// simple key for uniquely identifying a point. The odds of two random points being
			// identical are astronomical.
			keyBytes, err := json.Marshal(n)
			if err != nil {
				t.Errorf("converting to json error, expected nil, got %v", err)
				return
			}
			key := string(keyBytes)

			// verify the distance is correct.
			dx := n.XY()[0] - from.XY()[0]
			dy := n.XY()[1] - from.XY()[1]
			td := math.Sqrt(dx*dx + dy*dy)
			if td != d {
				t.Errorf("distances was not the expected value, expected %f got %f", td, d)
			}

			if d < lastD {
				t.Errorf("distances were not in ascending order, expected <%.3g got %.3g", lastD, d)
			}
			lastD = d

			if val, ok := touchedSet[key]; ok && val {
				t.Errorf("value was returned multiple times in the iterator, expected 1 got >1")
				return
			}
			touchedSet[key] = true
		}

		if len(touchedSet) != pointCount {
			t.Errorf("unexpected number of points returned, expected %d got %d", pointCount, len(touchedSet))
		}
	}

}
