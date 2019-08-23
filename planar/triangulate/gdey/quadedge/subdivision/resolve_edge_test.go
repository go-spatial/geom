package subdivision

import (
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"
)

func findEdgeWithDest(e *quadedge.Edge, dest geom.Point) *quadedge.Edge {
	if cmp.GeomPointEqual(dest, *e.Dest()) {
		return e
	}
	ne := e.ONext()
	for ne != e {
		if cmp.GeomPointEqual(dest, *ne.Dest()) {
			return ne
		}
		ne = ne.ONext()
	}
	return e
}

func TestResolveEdge(t *testing.T) {
	type tcase struct {
		Desc         string
		origin       geom.Point
		endpoints    []geom.Point
		dest         geom.Point
		expectedDest geom.Point
		err          error
		noValidation bool
	}

	fn := func(tc tcase) func(*testing.T) {

		return func(t *testing.T) {
			edge := quadedge.BuildEdgeGraphAroundPoint(
				tc.origin,
				tc.endpoints...,
			)
			// Validate our test case
			if !tc.noValidation {
				if err := quadedge.Validate(edge); err != nil {
					if e, ok := err.(quadedge.ErrInvalid); ok {
						for i, estr := range e {
							t.Logf("err %03v: %v", i, estr)
						}
					}
					t.Logf("Endpoints: %v", tc.endpoints)
					t.Errorf("Failed: %v", err)
					return
				}
			}

			expectedEdge := findEdgeWithDest(edge, tc.expectedDest)

			for _, ep := range tc.endpoints {
				startingEdge := findEdgeWithDest(edge, ep)
				t.Run(wkt.MustEncode(startingEdge.AsLine()), func(t *testing.T) {

					got, err := resolveEdge(startingEdge, tc.dest)
					if tc.err != err {
						t.Errorf("error, expected %v got %v", tc.err, err)
					}
					if !expectedEdge.IsEqual(got) {
						t.Logf("edges: %v", startingEdge.DumpAllEdges())
						t.Logf("dest: %v", wkt.MustEncode(tc.dest))
						t.Logf("ONext: %v", wkt.MustEncode(got.ONext().AsLine()))
						t.Errorf("resolve: expected: %v got %v", wkt.MustEncode(expectedEdge.AsLine()), wkt.MustEncode(got.AsLine()))
					}

				})
			}
		}
	}

	type genTestStruct struct {
		dest geom.Point // EdgeDest
		pts  []geom.Point
	}

	genTests := func(descFormat string, o geom.Point, errmap map[geom.Point]error, mappings ...genTestStruct) (tests []tcase) {
		var (
			seen      = make(map[geom.Point]bool, len(mappings))
			endpoints = make([]geom.Point, len(mappings))
		)

		for i, m := range mappings {
			if seen[m.dest] {
				panic(fmt.Sprintf("bad genTests, dest (%v)[%v] already seen", i, m.dest))
			}

			seen[m.dest] = true
			endpoints[i] = m.dest

			for _, dest := range m.pts {
				tests = append(tests,
					tcase{
						Desc:         fmt.Sprintf(descFormat, m.dest),
						origin:       o,
						endpoints:    endpoints,
						expectedDest: m.dest,
						dest:         dest,
						err:          errmap[dest],
					})
			}
		}
		return tests
	}

	tests := genTests(
		"core2Vec %v",
		geom.Point{0, 0},
		map[geom.Point]error{
			geom.Point{4, 0}: geom.ErrPointsAreCoLinear,
			geom.Point{1, 3}: geom.ErrPointsAreCoLinear,
		},
		genTestStruct{
			dest: geom.Point{2, 6},
			pts: []geom.Point{
				{3, -1},
				{2, -2},
				{1, -3},
				{0, -4},
				{-1, -3},
				{-2, -2},
				{-3, -1},
				{-4, 0},
				{-3, 1},
				{-2, 2},
				{-1, 3},
				{0, 4},
				{1, 3},
			},
		},
		genTestStruct{
			dest: geom.Point{6, 0},
			pts: []geom.Point{
				{2, 2},
				{3, 1},
				{4, 0},
			},
		},
	)
	tests = append(tests,
		genTests(
			"ab colinear %v",
			geom.Point{0, 0},
			map[geom.Point]error{
				geom.Point{1, 0}:  geom.ErrPointsAreCoLinear,
				geom.Point{-1, 0}: geom.ErrPointsAreCoLinear,
			},
			genTestStruct{
				dest: geom.Point{6, 0},
				pts: []geom.Point{
					geom.Point{1, 0},
					geom.Point{0, 1},
				},
			},
			genTestStruct{
				dest: geom.Point{-6, 0},
				pts: []geom.Point{
					geom.Point{0, -1},
					geom.Point{-1, 0},
				},
			},
		)...,
	)

	tests = append(tests,
		tcase{
			Desc:   "constraint test first triangle",
			origin: geom.Point{369, 793},
			endpoints: []geom.Point{
				geom.Point{475.500, 8853},
				geom.Point{-2511, -3640},
				geom.Point{273, 525},
				geom.Point{426, 539},
			},
			dest:         geom.Point{516, 661},
			expectedDest: geom.Point{426, 539},
		},
		tcase{
			Desc:   "find_intersect_test_00",
			origin: geom.Point{3779.594, 2406.835},
			endpoints: []geom.Point{
				{3780.216, 2406.145},
				{3780.668, 2407.771},
				{3779.301, 2407.278},
				{3778.979, 2407.590},
				{3778.690, 2405.340},
			},
			dest:         geom.Point{3778.301, 2408.181},
			expectedDest: geom.Point{3778.979, 2407.590},
		},
		tcase{
			Desc:   "eastcoast of samerica",
			origin: geom.Point{132.123, 228.096},
			endpoints: []geom.Point{
				//	{132.123,229.29 },
				//	{132.971,226.821},
				//	{134.545,226.851},
				//	{132.69 ,229.913},
				//	{130.794,228.497},
				//	{131.231,227.244},

				{130.794, 228.497},
				{131.231, 227.244},
				{132.971, 226.821},
				{134.545, 226.851},
				{132.69, 229.913},
				{132.123, 229.29},
			},
			dest:         geom.Point{132.123, 226.228},
			expectedDest: geom.Point{131.231, 227.244},
		},
		tcase{
			Desc:   "natural_earth_cities_test_01",
			origin: geom.Point{4082, 310},
			endpoints: []geom.Point{

				{4081, 310},
				{4082, 309},
				{4083, 309},
				{4084, 310},
				{4083, 312},
			},
			dest:         geom.Point{4080, 312},
			expectedDest: geom.Point{4083, 312},
		},
	)
	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}

type ByCounterClockwise []geom.Point

func (b ByCounterClockwise) Len() int {
	return len(b)
}

func (b ByCounterClockwise) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByCounterClockwise) Less(i, j int) bool {
	return (b[i][0]*b[j][1])-(b[i][1]*b[j][0]) < 0
}

func AsCounterClockwise(pts []geom.Point) []geom.Point {
	ccw := ByCounterClockwise(pts)
	log.Printf("pts: %v", pts)
	sort.Sort(ccw)
	log.Printf("pts: %v", pts)
	return pts
}
