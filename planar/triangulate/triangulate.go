package triangulate

import "github.com/go-spatial/geom"

// Triangulator describes an object that can take a set of points and produce
// a triangulation.
type Triangulator interface {

	// SetPoints sets the nodes to be used in the triangulation
	// data is any metadata to be attached to each point.
	// 	This is a one to one mapping. The length of data MUST be either zero,
	// 	or equal to the length of points
	SetPoints(points []geom.Point, data []interface{}) error

	// Triangles returns the triangles that are produced by the triangulation
	// 	If the triangulation uses a frame the includeFrame should be used to
	// 	determine if the triangles touching the frame should be included or not.
	Triangles(includeFrame bool) ([]geom.Triangle, error)
}

// Constrainer is a Triangulator that can take a set of points and ensure that the
// given set of edges (the constraints) exist in the triangulation
type Constrainer interface {
	Triangulator

	// AddConstraints adds constraint lines to the triangulation, this may require
	// the triangulation to be recalculated.
	// data is any metadata to be attached to each constraint.
	// 	This is a one to one mapping. The length of data MUST be either zero,
	// 	or equal to the length of constraints
	AddConstraints(constraints []geom.Line, data []interface{}) error
}
