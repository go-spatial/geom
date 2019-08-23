package qetriangulate_test

import (
	"context"
	"log"
	"math"
	"testing"

	"github.com/gdey/errors"
	"github.com/go-spatial/geom"
	qetriangulate "github.com/go-spatial/geom/planar/triangulate/gdey/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/subdivision"
)

func TestConstraint(t *testing.T) {

	fn := func(tc qetriangulate.Constrained) func(*testing.T) {
		return func(t *testing.T) {
			var err error

			pts := tc.Points
			for _, ct := range tc.Constraints {
				pts = append(pts, ct[0], ct[1])
			}
			var sd *subdivision.Subdivision

			ctx := context.Background()
			sd, err = subdivision.NewForPoints(ctx, pts)
			if err != nil {
				t.Errorf("error, expected nil, got %v", err)
				return
			}
			isValid := sd.IsValid(ctx)

			if !isValid {
				t.Errorf("triangulation is not valid")
				dumpSD(t, sd)
				return
			}

			// Run the constraints as sub tests
			vxidx := sd.VertexIndex()
			total := len(tc.Constraints)

			for i, ct := range tc.Constraints {
				start, end := geom.Point(ct[0]), geom.Point(ct[1])
				if startingEdge, ok := vxidx[start]; ok {
					if e := startingEdge.FindONextDest(end); e != nil {
						// Nothing to do, edge already in the subdivision.
						// t.Logf("%v already in subdivision", ct)
						continue
					}
				}

				t.Logf("Adding Constraint %v of %v", i, total)

				bLine, bEdges := dumpSDWithinStr(sd, start, end)

				err := sd.InsertConstraint(ctx, vxidx, start, end)
				if err != nil {

					log.Printf("errored: %v", err)
					t.Logf("Before:\nLine:%v\nEdges:\n%v", bLine, bEdges)
					t.Logf("failed to add constraint %v of %v", i, total)
					dumpSDWithin(t, sd, start, end)
					t.Errorf("got err: %v", err)
					return
				}
				if !sd.IsValid(ctx) {
					t.Logf("Subdivision not valid")
					dumpSDWithin(t, sd, start, end)
					t.Errorf("Left subdivision in an invalid state")
					return
				}
			}

			dumpSD(t, sd)
			// Get all the triangles and check to see if they are what we expect
			// them to be.
			tri, err := sd.Triangles(false)
			if err != nil {
				t.Errorf("error, expected nil, got %v", err)
				return
			}
			t.Logf("Number of triangles: %v", len(tri))

		}
	}

	tests := packageTests

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestGemConstrained(t *testing.T) {
	type tcase struct {
		Desc         string
		tri          qetriangulate.GeomConstrained
		includeFrame bool
		expTriangles []geom.Triangle
		err          error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			got, err := tc.tri.Triangles(context.Background(), tc.includeFrame)
			if tc.err != nil {
				if err == nil {
					t.Errorf("error, expected %v, got nil", tc.err)
					return
				}
				t.Logf("got error:%v", err)
			}
			if len(got) != len(tc.expTriangles) {
				t.Errorf("number of triangles, expected %v got %v", len(tc.expTriangles), len(got))
				return
			}
		}
	}
	tests := []tcase{
		{
			Desc: "bad lines",
			tri: qetriangulate.GeomConstrained{
				Constraints: []geom.Line{
					{{math.Copysign(0, -1), 4096}, {math.Copysign(0, -1), 4096}},
					{{math.Copysign(0, -1), 4096}, {0, 4096}},
					{{math.Copysign(0, -1), 4096}, {0, 4096}},
					{{0, 4096}, {0, 4096}},
				},
			},
			err: errors.String("invalid points/constraints"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}
