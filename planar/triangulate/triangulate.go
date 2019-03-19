package triangulate

import (
	"context"

	"github.com/go-spatial/geom"
)

// Triangulator describes an object that can take a set of points and produce
// a triangulation.
type Triangulator interface {

	// SetPoints sets the nodes to be used in the triangulation, this will replace
	// any current points in the triangulation
	SetPoints(ctx context.Context, points ...geom.Point) error

	// Triangles returns the triangles that are produced by the triangulation
	Triangles(ctx context.Context, includeFrame bool) ([]geom.Triangle, error)
}

// IncrementatlTriangulator describes an triangulation where points can be
// incrementally added.
type IncrementalTriangulator interface {
	Triangulator

	// InsertPoint inserts a point into an existing triangulation
	InsertPoints(ctx context.Context, points ...geom.Point) error
}

// ConstrainedTriangulator is a Triangulator that can take a set of points and ensure that the
// given set of edges (the constraints) exist in the triangulation
type ConstrainedTriangulator interface {
	Triangulator

	// AddConstraints adds constraint lines to the triangulation, this may require
	// the triangulation to be recalculated.
	AddConstraints(ctx context.Context, constraints ...geom.Line) error
}

// ConstrainedIncrementalTriangulator is an Incremental Triangulator that can take a set of points and ensure that the
// given set of edges (the constraints) exist in the triangulation
type ConstrainedIncrementalTriangulator interface {
	IncrementalTriangulator
	ConstrainedTriangulator
}
