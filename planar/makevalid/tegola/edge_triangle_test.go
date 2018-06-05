package tegola

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

func TestNewEdgeIndexTriangles(t *testing.T) {
	type tcase struct {
		g    geom.Geometry
		hm   planar.HitMapper
		edix *edgeIndexTriangles
		err  error
	}

	fn := func(t *testing.T, tc tcase) {
		edix, err := newEdgeIndexTriangles(context.Background(), tc.hm, tc.g)
		if tc.err != nil {
			if err == nil || tc.err.Error() != err.Error() {
				t.Errorf("error, expected %v got %v", tc.err, err)
			}
			return
		}
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}

		if !reflect.DeepEqual(*tc.edix, *edix) {
			t.Errorf("index,\n expected %+v\n got       %+v", *tc.edix, *edix)
		}

	}
	tests := []tcase{
		{
			g: geom.MultiLineString{
				{{0, 0}, {10, 0}},
				{{10, 0}, {10, 10}},
				{{10, 10}, {0, 10}},
				{{0, 10}, {0, 0}},
			},
			hm: hitmap.Inside,
			edix: &edgeIndexTriangles{
				triangles: []triangle{
					{{0, 0}, {10, 0}, {0, 10}},
					{{0, 10}, {10, 0}, {10, 10}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 0}):  []int{0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   []int{0},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 10}):   []int{0},
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): []int{1},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): []int{1},
				},
			},
		},
		{
			g: geom.Polygon{
				{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
				{{0, 7}, {7, 7}, {7, 2}},
			},
			hm: hitmap.Inside,
			edix: &edgeIndexTriangles{
				triangles: []triangle{
					{{0, 7}, {7, 7}, {0, 10}},
					{{0, 10}, {7, 7}, {10, 10}},
					{{7, 7}, {10, 0}, {10, 10}},
					{{0, 0}, {10, 0}, {7, 2}},
					{{0, 0}, {7, 2}, {0, 7}},
					{{0, 7}, {7, 2}, {7, 7}},
					{{7, 2}, {10, 0}, {7, 7}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): []int{1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   []int{3},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 7}):    []int{0, 5},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 0}):   []int{2, 6},
					sortedEdge([2]float64{0, 7}, [2]float64{0, 10}):   []int{0},
					sortedEdge([2]float64{0, 10}, [2]float64{7, 7}):   []int{0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{7, 2}):    []int{3, 4},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 10}):  []int{1, 2},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): []int{2},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 2}):    []int{4, 5},
					sortedEdge([2]float64{7, 2}, [2]float64{7, 7}):    []int{5, 6},
					sortedEdge([2]float64{7, 2}, [2]float64{10, 0}):   []int{3, 6},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 7}):    []int{4},
				},
			},
		},
		{
			g: geom.Polygon{
				{{0, 0}, {8, 0}, {8, 8}, {0, 8}},
				{{2, 2}, {2, 5}, {5, 5}, {5, 2}},
			},
			hm: hitmap.MustNewFromPolygons(nil, [][][2]float64{
				{{0, 0}, {8, 0}, {8, 8}, {0, 8}},
				{{2, 2}, {2, 5}, {5, 5}, {5, 2}},
			}),
			edix: &edgeIndexTriangles{

				triangles: []triangle{

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

					sortedEdge([2]float64{0, 0}, [2]float64{2, 2}): []int{5, 6},
					sortedEdge([2]float64{0, 0}, [2]float64{2, 5}): []int{0, 6},
					sortedEdge([2]float64{0, 8}, [2]float64{8, 8}): []int{2},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 8}): []int{0},
					sortedEdge([2]float64{2, 5}, [2]float64{5, 5}): []int{1},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 8}): []int{2, 3},
					sortedEdge([2]float64{8, 0}, [2]float64{8, 8}): []int{3},
					sortedEdge([2]float64{2, 2}, [2]float64{2, 5}): []int{6},
					sortedEdge([2]float64{0, 0}, [2]float64{5, 2}): []int{4, 5},
					sortedEdge([2]float64{5, 2}, [2]float64{8, 0}): []int{4, 7},
					sortedEdge([2]float64{2, 2}, [2]float64{5, 2}): []int{5},
					sortedEdge([2]float64{0, 0}, [2]float64{8, 0}): []int{4},
					sortedEdge([2]float64{5, 2}, [2]float64{5, 5}): []int{7},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 0}): []int{3, 7},
					sortedEdge([2]float64{0, 8}, [2]float64{2, 5}): []int{0, 1},
					sortedEdge([2]float64{0, 8}, [2]float64{5, 5}): []int{1, 2},
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
		edix          *edgeIndexTriangles
		idx           []int
		polygons      [][][][2]float64
		seenTriangles [][]int
	}
	fn := func(t *testing.T, idx int, tc tcase) {
		seen := make(map[int]bool)
		plyg := tc.edix.PolygonForTriangle(context.Background(), idx, seen)
		seenTriangles := make([]int, 0, len(seen))
		for k, _ := range seen {
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
			edix: &edgeIndexTriangles{
				triangles: []triangle{
					{{0, 0}, {10, 0}, {0, 10}},
					{{0, 10}, {10, 0}, {10, 10}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 0}):  []int{0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   []int{0},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 10}):   []int{0},
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): []int{1},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): []int{1},
				},
			},
			idx: []int{0, 1},
			polygons: [][][][2]float64{
				[][][2]float64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
				[][][2]float64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
			},
			seenTriangles: [][]int{
				[]int{0, 1},
				[]int{0, 1},
			},
		},
		{
			edix: &edgeIndexTriangles{
				triangles: []triangle{
					{{0, 7}, {7, 7}, {0, 10}},
					{{0, 10}, {7, 7}, {10, 10}},
					{{7, 7}, {10, 0}, {10, 10}},
					{{0, 0}, {10, 0}, {7, 2}},
					{{0, 0}, {7, 2}, {0, 7}},
					{{7, 2}, {10, 0}, {7, 7}},
				},
				edgeMap: map[[2][2]float64][]int{
					sortedEdge([2]float64{0, 10}, [2]float64{10, 10}): []int{1},
					sortedEdge([2]float64{0, 0}, [2]float64{10, 0}):   []int{3},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 7}):    []int{0},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 0}):   []int{2, 5},
					sortedEdge([2]float64{0, 7}, [2]float64{0, 10}):   []int{0},
					sortedEdge([2]float64{0, 10}, [2]float64{7, 7}):   []int{0, 1},
					sortedEdge([2]float64{0, 0}, [2]float64{7, 2}):    []int{3, 4},
					sortedEdge([2]float64{7, 7}, [2]float64{10, 10}):  []int{1, 2},
					sortedEdge([2]float64{10, 0}, [2]float64{10, 10}): []int{2},
					sortedEdge([2]float64{0, 7}, [2]float64{7, 2}):    []int{4},
					sortedEdge([2]float64{7, 2}, [2]float64{7, 7}):    []int{5},
					sortedEdge([2]float64{7, 2}, [2]float64{10, 0}):   []int{3, 5},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 7}):    []int{4},
				},
			},
			idx: []int{0},
			polygons: [][][][2]float64{
				[][][2]float64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 7}}, {{0, 7}, {7, 7}, {7, 2}}},
				[][][2]float64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
			},
			seenTriangles: [][]int{
				[]int{0, 1, 2, 3, 4, 5},
				[]int{0, 1},
			},
		},
		{ // 2
			edix: &edgeIndexTriangles{

				triangles: []triangle{

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

					sortedEdge([2]float64{0, 0}, [2]float64{2, 2}): []int{5, 6},
					sortedEdge([2]float64{0, 0}, [2]float64{2, 5}): []int{0, 6},
					sortedEdge([2]float64{0, 8}, [2]float64{8, 8}): []int{2},
					sortedEdge([2]float64{0, 0}, [2]float64{0, 8}): []int{0},
					sortedEdge([2]float64{2, 5}, [2]float64{5, 5}): []int{1},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 8}): []int{2, 3},
					sortedEdge([2]float64{8, 0}, [2]float64{8, 8}): []int{3},
					sortedEdge([2]float64{2, 2}, [2]float64{2, 5}): []int{6},
					sortedEdge([2]float64{0, 0}, [2]float64{5, 2}): []int{4, 5},
					sortedEdge([2]float64{5, 2}, [2]float64{8, 0}): []int{4, 7},
					sortedEdge([2]float64{2, 2}, [2]float64{5, 2}): []int{5},
					sortedEdge([2]float64{0, 0}, [2]float64{8, 0}): []int{4},
					sortedEdge([2]float64{5, 2}, [2]float64{5, 5}): []int{7},
					sortedEdge([2]float64{5, 5}, [2]float64{8, 0}): []int{3, 7},
					sortedEdge([2]float64{0, 8}, [2]float64{2, 5}): []int{0, 1},
					sortedEdge([2]float64{0, 8}, [2]float64{5, 5}): []int{1, 2},
				},
			},
			idx: []int{0},
			polygons: [][][][2]float64{
				[][][2]float64{{{0, 0}, {8, 0}, {8, 8}, {0, 8}}, {{2, 2}, {2, 5}, {5, 5}, {5, 2}}},
			},
			seenTriangles: [][]int{
				[]int{0, 1, 2, 3, 4, 5, 6, 7},
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
