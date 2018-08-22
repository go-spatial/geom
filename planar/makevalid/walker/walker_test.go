package walker

import (
	"context"
	"log"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func TestNew(t *testing.T) {
	type tcase struct {
		triangles []geom.Triangle
		w         *Walker
	}

	fn := func(t *testing.T, tc tcase) {

		w := New(tc.triangles)
		if !reflect.DeepEqual(*tc.w, *w) {
			t.Errorf("index,\n expected %+v\n got       %+v", *tc.w, *w)
		}

	}
	tests := []tcase{
		{
			triangles: []geom.Triangle{
				{{0, 0}, {10, 0}, {0, 10}},
				{{0, 10}, {10, 0}, {10, 10}},
			},
			w: &Walker{
				Triangles: []geom.Triangle{
					{{0, 0}, {10, 0}, {0, 10}},
					{{0, 10}, {10, 0}, {10, 10}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 0}):  {0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   {0},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 10}):   {0},
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): {1},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): {1},
				},
			},
		},
		{
			triangles: []geom.Triangle{
				{{0, 7}, {7, 7}, {0, 10}},
				{{0, 10}, {7, 7}, {10, 10}},
				{{7, 7}, {10, 0}, {10, 10}},
				{{0, 0}, {10, 0}, {7, 2}},
				{{0, 0}, {7, 2}, {0, 7}},
				{{0, 7}, {7, 2}, {7, 7}},
				{{7, 2}, {10, 0}, {7, 7}},
			},
			w: &Walker{
				Triangles: []geom.Triangle{
					{{0, 7}, {7, 7}, {0, 10}},
					{{0, 10}, {7, 7}, {10, 10}},
					{{7, 7}, {10, 0}, {10, 10}},
					{{0, 0}, {10, 0}, {7, 2}},
					{{0, 0}, {7, 2}, {0, 7}},
					{{0, 7}, {7, 2}, {7, 7}},
					{{7, 2}, {10, 0}, {7, 7}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): {1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   {3},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 7}):    {0, 5},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 0}):   {2, 6},
					sortedEdge([2]float64{0, 7}, [2]float64{0, 10}):   {0},
					sortedEdge([2]float64{0, 10}, [2]float64{7, 7}):   {0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{7, 2}):    {3, 4},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 10}):  {1, 2},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): {2},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 2}):    {4, 5},
					sortedEdge([2]float64{7, 2}, [2]float64{7, 7}):    {5, 6},
					sortedEdge([2]float64{7, 2}, [2]float64{10, 0}):   {3, 6},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 7}):    {4},
				},
			},
		},
		{
			triangles: []geom.Triangle{

				{{0, 0}, {2, 5}, {0, 8}},
				{{0, 8}, {2, 5}, {5, 5}},
				{{0, 8}, {5, 5}, {8, 8}},
				{{5, 5}, {8, 0}, {8, 8}},
				{{0, 0}, {8, 0}, {5, 2}},
				{{0, 0}, {5, 2}, {2, 2}},
				{{0, 0}, {2, 2}, {2, 5}},
				{{5, 2}, {8, 0}, {5, 5}},
			},
			w: &Walker{

				Triangles: []geom.Triangle{

					{{0, 0}, {2, 5}, {0, 8}},
					{{0, 8}, {2, 5}, {5, 5}},
					{{0, 8}, {5, 5}, {8, 8}},
					{{5, 5}, {8, 0}, {8, 8}},
					{{0, 0}, {8, 0}, {5, 2}},
					{{0, 0}, {5, 2}, {2, 2}},
					{{0, 0}, {2, 2}, {2, 5}},
					{{5, 2}, {8, 0}, {5, 5}},
				},

				edgeMap: map[[2][2]float64][]int{

					sortedEdge([2]float64{0, 0}, [2]float64{2, 2}): {5, 6},
					sortedEdge([2]float64{0, 0}, [2]float64{2, 5}): {0, 6},
					sortedEdge([2]float64{0, 8}, [2]float64{8, 8}): {2},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 8}): {0},
					sortedEdge([2]float64{2, 5}, [2]float64{5, 5}): {1},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 8}): {2, 3},
					sortedEdge([2]float64{8, 0}, [2]float64{8, 8}): {3},
					sortedEdge([2]float64{2, 2}, [2]float64{2, 5}): {6},
					sortedEdge([2]float64{0, 0}, [2]float64{5, 2}): {4, 5},
					sortedEdge([2]float64{5, 2}, [2]float64{8, 0}): {4, 7},
					sortedEdge([2]float64{2, 2}, [2]float64{5, 2}): {5},
					sortedEdge([2]float64{0, 0}, [2]float64{8, 0}): {4},
					sortedEdge([2]float64{5, 2}, [2]float64{5, 5}): {7},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 0}): {3, 7},
					sortedEdge([2]float64{0, 8}, [2]float64{2, 5}): {0, 1},
					sortedEdge([2]float64{0, 8}, [2]float64{5, 5}): {1, 2},
				},
			},
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}

}

