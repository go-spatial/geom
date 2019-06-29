package subdivision

import (
	"github.com/go-spatial/geom/planar/triangulate/geometry"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
)

func (t Triangle) AsGeom() (tri geom.Triangle) {
	e := t.StartingEdge()
	for i := 0; i < 3; e, i = e.RNext(), i+1 {
		tri[i] = geometry.UnwrapPoint(*e.Orig())
	}
	return tri
}

func (sd *Subdivision) EdgesAsGeom() (lines []geom.Line) {
	_ = sd.WalkAllEdges(func(e *quadedge.Edge) error {
		org := *e.Orig()
		dst := *e.Dest()
		lines = append(lines, geom.Line{
			geometry.UnwrapPoint(org),
			geometry.UnwrapPoint(dst),
		})
		return nil
	})
	return lines
}

func NewSubdivisionFromGeomLines(frame [3]geom.Point, lines []geom.Line) *Subdivision {
	lines = planar.NormalizeUniqueLines(lines)
	sd := New(
		geometry.NewPoint(frame[0][0], frame[0][1]),
		geometry.NewPoint(frame[1][0], frame[1][1]),
		geometry.NewPoint(frame[2][0], frame[2][1]),
	)
	indexMap := make(map[geometry.Point]*quadedge.Edge)
	addEdgeToIndexMap := func(e *quadedge.Edge) {
		if e == nil {
			return
		}
		var ok bool
		orig := *e.Orig()
		dest := *e.Dest()
		if _, ok = indexMap[orig]; !ok  {
			indexMap[orig] = e
		}
		if _, ok = indexMap[dest]; !ok {
			indexMap[dest] = e.Sym()
		}
	}

	// Fill out the indexMap.
	_ = sd.WalkAllEdges(func(e *quadedge.Edge) error {
		addEdgeToIndexMap(e)
		return nil
	})

	for _, l := range lines {
		if l.LenghtSquared() == 0.0 {
			continue
		}

		p0 := geometry.NewPoint(l[0][0], l[0][1])
		p1 := geometry.NewPoint(l[1][0], l[1][1])
		p0edge := indexMap[p0]
		p1edge := indexMap[p1]

		if p0edge == nil && p1edge == nil {
			// skipp edges that isn't in the network currently?
			continue
		}

		// assume both points are going to be added to the graph
		sd.ptcount += 2
		p0p1edge := quadedge.New()
		p0p1edge.EndPoints(&p0, &p1)
		if p0edge != nil {
			// remove this point from count since it already
			// exitst in the graph
			sd.ptcount--
			quadedge.Splice(p0edge, p0p1edge)
		}
		if p1edge != nil {
			// remove this point from count since it already
			// exitst in the graph
			sd.ptcount--
			quadedge.Splice(p1edge, p0p1edge.Sym())
		}
		addEdgeToIndexMap(p0p1edge)
	}
	return sd
}
