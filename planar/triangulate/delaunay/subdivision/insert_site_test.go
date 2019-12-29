package subdivision_test

import (
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/subdivision"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/subdivision/internal/phenix"
)

func checks(script string) func(pb *phenix.PointBag) []phenix.Check {
	// assume that the script looks like
	// check: [debug]
	// 	t0 : p0 t1 ...
	//  e0 = t0.onext.sym : t1 t2

	type chxset struct {
		lines []string
		debug bool
	}
	var chks []chxset

	lines := strings.Split(script, "\n")
	for i := range lines {
		ln := strings.TrimSpace(lines[i])
		if ln == "" {
			continue
		}
		if strings.HasPrefix(ln, "//") {
			continue
		}
		if strings.HasPrefix(ln, "check") {
			chks = append(chks, chxset{
				debug: strings.Contains(ln, "debug"),
			})
			continue
		}
		chidx := len(chks) - 1
		if chidx == -1 {
			continue
		}
		chks[chidx].lines = append(chks[chidx].lines, ln)
	}
	return func(pb *phenix.PointBag) (pchks []phenix.Check) {
		for i := range chks {
			pchks = append(pchks, pb.NewCheck(
				chks[i].debug,
				chks[i].lines...,
			))
		}
		return pchks
	}
}
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
		}).NewPhenix("Dup Test", checks(`
			check :
				t0 : p0 t1
				t1 : p0 t2
				t2 : t1 p0

				e0 = t0.onext.sym : t2 t1

			check:
				t0 : p0 t1
				t1 : p0 p1 t2
				t2 : t1 p1 p0

				e0 = t0.onext.sym : t2 p1 t1
				e1 = t2.onext.onext.sym : t1 p0

			check:
				t0 : p0 t1
				t1 : p0 p2 t2
				t2 : t1 p2 p1 p0

				e0 = t0.onext.sym : t2 p1 p2 t1
				e1 = t2.onext.onext.onext.sym : p2 p0
				e2 = t2.onext.onext.sym : t1 p0 p1

			check:
				t0 : p0 t1
				t1 : p0 p3 t2
				t2 : t1 p3 p1 p0

				e0 = t0.onext.sym : t2 p1 p2 p3 t1
				e1 = t2.onext.onext.onext.sym : p3 p2 p0
				e2 = e1.onext.onext.sym : p3 p0
				e3 = t2.onext.onext.sym : t1 p0 p2 p1

			check:
				t0 : p0 t1
				t1 : p0 p4 t2
				t2 : t1 p4 p1 p0

				e0 = t0.onext.sym : t2 p1 p2 p3 p4 t1
				e1 = t2.onext.onext.onext.sym : p4 p3 p2 p0
				e2 = e1.onext.onext.onext.sym : p3 p0
				e3 = e2.onext.sym : p1 p4 p0
				e4 = e3.onext.onext.sym : p1 t2 t1 p0
			`)),
		(&phenix.PointBag{
			TrianglePoints: [3]geom.Point{
				{-13083955.52, 3794319.772},
				{-13050158.8, 4005076.284},
				{-13016362.08, 3794319.772},
			},
			Points: []geom.Point{
				{-13047467.01, 3867715.532},
				{-13047466.45, 3867712.862},
				{-13047465.89, 3867710.192},
			},
		}).NewPhenix("intersecting lines", checks(`
			check :
				t0 : p0 t1
				t1 : p0 t2
				t2 : t1 p0

				e0 = t0.onext.sym : t2 t1

			check :
				t0 : p1 p0 t1
				t1 : p0 t2
				t2 : t1 p0 p1

				e0 = t1.onext.sym : t0 p1 t2
				e1 = t0.onext.sym : t2 p0

			check : [debug]
				t0 : p2 p1 p0 t1
				t1 : p0 t2
				t2 : t1 p0 p1 p2

				e0 = t1.onext.sym : t0 p1 t2
				e1 = t0.onext.onext.sym : p2 t2 p0
				e2 = t0.onext.sym : t2 p1

		`)),
	}
	phenix.ToggleDebug = func() {
		subdivision.ToggleDebug()
	}
	for i := range tests {
		t.Run(tests[i].Name, tests[i].Test)
	}
	phenix.ToggleDebug = nil
}
