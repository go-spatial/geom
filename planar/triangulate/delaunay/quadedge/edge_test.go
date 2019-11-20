package quadedge

import (
	"context"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/winding"
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
	order := winding.Order{
		YPositiveDown: false,
	}
	//const debug = true

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			e := Validate(tc.edge, order)
			err, _ := e.(ErrInvalid)
			if len(err) != len(tc.err) {
				t.Errorf("len(error), expected %v got %v", len(tc.err), len(err))
				for i, estr := range err {
					t.Logf("error %v: %v", i, estr)
				}
			}
			if debug {
				t.Logf("expected errors:")
				for i, estr := range tc.err {
					t.Logf("error %v: %v", i, estr)
				}
				t.Logf("got      errors:")
				for i, estr := range err {
					t.Logf("error %v: %v", i, estr)
				}
			}
			t.Logf("edges:\n%v", tc.edge.DumpAllEdges())
		}
	}

	tests := []tcase{
		{
			desc: "nil edge",
			err:  ErrInvalid{"expected edge to have origin"},
		},
		{
			desc: "empty dest",
			edge: NewWithEndPoints(&geom.Point{1, 0}, nil),
			err: ErrInvalid{
				"expected edge to have dest",
			},
		},
		{
			desc: "one edge",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{0, 0},
				geom.Point{0, 1},
			),
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
				"expected edge 3 to have same origin POINT (375 113) instead of POINT (378 113)",
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
			desc: "colinear counterclockwise",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{0, 0},
				geom.Point{-2, 1},
				geom.Point{-2, 0},
				geom.Point{-2, -1},
			),
		},
		{
			desc: "colinear clockwise",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{0, 0},
				geom.Point{-2, -1},
				geom.Point{-2, 0},
				geom.Point{-2, 1},
			),
			err: ErrInvalid{
				"expected all points to be counter-clockwise: MULTIPOINT (-2 -1,-2 0,-2 1)",
			},
		},
		{
			desc: "initial good",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{372, 114},
				geom.Point{384, 112},
				geom.Point{368, 117},
			),
		},
		{
			desc: "initial good, rotate 1",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{384, 112},
				geom.Point{368, 117},
				geom.Point{372, 114},
			),
		},
		{
			desc: "initial good, rotate 2",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{368, 117},
				geom.Point{372, 114},
				geom.Point{384, 112},
			),
		},
		{
			desc: "initial bad",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{372, 114},
				geom.Point{368, 117},
				geom.Point{384, 112},
			),
			err: ErrInvalid{
				"expected all points to be counter-clockwise: MULTIPOINT (368 117,384 112,372 114)",
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
				"MULTILINESTRING ((375 113,368 114),(375 113,368 114))",
			},
		},
		{
			desc: "initial good four point",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{0, 0},
				geom.Point{-1, 0},
				geom.Point{0, -1},
				geom.Point{1, 0},
				geom.Point{0, 1},
			),
		},
		{
			desc: "initial bad four point clockwise ",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{0, 0},
				geom.Point{-1, 0},
				geom.Point{0, 1},
				geom.Point{1, 0},
				geom.Point{0, -1},
			),
			err: ErrInvalid{
				"expected all points to be counter-clockwise: MULTIPOINT (-1 0,-1 0,0 -1,1 0,0 1)",
			},
		},
		{
			desc: "initial bad four point ",
			edge: BuildEdgeGraphAroundPoint(
				geom.Point{375, 113},
				geom.Point{368, 117},
				geom.Point{376, 119},
				geom.Point{372, 114},
				geom.Point{384, 112},
			),
			err: ErrInvalid{
				"found self interstion for vertics POINT (376 119) and POINT (384 112)",
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
				"expected all points to be counter-clockwise: MULTIPOINT (-2511 -3640,-2511 -3640,273 525,426 539,369 793,475.500 8853)",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, fn(tc))
	}

}
