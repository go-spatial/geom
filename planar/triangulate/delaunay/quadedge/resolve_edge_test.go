package quadedge

import (
	"fmt"
	"testing"

	"github.com/go-spatial/geom"
)

func findEdgeWithDest(e *Edge, dest geom.Point) *Edge {
	var edg *Edge = e
	e.WalkAllONext(func(e *Edge) bool {
		if cmp.GeomPointEqual(dest, *e.Dest()) {
			edg = e
			return false
		}
		return true
	})
	return edg
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
			edge := BuildEdgeGraphAroundPoint(
				tc.origin,
				tc.endpoints...,
			)
			// Validate our test case
			if !tc.noValidation {
				if err := Validate(edge); err != nil {
					if e, ok := err.(ErrInvalid); ok {
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

					got, err := ResolveEdge(startingEdge, tc.dest)
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
			geom.Point{0, 1}: geom.ErrPointsAreCoLinear,
			geom.Point{1, 0}: geom.ErrPointsAreCoLinear,
		},
		genTestStruct{
			dest: geom.Point{0, 6}, // ab = ⟳
			pts: []geom.Point{
				{1, 1}, // case 4
				{0, 1}, // case 7
			},
		},
		genTestStruct{
			dest: geom.Point{6, 0},
			pts: []geom.Point{
				{-1, 1},  // case 9
				{-1, -1}, // case 12
				{1, -1},  // case 13
				{0, -1},  // case 14
				{-1, 0},  // case 15

				{1, 0}, // case 16
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
					geom.Point{0, -1},
					geom.Point{1, 0},
				},
			},
			genTestStruct{
				dest: geom.Point{-6, 0},
				pts: []geom.Point{
					geom.Point{-1, 0},
					geom.Point{0, 1},
				},
			},
		)...,
	)

	tests = append(tests,
		tcase{
			Desc:   "constraint test first triangle",
			origin: geom.Point{369, 793},
			endpoints: []geom.Point{
				geom.Point{426, 539},
				geom.Point{273, 525},
				geom.Point{-2511, -3640},
				geom.Point{475.500, 8853},
			},
			dest:         geom.Point{516, 661},
			expectedDest: geom.Point{475.500, 8853},
		},
		tcase{
			Desc:   "find_intersect_test_00",
			origin: geom.Point{3779.594, 2406.835},
			endpoints: []geom.Point{
				{3778.690, 2405.340},
				{3778.979, 2407.590},
				{3779.301, 2407.278},
				{3780.668, 2407.771},
				{3780.216, 2406.145},
			},
			dest:         geom.Point{3778.301, 2408.181},
			expectedDest: geom.Point{3778.690, 2405.340},
		},
		tcase{
			Desc:   "eastcoast of samerica",
			origin: geom.Point{132.123, 228.096},
			endpoints: []geom.Point{
				{132.123, 229.29},
				{132.69, 229.913},
				{134.545, 226.851},
				{132.971, 226.821},
				{131.231, 227.244},
				{130.794, 228.497},
			},
			dest:         geom.Point{132.123, 226.228},
			expectedDest: geom.Point{132.971, 226.821},
		},
		tcase{
			Desc:   "natural_earth_cities_test_01",
			origin: geom.Point{4082, 310},
			endpoints: []geom.Point{

				{4083, 312},
				{4084, 310},
				{4083, 309},
				{4082, 309},
				{4081, 310},
			},
			dest:         geom.Point{4080, 312},
			expectedDest: geom.Point{4081, 310},
		},
	)
	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}
