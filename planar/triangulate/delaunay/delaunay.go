package delaunay

import (
	"errors"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

var (
	// ErrInvalidPseudoPolygonSize is caused due to the size of the polygon
	ErrInvalidPseudoPolygonSize  = errors.New("invalid polygon, not enough points.")
	ErrUnableToUpdateVertexIndex = errors.New("unable to update vertex index")
)

// Builder is a utility class which create a Telaunay Triangulation from a collection of points.
type Builder struct {
	// Tolerance for vertexs comparision
	Tolerance float64

	// siteCoords are the points in the triangulation.
	siteCoords []quadedge.Vertex
	// subdiv is the quadEdge Subdivisions
	subdiv *quadedge.QuadEdgeSubdivision

	// This is for debugging purposes
	recorder debugger.Recorder
}

func (b *Builder) debugRecord(geom interface{}, category string, descriptionFormat string, data ...interface{}) {
	if debug {
		debugger.RecordFFLOn(
			b.recorder,
			debugger.FFL(0),
			geom,
			category,
			descriptionFormat, data...,
		)
	}
}

func (b *Builder) debugAugementRecorder() debugger.Recorder {
	b.recorder, _ = debugger.AugmentRecorder(b.recorder, debugger.FFL(1).Func)
	return b.recorder
}

func New(tolerance float64, points ...geom.Point) (b Builder) {

	// Make a copy so we don't mess with the original.
	pts := make([]geom.Point, len(points))
	for i := range points {
		pts[i] = geom.Point(points[i])
	}

	uniquePoints := planar.SortUniquePoints(pts)
	b.Tolerance = tolerance
	b.siteCoords = make([]quadedge.Vertex, len(uniquePoints))

	// free up memory.
	for i := range uniquePoints {
		b.siteCoords[i] = quadedge.Vertex(uniquePoints[i])
	}
	return b
}

func (b *Builder) setRecorder(r debugger.Recorder) {
	b.recorder = r
}

func (b *Builder) initSubdiv() error {
	if debug {
		defer b.debugAugementRecorder().Close()
	}

	if b.subdiv != nil {
		log.Println("subdiv not nil")
		return nil
	}
	if len(b.siteCoords) == 0 {
		return errors.New("No site coords provided.")
	}
	siteEnv := geom.NewExtent([2]float64(b.siteCoords[0]))
	for i := 1; i < len(b.siteCoords); i++ {
		if debug {
			b.debugRecord(
				[2]float64(b.siteCoords[i]),
				DebuggerCategoryBuilder.With("point", i),
				"initial point %v", i,
			)
		}
		siteEnv.AddPoints([2]float64(b.siteCoords[i]))
	}
	if debug {
		b.debugRecord(
			siteEnv,
			DebuggerCategoryBuilder.With("extent"),
			"initial extent of all points.",
		)
	}

	b.subdiv = quadedge.NewQuadEdgeSubdivision(*siteEnv, b.Tolerance)
	if b.recorder.IsValid() {
		b.subdiv.Recorder = b.recorder
	}
	for i := range b.siteCoords {
		if _, err := b.subdiv.InsertSite(b.siteCoords[i]); err != nil {
			if debug {
				b.debugRecord(
					[2]float64(b.siteCoords[i]),
					DebuggerCategoryBuilder.With("failed", "insert", i),
					"failed to insert point %v : %v", i, err,
				)
				log.Println("Returning error:", err)
			}
			return err
		}
	}
	return nil
}

func (b Builder) AddPoints(points []geom.Point, data []interface{}) error {
	pts := make([]geom.Point, len(points)+len(b.siteCoords))
	for i := range b.siteCoords {
		pts[i] = geom.Point(b.siteCoords[i])
	}
	copy(pts[len(b.siteCoords):], points)

	uniquePoints := planar.SortUniquePoints(pts)
	b.siteCoords = b.siteCoords[:0]
	for i := range uniquePoints {
		b.siteCoords = append(b.siteCoords, quadedge.Vertex(uniquePoints[i]))
	}
	return nil
}

func (b Builder) Triangles(withFrame bool) (tris []geom.Triangle, err error) {
	if err = b.initSubdiv(); err != nil {
		return nil, err
	}
	b.subdiv.VisitTriangles(func(triEdges []*quadedge.QuadEdge) {
		var triangle geom.Triangle
		for i := 0; i < 3; i++ {
			v := triEdges[i].Orig()
			triangle[i] = [2]float64(v)
		}
		tris = append(tris, triangle)
	}, withFrame)
	return tris, nil
}
