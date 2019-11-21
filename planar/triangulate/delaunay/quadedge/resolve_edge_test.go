package quadedge

import (
	"fmt"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/winding"
)

func TestResolveEdge(t *testing.T) {
	type tcase struct {
		edge         *Edge
		order        winding.Order
		dest         geom.Point
		expectedEdge *Edge
		err          error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			gotEdge, gotErr := ResolveEdge(tc.order, tc.edge, tc.dest)
			if tc.err != gotErr {
				t.Errorf("error for %v, expected %v got %v", wkt.MustEncode(tc.dest), tc.err, gotErr)
			}
			if gotEdge != tc.expectedEdge {
				t.Log("edge:", gotEdge)
				t.Errorf("edge  for %v, expected %v got %v", wkt.MustEncode(tc.dest), wkt.MustEncode(tc.expectedEdge.AsLine()), wkt.MustEncode(gotEdge.AsLine()))
			}
		}
	}

	tests := map[string]tcase{}

	{ // build out y-up tests case for edge with POINTS(0 0,5 0, 0 -5).
		//	 where ab is counter clockwise, this will, also, implicitly cover ab clockwise.
		nameFormat := "y-up resolve ccw case %v"
		order := winding.Order{}
		edge := BuildEdgeGraphAroundPoint(
			geom.Point{0, 0},
			geom.Point{5, 0},
			geom.Point{0, -5},
		)
		edge05 := edge.FindONextDest(geom.Point{0, -5})
		edge50 := edge.FindONextDest(geom.Point{5, 0})
		edge = edge05
		cases := []tcase{
			{ // case 0
				dest:         geom.Point{0, 0},
				expectedEdge: nil,
				err:          ErrInvalidEndVertex,
			},
			{ // case 1
				dest:         geom.Point{-3, -3},
				expectedEdge: edge50,
			},
			{ // case 2
				dest:         geom.Point{-3, 3},
				expectedEdge: edge50,
			},
			{ // case 3
				dest:         geom.Point{-3, 0},
				expectedEdge: edge50,
			},
			{ // case 4
				dest:         geom.Point{3, -3},
				expectedEdge: edge05,
			},
			{ // case 5
				dest:         geom.Point{3, 3},
				expectedEdge: edge50,
			},
			{ // case 6
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case 7
				dest:         geom.Point{0, -3},
				expectedEdge: edge05,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case 8
				dest:         geom.Point{0, 3},
				expectedEdge: edge50,
			},
		}

		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}

		nameFormat = "y-up resolve cw %v"
		edge = edge50
		cases = []tcase{
			{ // case cw ccw cw
				dest:         geom.Point{3, -3},
				expectedEdge: edge05,
			},
			{ // case cw ccw zdb
				dest:         geom.Point{0, -3},
				expectedEdge: edge05,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case cw zda cw
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
		}
		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}
		// for colinear edges
		edge = BuildEdgeGraphAroundPoint(
			geom.Point{0, 0},
			geom.Point{5, 0},
			geom.Point{-5, 0},
		)
		nameFormat = "y-up resolve zab 1 %v"
		edge_50 := edge.FindONextDest(geom.Point{-5, 0})
		edge50 = edge.FindONextDest(geom.Point{5, 0})
		edge = edge50
		cases = []tcase{
			{ // case zab ccw cw
				dest:         geom.Point{0, -3},
				expectedEdge: edge_50,
			},
			{ // case zab cw ccw
				dest:         geom.Point{0, 3},
				expectedEdge: edge50,
			},
			{ // case zab zda zdb
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case zab zda zdb
				dest:         geom.Point{-3, 0},
				expectedEdge: edge_50,
				err:          geom.ErrPointsAreCoLinear,
			},
		}
		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}
		// for colinear edges (one edge)
		edge = BuildEdgeGraphAroundPoint(
			geom.Point{0, 0},
			geom.Point{5, 0},
		)
		nameFormat = "y-up resolve zab 2 %v"
		edge50 = edge.FindONextDest(geom.Point{5, 0})
		edge = edge50
		cases = []tcase{
			{ // case zab ccw cw
				dest:         geom.Point{0, -3},
				expectedEdge: edge50,
			},
			{ // case zab cw ccw
				dest:         geom.Point{0, 3},
				expectedEdge: edge50,
			},
			{ // case zab zda zdb
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case zab zda zdb
				dest:         geom.Point{-3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
		}
		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}

	}
	{ // build out y-down tests case for edge with POINTS(0 0,5 0, 0 5).
		//	 where ab is counter clockwise, this will, also, implicitly cover ab clockwise.
		nameFormat := "y-down resolve ccw case %v"
		order := winding.Order{
			YPositiveDown: true,
		}
		edge := BuildEdgeGraphAroundPoint(
			geom.Point{0, 0},
			geom.Point{5, 0},
			geom.Point{0, 5},
		)
		edge05 := edge.FindONextDest(geom.Point{0, 5})
		edge50 := edge.FindONextDest(geom.Point{5, 0})
		edge = edge05
		cases := []tcase{
			{ // case 0
				dest:         geom.Point{0, 0},
				expectedEdge: nil,
				err:          ErrInvalidEndVertex,
			},
			{ // case 1
				dest:         geom.Point{-3, 3},
				expectedEdge: edge50,
			},
			{ // case 2
				dest:         geom.Point{-3, -3},
				expectedEdge: edge50,
			},
			{ // case 3
				dest:         geom.Point{-3, 0},
				expectedEdge: edge50,
			},
			{ // case 4
				dest:         geom.Point{3, 3},
				expectedEdge: edge05,
			},
			{ // case 5
				dest:         geom.Point{3, -3},
				expectedEdge: edge50,
			},
			{ // case 6
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case 7
				dest:         geom.Point{0, 3},
				expectedEdge: edge05,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case 8
				dest:         geom.Point{0, -3},
				expectedEdge: edge50,
			},
		}

		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}

		nameFormat = "y-down resolve cw %v"
		edge = edge50
		cases = []tcase{
			{ // case cw ccw cw
				dest:         geom.Point{3, 3},
				expectedEdge: edge05,
			},
			{ // case cw ccw zdb
				dest:         geom.Point{0, 3},
				expectedEdge: edge05,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case cw zda cw
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
		}
		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}
		// for colinear edges
		edge = BuildEdgeGraphAroundPoint(
			geom.Point{0, 0},
			geom.Point{5, 0},
			geom.Point{-5, 0},
		)
		nameFormat = "y-down resolve zab 1 %v"
		edge_50 := edge.FindONextDest(geom.Point{-5, 0})
		edge50 = edge.FindONextDest(geom.Point{5, 0})
		edge = edge50
		cases = []tcase{
			{ // case zab ccw cw
				dest:         geom.Point{0, 3},
				expectedEdge: edge_50,
			},
			{ // case zab cw ccw
				dest:         geom.Point{0, -3},
				expectedEdge: edge50,
			},
			{ // case zab zda zdb
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case zab zda zdb
				dest:         geom.Point{-3, 0},
				expectedEdge: edge_50,
				err:          geom.ErrPointsAreCoLinear,
			},
		}
		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}
		// for colinear edges (one edge)
		edge = BuildEdgeGraphAroundPoint(
			geom.Point{0, 0},
			geom.Point{5, 0},
		)
		nameFormat = "y-down resolve zab 2 %v"
		edge50 = edge.FindONextDest(geom.Point{5, 0})
		edge = edge50
		cases = []tcase{
			{ // case zab ccw cw
				dest:         geom.Point{0, 3},
				expectedEdge: edge50,
			},
			{ // case zab cw ccw
				dest:         geom.Point{0, -3},
				expectedEdge: edge50,
			},
			{ // case zab zda zdb
				dest:         geom.Point{3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
			{ // case zab zda zdb
				dest:         geom.Point{-3, 0},
				expectedEdge: edge50,
				err:          geom.ErrPointsAreCoLinear,
			},
		}
		for i := range cases {
			cases[i].edge = edge
			cases[i].order = order
			tests[fmt.Sprintf(nameFormat, i)] = cases[i]
		}

	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}

}

/*

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
	order := winding.Order{
		YPositiveDown: true,
	}

	fn := func(tc tcase) func(*testing.T) {

		return func(t *testing.T) {
			edge := BuildEdgeGraphAroundPoint(
				tc.origin,
				tc.endpoints...,
			)
			// Validate our test case
			if !tc.noValidation {
				if err := Validate(edge, order); err != nil {
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

					got, err := ResolveEdge(order, startingEdge, tc.dest)
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

*/
