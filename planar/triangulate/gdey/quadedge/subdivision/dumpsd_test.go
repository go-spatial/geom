package subdivision

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"
)

func dumpSD(t *testing.T, sd *Subdivision) {

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
func dumpSDWithin(t *testing.T, sd *Subdivision, start, end geom.Point) {
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
	t.Logf("line\n%v\n", wkt.MustEncode(geom.Line{[2]float64(start), [2]float64(end)}))
	t.Logf("sd edges\n%v\n\n", wkt.MustEncode(ml))
}
