package quadedge

import (
	"context"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestFindONextDest(t *testing.T) {

	type tcase struct {
		startingEdge *Edge
		dest         geom.Point
		expected     *Edge
	}

	fn := func(ctx context.Context, tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			got := tc.startingEdge.FindONextDest(tc.dest)
			if got != tc.expected {
				t.Errorf("edge, expected %p got %p", tc.expected, got)
			}
		}
	}
	newTCase := func(graph []geom.Point, dest geom.Point) (tc tcase) {
		tc.dest = dest
		// first let's construct our edges.
		// and edges will be graph[0],graph[1...n]
		if len(graph) < 2 {
			panic("graph should have at least two points for on edge.")
		}
		tc.startingEdge = NewWithEndPoints(&graph[0], &graph[1])
		if cmp.GeomPointEqual(dest, graph[1]) {
			tc.expected = tc.startingEdge

		}
		if len(graph) > 2 {
			for i, dst := range graph[2:] {
				if tc.expected != nil && cmp.GeomPointEqual(dest, dst) {
					panic("Duplicate dest in list of dest.")
				}
				e := NewWithEndPoints(&graph[0], &graph[i+2])
				Splice(tc.startingEdge, e)
				if tc.expected == nil && cmp.GeomPointEqual(dest, dst) {
					tc.expected = e
				}
			}
		}
		return tc
	}
	tests := map[string]tcase{
		"nf nil": tcase{},
		"nf one edge": newTCase([]geom.Point{
			geom.Point{0, 0},
			geom.Point{10, 10},
		},
			geom.Point{10, 101},
		),
		"nf three edges": newTCase([]geom.Point{
			geom.Point{0, 0},
			geom.Point{10, 10},
			geom.Point{20, 20},
			geom.Point{30, 30},
		},
			geom.Point{20, 20},
		),
		"one edge": newTCase([]geom.Point{
			geom.Point{0, 0},
			geom.Point{10, 10},
		},
			geom.Point{10, 10},
		),
		"two edges": newTCase([]geom.Point{
			geom.Point{0, 0},
			geom.Point{10, 10},
			geom.Point{20, 20},
		},
			geom.Point{20, 20},
		),
		"three edges": newTCase([]geom.Point{
			geom.Point{0, 0},
			geom.Point{10, 10},
			geom.Point{20, 20},
			geom.Point{30, 30},
		},
			geom.Point{20, 20},
		),
	}
	ctx := context.Background()
	for name, tc := range tests {
		t.Run(name, fn(ctx, tc))
	}
}

