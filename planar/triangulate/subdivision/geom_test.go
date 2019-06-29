package subdivision

import (
	"context"
	"log"
	"testing"

	"github.com/go-spatial/geom/planar/triangulate/debugger"
	"github.com/go-spatial/geom/planar/triangulate/geometry"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
	"github.com/go-spatial/geom"
)

func newEdge(a, b, c, d float64) *quadedge.Edge {
	orig := geometry.NewPoint(a, b)
	dest := geometry.NewPoint(c, d)
	return quadedge.NewWithEndPoints(&orig, &dest)
}

/*
func BuildTestCase1() {
	e01 := newEdge(0,3,5,9)
	e02 := newEdge(0,3,5,6)
	e03 := newEdge(0,3,3,3) // edge d
	e04 := newEdge(0,3,3,0)

	e05 := newEdge(5,9,11,7)
	e06 := newEdge(5,9,5,6)

	e07 := newEdge(5,6,11,7)
	e08 := newEdge(5,6,8,4)
	e09 := newEdge(5,6,3,3)

	e10 := newEdge(3,3,7,3)
	e11 := newEdge(3,3,5,1)
	e12 := newEdge(3,3,3,0)
}
*/

func BuildTestCase0() (es []*quadedge.Edge) {
	es = make([]*quadedge.Edge, 5)
	es[0] = newEdge(0, 3, 3, 6)
	es[1] = newEdge(0, 3, 3, 0)

	es[2] = newEdge(3, 6, 6, 6)
	es[3] = newEdge(3, 0, 6, 0)

	es[4] = newEdge(6, 0, 6, 6)

	quadedge.Splice(es[0], es[1])
	quadedge.Splice(es[0].Sym(), es[2])
	quadedge.Splice(es[1].Sym(), es[3])
	quadedge.Splice(es[3].Sym(), es[4])
	quadedge.Splice(es[4].Sym(), es[4].Sym())
	for i, e := range es {
		log.Printf("edge %v : %p <=> %p (%v -> %v) ", i+1, e, e.Sym(), *e.Orig(), *e.Dest())
	}

	return es
}

func TestFindImmediateRightOfEdge(t *testing.T) {
	type tcase struct {
		edges []*quadedge.Edge
		dest  geometry.Point
		// SEidx is the starting index for the SubdivisionEdges to use as
		// the startingedge, it's Origin is going to be the starting point.
		// To keep consistant with the toInd and fromIdx this starts from 1 as well.
		seIdx int
		// ToIdx and FromIdx is 0 means it's nil, it the index is negative it is the sym edge of the edge at abs(index)+1
		// if it's positive it the edge index+1
		toIdx   int
		fromIdx int
	}

	fn := func(ctx context.Context, tc tcase) func(*testing.T) {

		findEdgeIndex := func(e *quadedge.Edge) int {
			for i, ee := range tc.edges {
				if ee == e {
					return i + 1
				}
				if ee == e.Sym() {
					return (i + 1) * -1
				}
			}
			return 0
		}

		edgeAtIndex := func(idx int) *quadedge.Edge {
			switch {
			case idx == 0:
				return nil
			case idx < 0:
				return tc.edges[(idx*-1)-1].Sym()
			default:
				return tc.edges[idx-1]
			}
		}

		return func(t *testing.T) {
			var from, to = edgeAtIndex(tc.fromIdx), edgeAtIndex(tc.toIdx)
			ctx = debugger.SetTestName(ctx, t.Name())
			var showDebug bool
			se := edgeAtIndex(tc.seIdx)

			gotFrom, gotTo := findImmediateRightOfEdges(se, tc.dest)

			if gotFrom != from {
				showDebug = true
				t.Errorf("from, expected edge @%v got edge @%v", tc.fromIdx, findEdgeIndex(gotFrom))
			}
			if gotTo != to {
				showDebug = true
				t.Errorf("to, expected edge @%v got edge @%v", tc.toIdx, findEdgeIndex(gotTo))
			}
			if showDebug {
				_ = WalkAllEdges(tc.edges[0], func(e *quadedge.Edge) error {
					idx := findEdgeIndex(e)
					debugger.Record(ctx, e.AsGeomLine(), debugger.CategoryInput, "subdivision edge %v", idx)
					return nil
				})
				debugger.Record(ctx,
					geom.Line{geometry.UnwrapPoint(*se.Orig()), geometry.UnwrapPoint(tc.dest)},
					debugger.CategoryInput,
					"Edge to add",
				)
				if gotFrom != nil {
					idx := findEdgeIndex(gotFrom)
					debugger.Record(ctx, gotFrom.AsGeomLine(), debugger.CategoryGot, "from edge %v", idx)
				}
				if from != nil {
					debugger.Record(ctx, from.AsGeomLine(), debugger.CategoryExpected, "from edge %v", tc.fromIdx)
				}
				if gotTo != nil {
					idx := findEdgeIndex(gotTo)
					debugger.Record(ctx, gotTo.AsGeomLine(), debugger.CategoryGot, "to edge %v", idx)
				}
				if to != nil {
					debugger.Record(ctx, to.AsGeomLine(), debugger.CategoryExpected, "to edge %v", tc.toIdx)
				}
			}
		}
	}

	tests := map[string]tcase{
		"case0e4dest3,6": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 6),
			seIdx:   4,
			fromIdx: -2,
			toIdx:   3,
		},
		"case0e3dest3,0": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 0),
			seIdx:   3,
			fromIdx: 3,
			toIdx:   -2,
		},
		"case0e-1dest3,0": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 0),
			seIdx:   -1,
			fromIdx: 3,
			toIdx:   -2,
		},
		"case0e1dest3,0": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 0),
			seIdx:   1,
			fromIdx: 2,
		},
	}

	ctx := context.Background()
	ctx = debugger.AugmentContext(ctx, "")
	defer debugger.CloseWait(ctx)

	for name, tc := range tests {
		t.Run(name, fn(ctx, tc))
	}

}