func TestPolygonForTriangle(t *testing.T) {
	type tcase struct {
		w             *Walker
		idx           []int
		polygons      [][][][2]float64
		seenTriangles [][]int
	}
	fn := func(t *testing.T, idx int, tc tcase) {
		seen := make(map[int]bool)
		plyg := tc.w.PolygonForTriangle(context.Background(), idx, seen)
		seenTriangles := make([]int, 0, len(seen))
		for k := range seen {
			seenTriangles = append(seenTriangles, k)
		}
		sort.Ints(seenTriangles)
		if !reflect.DeepEqual(tc.polygons[idx], plyg) {
			t.Errorf("rings, \n\texpected %v \n\tgot      %v", tc.polygons[idx], plyg)
		}
		if !reflect.DeepEqual(tc.seenTriangles[idx], seenTriangles) {
			t.Errorf("seenTriangles, expected %v got %v", tc.seenTriangles[idx], seenTriangles)
		}
	}

	tests := []tcase{
		{ // simple squire
			w: &Walker{
				Triangles: []geom.Triangle{
					{{0, 0}, {10, 0}, {0, 10}},
					{{0, 10}, {10, 0}, {10, 10}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 0}):  {0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   {0},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 10}):   {0},
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): {1},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): {1},
				},
			},
			idx: []int{0, 1},
			polygons: [][][][2]float64{
				{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
				{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
			},
			seenTriangles: [][]int{
				{0, 1},
				{0, 1},
			},
		},
		{
			w: &Walker{
				Triangles: []geom.Triangle{
					{{0, 7}, {7, 7}, {0, 10}},
					{{0, 10}, {7, 7}, {10, 10}},
					{{7, 7}, {10, 0}, {10, 10}},
					{{0, 0}, {10, 0}, {7, 2}},
					{{0, 0}, {7, 2}, {0, 7}},
					{{7, 2}, {10, 0}, {7, 7}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): {1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   {3},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 7}):    {0},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 0}):   {2, 5},
					sortedEdge([2]float64{0, 7}, [2]float64{0, 10}):   {0},
					sortedEdge([2]float64{0, 10}, [2]float64{7, 7}):   {0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{7, 2}):    {3, 4},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 10}):  {1, 2},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): {2},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 2}):    {4},
					sortedEdge([2]float64{7, 2}, [2]float64{7, 7}):    {5},
					sortedEdge([2]float64{7, 2}, [2]float64{10, 0}):   {3, 5},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 7}):    {4},
				},
			},
			idx: []int{0},
			polygons: [][][][2]float64{
				{{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 7}}, {{0, 7}, {7, 7}, {7, 2}}},
				{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
			},
			seenTriangles: [][]int{
				{0, 1, 2, 3, 4, 5},
				{0, 1},
			},
		},
		{ // 2
			w: &Walker{

				Triangles: []geom.Triangle{

					{{0, 0}, {2, 5}, {0, 8}},
					{{0, 8}, {2, 5}, {5, 5}},
					{{0, 8}, {5, 5}, {8, 8}},
					{{5, 5}, {8, 0}, {8, 8}},
					{{0, 0}, {8, 0}, {5, 2}},
					{{0, 0}, {5, 2}, {2, 2}},
					{{0, 0}, {2, 2}, {2, 5}},
					{{5, 2}, {8, 0}, {5, 5}},
				},

				edgeMap: map[[2][2]float64][]int{

					sortedEdge([2]float64{0, 0}, [2]float64{2, 2}): {5, 6},
					sortedEdge([2]float64{0, 0}, [2]float64{2, 5}): {0, 6},
					sortedEdge([2]float64{0, 8}, [2]float64{8, 8}): {2},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 8}): {0},
					sortedEdge([2]float64{2, 5}, [2]float64{5, 5}): {1},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 8}): {2, 3},
					sortedEdge([2]float64{8, 0}, [2]float64{8, 8}): {3},
					sortedEdge([2]float64{2, 2}, [2]float64{2, 5}): {6},
					sortedEdge([2]float64{0, 0}, [2]float64{5, 2}): {4, 5},
					sortedEdge([2]float64{5, 2}, [2]float64{8, 0}): {4, 7},
					sortedEdge([2]float64{2, 2}, [2]float64{5, 2}): {5},
					sortedEdge([2]float64{0, 0}, [2]float64{8, 0}): {4},
					sortedEdge([2]float64{5, 2}, [2]float64{5, 5}): {7},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 0}): {3, 7},
					sortedEdge([2]float64{0, 8}, [2]float64{2, 5}): {0, 1},
					sortedEdge([2]float64{0, 8}, [2]float64{5, 5}): {1, 2},
				},
			},
			idx: []int{0},
			polygons: [][][][2]float64{
				{{{0, 0}, {8, 0}, {8, 8}, {0, 8}}, {{2, 2}, {2, 5}, {5, 5}, {5, 2}}},
			},
			seenTriangles: [][]int{
				{0, 1, 2, 3, 4, 5, 6, 7},
			},
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for _, idx := range tc.idx {
				idx := idx
				t.Run(strconv.Itoa(idx), func(t *testing.T) { fn(t, idx, tc) })
			}
		})
	}
}

func TestPolygonForRing(t *testing.T) {
	type tcase struct {
		rng  [][2]float64
		plyg [][][2]float64
	}

	fn := func(t *testing.T, tc tcase) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic'd, expected nil, got %v", r)
			}
		}()

		plyg := PolygonForRing(context.Background(), tc.rng)
		if !cmp.PolygonEqual(tc.plyg, plyg) {
			t.Errorf("polygon, expected %v got %v", tc.plyg, plyg)
		}
	}

	tests := [...]tcase{

		{ // issue-12
			rng: [][2]float64{
				{985, 1485}, {986, 1484}, {986, 1483}, {989, 1483}, {991, 1482},
				{993, 1483}, {993, 1485}, {992, 1487}, {989, 1489}, {988, 1490},
				{987, 1489}, {986, 1487}, {988, 1485}, {986, 1487},
			},
			plyg: [][][2]float64{
				{
					{985, 1485}, {986, 1484}, {986, 1483}, {989, 1483}, {991, 1482},
					{993, 1483}, {993, 1485}, {992, 1487}, {989, 1489}, {988, 1490},
					{987, 1489},
				},
			},
		},
	}

	for i, tc := range tests {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}

}
