package subdivision_test

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/subdivision/internal/phenix"
)

func TestInsertSite(t *testing.T) {
	tests := []*phenix.Phenix{

		(&phenix.PointBag{
			TrianglePoints: [3]geom.Point{
				geom.Point{-2000, -5000},
				geom.Point{-700, 10000},
				geom.Point{2000, -5000},
			},
			Points: []geom.Point{
				{-103.71, 8.965},
				{-48.62, 10.678},
				{-16.23, 33.3337},
				{-11.2, 36.286},
				{-11.19, 36.292},
			},
		}).NewPhenix(
			"Dup Test",
			func(pb *phenix.PointBag) []phenix.Check {
				return []phenix.Check{
					pb.NewCheck(false,
						"t0 : p0 t1",
						"t1 : p0 t2",
						"t2 : t1 p0",

						"e0 = t0.onext.sym : t2 t1",
					),
					pb.NewCheck(false,
						"t0 : p0 t1",
						"t1 : p0 p1 t2",
						"t2 : t1 p1 p0",

						"e0 = t0.onext.sym : t2 p1 t1",
						"e1 = t2.onext.onext.sym : t1 p0",
					), // check point 1
					pb.NewCheck(false,
						"t0 : p0 t1",
						"t1 : p0 p2 t2",
						"t2 : t1 p2 p1 p0",

						"e0 = t0.onext.sym : t2 p1 p2 t1",
						"e1 = t2.onext.onext.onext.sym : p2 p0",
						"e2 = t2.onext.onext.sym : t1 p0 p1",
					), // check point 2
					pb.NewCheck(false,
						"t0 : p0 t1",
						"t1 : p0 p3 t2",
						"t2 : t1 p3 p1 p0",

						"e0 = t0.onext.sym : t2 p1 p2 p3 t1",
						"e1 = t2.onext.onext.onext.sym : p3 p2 p0",
						"e2 = e1.onext.onext.sym : p3 p0",
						"e3 = t2.onext.onext.sym : t1 p0 p2 p1",
					), // check point 3
					pb.NewCheck(false,
						"t0 : p0 t1",
						"t1 : p0 p4 t2",
						"t2 : t1 p4 p1 p0",

						"e0 = t0.onext.sym : t2 p1 p2 p3 p4 t1",
						"e1 = t2.onext.onext.onext.sym : p4 p3 p2 p0",
						"e2 = e1.onext.onext.onext.sym : p3 p0",
						"e3 = e2.onext.sym : p1 p4 p0",
						"e4 = e3.onext.onext.sym : p1 t2 t1 p0",
					), // check point 4
				}
			},
		),
	}
	for i := range tests {
		t.Run(tests[i].Name, tests[i].Test)
	}
}