func TestResolveEdge(t *testing.T) {
	type tcase struct {
		edges []*quadedge.Edge
		dest  geometry.Point
		// SEidx is the starting index for the SubdivisionEdges to use as
		// the startingedge, it's Origin is going to be the starting point.
		// To keep consistant with the toInd and fromIdx this starts from 1 as well.
		seIdx int
		// ToIdx and FromIdx is 0 means it's nil, it the index is negative it is the sym edge of the edge at abs(index)+1
		// if it's positive it the edge index+1
		foundIdx int
	}

	fn := func(ctx context.Context, tc tcase) func(*testing.T) {

		findEdgeIndex := func(e *quadedge.Edge) int {
			for i, ee := range tc.edges {
				if ee == e {
					return i + 1
				}
				if ee == e.Sym() {
					return (i + 1) * -1
				}
			}
			return 0
		}

		edgeAtIndex := func(idx int) *quadedge.Edge {
			switch {
			case idx == 0:
				return nil
			case idx < 0:
				return tc.edges[(idx*-1)-1].Sym()
			default:
				return tc.edges[idx-1]
			}
		}

		return func(t *testing.T) {
			ctx = debugger.SetTestName(ctx, t.Name())
			var showDebug bool
			found := edgeAtIndex(tc.foundIdx)
			se := edgeAtIndex(tc.seIdx)

			gotFound := resolveEdge(se, tc.dest)

			if gotFound != found {
				showDebug = true
				t.Errorf("found, expected edge @%v got edge @%v", tc.foundIdx, findEdgeIndex(gotFound))
			}
			if showDebug {
				_ = WalkAllEdges(tc.edges[0], func(e *quadedge.Edge) error {
					idx := findEdgeIndex(e)
					debugger.Record(ctx, e.AsGeomLine(), debugger.CategoryInput, "subdivision edge %v", idx)
					return nil
				})
				debugger.Record(ctx,
					geom.Line{geometry.UnwrapPoint(*se.Orig()), geometry.UnwrapPoint(tc.dest)},
					debugger.CategoryInput,
					"Edge to add",
				)
				if gotFound != nil {
					idx := findEdgeIndex(gotFound)
					debugger.Record(ctx, gotFound.AsGeomLine(), debugger.CategoryGot, "found edge %v", idx)
				}
				if found != nil {
					debugger.Record(ctx, found.AsGeomLine(), debugger.CategoryExpected, "found edge %v", tc.foundIdx)
				}
			}
		}
	}

	tests := map[string]tcase{
		"case0e4dest3,6": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 6),
			seIdx:   4,
			foundIdx: -2,
		},
		"case0e3dest3,0": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 0),
			seIdx:   3,
			foundIdx: 3,
		},
		"case0e-1dest3,0": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 0),
			seIdx:   -1,
			foundIdx: 3,
		},
		"case0e1dest3,0": {
			edges:   BuildTestCase0(),
			dest:    geometry.NewPoint(3, 0),
			seIdx:   1,
			foundIdx: 1,
		},
	}

	ctx := context.Background()
	ctx = debugger.AugmentContext(ctx, "")
	defer debugger.CloseWait(ctx)

	for name, tc := range tests {
		t.Run(name, fn(ctx, tc))
	}
}
