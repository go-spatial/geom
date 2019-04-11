package quadedge

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-spatial/geom"
)

var ErrUnexpectedDeadNode = errors.New("unexpected dead node")
var ErrCoincidentEdges = errors.New("coincident edges")

func (qes *QuadEdgeSubdivision) InsertSite(v Vertex) (*QuadEdge, error) {

	if debug {
		defer qes.debugAugementRecorder().Close()
		/*
		for i, ls := range qes.GetEdgesAsMultiLineString() {
			qes.debugRecord(ls,
				DebuggerCategoryQES.With(i, "edge"),
				"Initial:%v", i,
			)
		}
		qes.debugRecord([2]float64(v),
			DebuggerCategoryQES.With("vertex"),
			"inserting: %v", v,
		)
		*/
	}
	/*
		This code is based on Guibas and Stolfi (1985), with minor modifications
		and a bug fix from Dani Lischinski (Graphic Gems 1993). (The modification
		I believe is the test for the inserted site falling exactly on an
		existing edge. Without this test zero-width triangles have been observed
		to be created)
	*/
	e, err := qes.Locate(v)
	if err != nil {
		return nil, err
	}

	if qes.IsVertexOfEdge(e, v) {
		// point is already in subdivision.
		return e, nil
	}
	if qes.IsOnEdge(e, v) {
		// the point lies exactly on an edge, so delete the edge
		// (it will be replaced by a pair of edges which have the point as a vertex)
		e = e.OPrev()
		qes.Delete(e.ONext())
	}

	/*
		Connect the new point to the vertices of the containing triangle
		(or quadrilateral, if the new point fell on an existing edge.)
	*/
	base := MakeEdge(e.Orig(), v)
	Splice(base, e)
	startEdge := base

	for {
		base = qes.Connect(e, base.Sym())
		e = base.OPrev()
		if e.LNext() == startEdge {
			break
		}
	}

	// Examine suspect edges to ensure that the Delaunay condition
	// is satisfied.
	for {
		t := e.OPrev()
		switch {
		case t.Dest().RightOf(*e) && v.IsInCircle(e.Orig(), t.Dest(), e.Dest()):
			Swap(e)
			e = e.OPrev()

		case e.ONext() == startEdge:
			/*
			if debug {
				for i, ls := range qes.GetEdgesAsMultiLineString() {
					qes.debugRecord(ls,
						DebuggerCategoryQES.With("new_state", i, "edge"),
						"New State:%v", i,
					)
				}
			}
			*/
			return base, nil // no more suspect edges.

		}
		e = e.ONext().LPrev()
	}
}

// LocateSegment will find the next segment defined by start.Vertex, v2 starting at start.
func LocateSegment(start *QuadEdge, v2 Vertex) (*QuadEdge, error) {
	if start == nil {
		return nil, ErrLocateFailure{QE: start}
	}
	v1 := start.vertex
	qe := start
	for {
		if qe.IsLive() == false {
			if debug {
				log.Printf("unexpected dead node: %v", qe)
			}
			return nil, fmt.Errorf("nil or dead qe when locationg segment %v %v", v1, v2)
		}
		if v2.Equals(qe.Dest()) {
			return qe, nil
		}
		qe = qe.ONext()
		// got back could not find segment
		if qe == start {
			return nil, nil
		}
	}
}

func FindIntersectingTriangle(start *QuadEdge, end Vertex) (Triangle, error) {
	startVertex := start.Orig()
	left := start
	right := left.OPrev()
	// walk around all the triangles that share start.Orig()
	for {
		if !left.IsLive() {
			return Triangle{}, ErrUnexpectedDeadNode
		}

		// create the two quad edges around the segment
		right = left.OPrev()

		lc := end.Classify(left.Orig(), left.Dest())
		rc := end.Classify(right.Orig(), right.Dest())

		if (lc == RIGHT && rc == LEFT) ||
			lc == BETWEEN ||
			lc == DESTINATION ||
			lc == BEYOND {
			return Triangle{left}, nil
		}

		if lc != RIGHT && lc != LEFT &&
			rc != RIGHT && rc != LEFT {
			return Triangle{left}, ErrCoincidentEdges
		}
		left = right
		if left == start {
			// we have walked all the around the vertex.
			break
		}
	}
	return Triangle{}, fmt.Errorf("no intersecting triangle: %v - %v", startVertex, end)
}

// SegmentExists will attempt to find a segment off the given edge with the dest vertex equal to dest. If it finds such an edge it will return true; false otherwise.
func SegmentExists(edge *QuadEdge, dest Vertex) bool {
	if edge == nil {
		return false
	}
	edge1, err := LocateSegment(edge, dest)
	if _, ok := err.(ErrLocateFailure); err != nil && !ok {
		return false
	}
	// Does the edge exists
	return edge1 != nil
}

func (qes *QuadEdgeSubdivision) InsertEdge(Index map[Vertex]*QuadEdge, orig, dest Vertex) error {

	var ct Triangle
	var err error
	edge := Index[orig]

	// Edge already exists, skip.
	if SegmentExists(edge, dest) {
		return nil
	}

	if ct, err = FindIntersectingTriangle(edge, dest); err != nil && err != ErrCoincidentEdges {
		return err
	}
	from := ct.QE.Sym()

	symEdge := Index[dest]

	if ct, err = FindIntersectingTriangle(symEdge, orig); err != nil && err != ErrCoincidentEdges {
		return err
	}

	to := ct.QE.OPrev()
	_ = Connect(from, to)
	if debug {
		return qes.Validate()
	}
	return nil
}

func (qes *QuadEdgeSubdivision) VertexIndex() map[Vertex]*QuadEdge {
	edges := qes.GetEdges()
	vertexIndex := make(map[Vertex]*QuadEdge, len(edges)*2)
	for i := range edges {
		if _, ok := vertexIndex[edges[i].Orig()]; !ok {
			vertexIndex[edges[i].Orig()] = edges[i]
		}
		if _, ok := vertexIndex[edges[i].Dest()]; !ok {
			vertexIndex[edges[i].Dest()] = edges[i].Sym()
		}
	}
	return vertexIndex
}

// EdgeAsGeomLine returns the geom.Line define by edge.
func EdgeAsGeomLine(edge *QuadEdge) geom.Line {
	return geom.Line{geom.Point(edge.Orig()), geom.Point(edge.Dest())}
}

// VertexAsGeomLine returns the geom.Line defined from p1 to p2.
func VertexAsGeomLine(p1, p2 Vertex) geom.Line { return geom.Line{geom.Point(p1), geom.Point(p2)} }
