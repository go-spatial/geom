package subdivision

import (
	"context"
	"log"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/test/must"
	"github.com/go-spatial/geom/winding"
)

func TestFindIntersectingEdges(t *testing.T) {
	type tcase struct {
		Desc string
		// Lines that make up the subdivision
		Lines []geom.Line

		// this is the intersecting line
		Start geom.Point
		End   geom.Point

		// Use these to force a starting and ending edge
		StartingDest *geom.Point
		EndingDest   *geom.Point

		Order winding.Order

		// expected intersected edges
		ExpectedLines []geom.Line
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			var (
				ctx          = context.Background()
				err          error
				sd           *Subdivision
				startingEdge *quadedge.Edge
				endingEdge   *quadedge.Edge

				pts [][2]float64
			)

			log.Printf("Starting test: %v", t.Name())

			for _, ln := range tc.Lines {
				pts = append(pts, ln[0], ln[1])
			}
			sd, err = NewForPoints(ctx, tc.Order, pts)
			if err != nil {
				panic(err)
			}

			vertexIndex := sd.VertexIndex()
			for i, ln := range tc.Lines {

				start, end := geom.Point(tc.Lines[i][0]), geom.Point(tc.Lines[i][1])
				if _, _, exists, _ := ResolveStartingEndingEdges(sd.Order, vertexIndex, start, end); exists {
					continue
				}
				t.Logf("Adding edge %05v of %05v -- %v", i, len(tc.Lines), wkt.MustEncode(tc.Lines[i]))

				if err = sd.InsertConstraint(ctx, vertexIndex, geom.Point(ln[0]), geom.Point(ln[1])); err != nil {
					t.Logf("Failed to insert: %05v - %v", i, wkt.MustEncode(ln))
					t.Logf("Err:%v", err)
					t.FailNow()
				}
				if err = sd.Validate(ctx); err != nil {
					t.Logf("Failed to validate: %05v - %v", i, wkt.MustEncode(ln))
					t.Logf("Err:%v", err)
					t.FailNow()
				}
			}

			sd.WalkAllEdges(func(ee *quadedge.Edge) error {
				switch {
				case cmp.GeomPointEqual(tc.Start, *ee.Orig()):
					if startingEdge != nil {
						break
					}
					if tc.StartingDest != nil && !cmp.GeomPointEqual(*tc.StartingDest, *ee.Dest()) {
						break
					}
					startingEdge = ee

				case cmp.GeomPointEqual(tc.Start, *ee.Dest()):
					if startingEdge != nil {
						break
					}
					if tc.StartingDest != nil && !cmp.GeomPointEqual(*tc.StartingDest, *ee.Orig()) {
						break
					}
					startingEdge = ee.Sym()

				case cmp.GeomPointEqual(tc.End, *ee.Orig()):
					if endingEdge != nil {
						break
					}
					if tc.EndingDest != nil && !cmp.GeomPointEqual(*tc.EndingDest, *ee.Dest()) {
						break
					}
					endingEdge = ee

				case cmp.GeomPointEqual(tc.End, *ee.Dest()):
					if endingEdge != nil {
						break
					}
					if tc.EndingDest != nil && !cmp.GeomPointEqual(*tc.EndingDest, *ee.Orig()) {
						break
					}
					endingEdge = ee.Sym()

				}
				return nil

			})

			if startingEdge == nil {
				//	t.Logf("lines: %v", tc.Lines)
				dumpSD(t, sd)
				t.Errorf("Failed to find startingEdge:%v", tc.Start)
				return
			}
			if endingEdge == nil {
				dumpSD(t, sd)
				t.Errorf("Failed to find endingEdge:%v", tc.End)
				return
			}

			displaysd := false
			log.Printf("\n\nStarting Test\n\n")

			t.Logf("Starting edge: %v\nEnding edge: %v\n%v",
				wkt.MustEncode(startingEdge.AsLine()),
				wkt.MustEncode(endingEdge.AsLine()),
				wkt.MustEncode(
					geom.Line{[2]float64(startingEdge.AsLine()[0]), [2]float64(endingEdge.AsLine()[0])},
				),
			)
			if debug {
				displaysd = true
			}

			gotLines, err := FindIntersectingEdges(sd.Order, startingEdge, endingEdge)
			if err != nil {
				displaysd = true
				t.Errorf("error, expected nil, got: %v", err)
			}
			if len(gotLines) != len(tc.ExpectedLines) {
				t.Errorf("lines len, expected %v, got %v", len(tc.ExpectedLines), len(gotLines))
				displaysd = true
				for i, ln := range gotLines {
					t.Logf("%03v:%v", i, wkt.MustEncode(ln.AsLine()))
				}
			}
			if displaysd {
				dumpSD(t, sd)
			}
		}
	}

	tests := []tcase{
		// Subtests go here
		{
			Desc: "first_issue",
			// starting edge: LINESTRING (2674.923 3448.779,2676.168 3439.720)
			// ending edge: LINESTRING (2687.408 3432.536,2685.657 3436.985)
			// intersecting edge: LINESTRING (2674.923 3448.779,2687.408 3432.536)
			Lines: must.ReadMultilines("testdata/intersecting_lines_97_trucated.lines"),
			Start: geom.Point{2674.923, 3448.779},
			End:   geom.Point{2687.408, 3432.536},
			ExpectedLines: must.ParseMultilines([]byte(
				`MULTILINESTRING((2676.168 3439.720,2678.653 3446.005),(2676.168 3439.720,2678.653 3446.005),(2676.168 3439.720,2685.657 3436.985),(2680.390 3431.154,2685.657 3436.985))`,
			)),
		},
		{
			Desc: "asia issue",
			// starting edge: LINESTRING (1469.542 3159.987,1482.934 3156.923)
			// ending edge: LINESTRING (1492.312 3183.492,1484.801 3180.385)
			// intersecting edge: LINESTRING (1469.542 3159.987,1492.312 3183.492)
			Lines: must.ReadMultilines("testdata/asia_issue.lines"),
			Start: geom.Point{1469.542, 3159.987},
			End:   geom.Point{1492.312, 3183.492},
			ExpectedLines: []geom.Line{
				{{1470.727, 3163.057}, {1482.934, 3156.923}},
				{{1471.468, 3164.378}, {1482.934, 3156.923}},
				{{1471.468, 3164.378}, {1483.423, 3157.457}},
				{{1472.934, 3168.822}, {1483.423, 3157.457}},
				{{1473.957, 3170.003}, {1483.423, 3157.457}},
				{{1473.957, 3170.003}, {1483.912, 3157.698}},
				{{1475.934, 3171.850}, {1483.912, 3157.698}},
				{{1475.934, 3171.850}, {1488.957, 3164.796}},
				{{1478.786, 3173.961}, {1488.957, 3164.796}},
				{{1478.786, 3173.961}, {1489.897, 3166.707}},
				{{1482.134, 3176.436}, {1489.897, 3166.707}},
				{{1482.934, 3177.233}, {1489.897, 3166.707}},
				{{1483.497, 3178.046}, {1489.897, 3166.707}},
				{{1483.497, 3178.046}, {1498.379, 3177.820}},
				{{1484.268, 3179.580}, {1498.379, 3177.820}},
				{{1484.801, 3180.385}, {1498.379, 3177.820}},
			},
		},
		{
			// starting edge: LINESTRING (4080 312,4081 310)
			// ending edge: LINESTRING (4082 310,4082 309)
			// intersecting edge: LINESTRING (4080 312,4082 310)
			Lines:         must.ReadMultilines("testdata/find_intersects_test_02.lines"),
			Start:         geom.Point{4080, 312},
			End:           geom.Point{4082, 310},
			ExpectedLines: []geom.Line{},
		},
		{
			// starting edge: LINESTRING (4080 312,4081 310)
			// ending edge: LINESTRING (4082 310,4082 309)
			// intersecting edge: LINESTRING (4080 312,4082 310)
			Lines: must.ReadMultilines("testdata/find_intersects_test_02.lines"),
			Start: geom.Point{4081, 310},
			End:   geom.Point{4083, 312},
			ExpectedLines: []geom.Line{
				{{4082, 310}, {4080, 312}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}

}
