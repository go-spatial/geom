package quadedge

import (
	"context"
	"testing"

	"github.com/go-spatial/geom/planar/triangulate/geometry"
)

func TestFindONextDest(t *testing.T) {

	type tcase struct {
		startingEdge *Edge
		dest         geometry.Point
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
	newTCase := func(graph []geometry.Point, dest geometry.Point) (tc tcase) {
		tc.dest = dest
		// first let's construct our edges.
		// and edges will be graph[0],graph[1...n]
		if len(graph) < 2 {
			panic("graph should have at least two points for on edge.")
		}
		tc.startingEdge = NewWithEndPoints(&graph[0], &graph[1])
		if geometry.ArePointsEqual(dest, graph[1]) {
			tc.expected = tc.startingEdge

		}
		if len(graph) > 2 {
			for i, dst := range graph[2:] {
				if tc.expected != nil && geometry.ArePointsEqual(dest, dst) {
					panic("Duplicate dest in list of dest.")
				}
				e := NewWithEndPoints(&graph[0], &graph[i+2])
				Splice(tc.startingEdge, e)
				if tc.expected == nil && geometry.ArePointsEqual(dest, dst) {
					tc.expected = e
				}
			}
		}
		return tc
	}
	tests := map[string]tcase{
		"nf nil": tcase{},
		"nf one edge": newTCase([]geometry.Point{
			geometry.NewPoint(0, 0),
			geometry.NewPoint(10, 10),
		},
			geometry.NewPoint(10, 101),
		),
		"nf three edges": newTCase([]geometry.Point{
			geometry.NewPoint(0, 0),
			geometry.NewPoint(10, 10),
			geometry.NewPoint(20, 20),
			geometry.NewPoint(30, 30),
		},
			geometry.NewPoint(20, 20),
		),
		"one edge": newTCase([]geometry.Point{
			geometry.NewPoint(0, 0),
			geometry.NewPoint(10, 10),
		},
			geometry.NewPoint(10, 10),
		),
		"two edges": newTCase([]geometry.Point{
			geometry.NewPoint(0, 0),
			geometry.NewPoint(10, 10),
			geometry.NewPoint(20, 20),
		},
			geometry.NewPoint(20, 20),
		),
		"three edges": newTCase([]geometry.Point{
			geometry.NewPoint(0, 0),
			geometry.NewPoint(10, 10),
			geometry.NewPoint(20, 20),
			geometry.NewPoint(30, 30),
		},
			geometry.NewPoint(20, 20),
		),
	}
	ctx := context.Background()
	for name, tc := range tests {
		t.Run(name, fn(ctx, tc))
	}

}
