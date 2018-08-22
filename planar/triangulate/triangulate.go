package triangulate

import "github.com/go-spatial/geom"

type Interface interface {
	// SetPoints sets the nodes to be used in the triangulation
	SetPoints(pts []geom.Point, data []interface{})
	// Triangles returns the triangles that are produced by the triangulation
	Triangles() []geom.Triangle
}

type Constrainer interface {
	// AddConstraints adds constraint lines to the triangulation, this may require
	// the triangulation to be recalculated.
	AddConstraints(constraints ...geom.Line)
}
