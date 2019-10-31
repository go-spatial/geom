package delaunay_test

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/subdivision"
)

func dumpSD(t *testing.T, sd *subdivision.Subdivision) {

	var ml geom.MultiLineString
	err := sd.WalkAllEdges(func(e *quadedge.Edge) error {
		ln := e.AsLine()
		ml = append(ml, ln[:])
		return nil
	})
	if err != nil {
		panic(err)
	}
	t.Logf("sd edges\n%v\n\n", wkt.MustEncode(ml))
}
func dumpSDWithin(t *testing.T, sd *subdivision.Subdivision, start, end geom.Point) {
	line, edges := dumpSDWithinStr(sd, start, end)
	t.Logf("line\n%v\n", line)
	t.Logf("sd edges\n%v\n\n", edges)
}

func dumpSDWithinStr(sd *subdivision.Subdivision, start, end geom.Point) (line, edges string) {
	// get the distance this will be the radius for our two circles
	ptDistance := planar.PointDistance(start, end)
	cStart := geom.Circle{
		Center: [2]float64(start),
		Radius: ptDistance,
	}
	cEnd := geom.Circle{
		Center: [2]float64(end),
		Radius: ptDistance,
	}
	ext := geom.NewExtentFromPoints(cStart.AsPoints(30)...)
	ext1 := geom.NewExtentFromPoints(cEnd.AsPoints(30)...)
	ext.Add(ext1)
	ext = ext.ExpandBy(10)

	var ml geom.MultiLineString
	err := sd.WalkAllEdges(func(e *quadedge.Edge) error {
		ln := e.AsLine()
		if !ext.ContainsPoint(ln[0]) && !ext.ContainsPoint(ln[1]) {
			return nil
		}

		ml = append(ml, ln[:])
		return nil
	})
	if err != nil {
		panic(err)
	}
	line = wkt.MustEncode(geom.Line{[2]float64(start), [2]float64(end)})
	edges = wkt.MustEncode(ml)
	return line, edges
}