func TestValidate(t *testing.T) {

	type tcase struct {
		desc string
		edge *Edge
		err  ErrInvalid
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			e := Validate(tc.edge)
			err, _ := e.(ErrInvalid)
			if len(err) != len(tc.err) {
				t.Errorf("len(error), expected %v got %v", len(tc.err), len(err))
				for i, estr := range err {
					t.Logf("error %v: %v", i, estr)
				}
			}
			t.Logf("edges:\n%v", tc.edge.DumpAllEdges())
		}
	}

	tests := []tcase{
		{
			desc: "empty dest",
			edge: NewWithEndPoints(&geom.Point{1, 0}, nil),
			err: ErrInvalid{
				"expected edge to have dest",
			},
		},
		{
			desc: "empty orig",
			edge: NewWithEndPoints(nil, &geom.Point{1, 0}),
			err: ErrInvalid{
				"expected edge to have origin",
			},
		},
		{
			desc: "bad origins",
			edge: func() *Edge {
				org := geom.Point{375, 113}
				ed1 := BuildEdgeGraphAroundPoint(
					org,
					geom.Point{368, 117},
					geom.Point{372, 112},
				)

				ed2 := NewWithEndPoints(&geom.Point{378, 113}, &geom.Point{384, 114})
				Splice(ed1, ed2)
				return ed1
			}(),
			err: ErrInvalid{
				"orig not equal for edge",
			},
		},
		{
			desc: "only one edge",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{372, 114},
			),
		},
		{
			desc: "only two edges",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{372, 114},
				geom.Point{368, 117},
			),
		},
		{
			desc: "initial good",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{384, 112},
				geom.Point{368, 117},
				geom.Point{372, 114},
			),
		},
		{
			desc: "initial good, rotate 1",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{368, 117},
				geom.Point{372, 114},
				geom.Point{384, 112},
			),
		},
		{
			desc: "initial good, rotate 2",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{372, 114},
				geom.Point{384, 112},
				geom.Point{368, 117},
			),
		},
		{
			desc: "initial bad",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{368, 117},
				geom.Point{384, 112},
				geom.Point{372, 114},
			),
			err: ErrInvalid{
				"expected to be counterclockwise: POINT (474.388 101.957) POINT (280.132 144.623) POINT (288.176 162.614)",
				"expected to be counterclockwise: POINT (280.132 144.623) POINT (288.176 162.614) POINT (474.388 101.957)",
				"expected to be counterclockwise: POINT (288.176 162.614) POINT (474.388 101.957) POINT (280.132 144.623)",
			},
		},
		{
			desc: "initial bad, same point",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{368, 114},
				geom.Point{368, 114},
			),
			err: ErrInvalid{
				"dest not unique",
			},
		},
		{
			desc: "initial good four point",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{384, 112},
				geom.Point{376, 119},
				geom.Point{368, 117},
				geom.Point{372, 114},
			),
		},
		{
			desc: "initial bad four point clockwise ",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{384, 112},
				geom.Point{372, 114},
				geom.Point{368, 117},
				geom.Point{376, 119},
			),
			err: ErrInvalid{
				"expected to be counterclockwise: POINT (366.318 117.961) POINT (376.644 122.864) POINT (384.939 111.896)",
				"expected to be counterclockwise: POINT (376.644 122.864) POINT (384.939 111.896) POINT (365.513 116.162)",
				"expected to be counterclockwise: POINT (384.939 111.896) POINT (365.513 116.162) POINT (366.318 117.961)",
				"expected to be counterclockwise: POINT (365.513 116.162) POINT (366.318 117.961) POINT (376.644 122.864)",
			},
		},
		{
			desc: "initial bad four point ",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{384, 112},
				geom.Point{372, 114},
				geom.Point{376, 119},
				geom.Point{368, 117},
			),
			err: ErrInvalid{
				"expected to be counterclockwise: POINT (288.176 162.614) POINT (474.388 101.957) POINT (280.132 144.623)",
				"expected to be counterclockwise: POINT (474.388 101.957) POINT (280.132 144.623) POINT (391.440 211.639)",
			},
		},
		{
			desc: "points in counterclockwise order",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{204, 694},
				geom.Point{-2511, -3640},
				geom.Point{273, 525},
				geom.Point{426, 539},
				geom.Point{369, 793},
				geom.Point{475.500, 8853},
			),
		},
		{
			desc: "points in clockwise order",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{204, 694},
				geom.Point{-2511, -3640},
				geom.Point{475.500, 8853},
				geom.Point{369, 793},
				geom.Point{426, 539},
				geom.Point{273, 525},
			),
			err: ErrInvalid{
				"expected to be counterclockwise: POINT (285.993 636.753) POINT (241.799 601.419) POINT (150.912 609.255)",
				"expected to be counterclockwise: POINT (241.799 601.419) POINT (150.912 609.255) POINT (207.326 793.945)",
				"expected to be counterclockwise: POINT (150.912 609.255) POINT (207.326 793.945) POINT (289.749 745.450)",
				"expected to be counterclockwise: POINT (207.326 793.945) POINT (289.749 745.450) POINT (285.993 636.753)",
				"expected to be counterclockwise: POINT (289.749 745.450) POINT (285.993 636.753) POINT (241.799 601.419)",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, fn(tc))
	}

}
