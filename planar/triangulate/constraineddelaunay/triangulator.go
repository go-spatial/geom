package constraineddelaunay

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar/triangulate"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

var ErrInvalidPointClassification = errors.New("invalid point classification")
var ErrLinesDoNotIntersect = errors.New("line segments do not intersect")

// these errors indicate a problem with the algorithm.
var ErrUnableToUpdateVertexIndex = errors.New("unable to update vertex index")
var ErrUnexpectedDeadNode = errors.New("unexpected dead node")
var ErrUnsupportedCoincidentEdges = errors.New("unsupported coincident edges")

/*
Triangulator provides methods for performing a constrainted delaunay
triangulation.

Domiter, Vid. "Constrained Delaunay triangulation using plane subdivision."
Proceedings of the 8th central European seminar on computer graphics.
Budmerice. 2004.
http://old.cescg.org/CESCG-2004/web/Domiter-Vid/CDT.pdf
*/
type Triangulator struct {
	builder *triangulate.DelaunayTriangulationBuilder
	// a map of constraints where the segments have the lesser point first.
	constraints map[triangulate.Segment]bool
	subdiv      *quadedge.QuadEdgeSubdivision
	tolerance   float64
	// run validation after many modification operations. This is expensive,
	// but very useful when debugging.
	validate bool
	// maintain an index of vertices to quad edges. Each vertex will point to
	// one quad edge that has the vertex as an origin. The other quad edges
	// that point to this vertex can be reached from there.
	vertexIndex map[quadedge.Vertex]*quadedge.QuadEdge
}

/*
appendNonRepeat only appends the provided value if it does not repeat the last
value that was appended onto the array.
*/
func appendNonRepeat(arr []quadedge.Vertex, v quadedge.Vertex) []quadedge.Vertex {
	if len(arr) == 0 || arr[len(arr)-1].Equals(v) == false {
		arr = append(arr, v)
	}
	return arr
}

/*
createSegment creates a segment with vertices a & b, if it doesn't already
exist. All the vertices must already exist in the triangulator.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) createSegment(s triangulate.Segment) error {
	qe, err := tri.LocateSegment(s.GetStart(), s.GetEnd())
	if err != nil && err != quadedge.ErrLocateFailure {
		return err
	}
	if qe != nil {
		// if the segment already exists
		return nil
	}

	ct, err := tri.findIntersectingTriangle(s)
	if err != nil {
		return err
	}
	from := ct.qe.Sym()

	ct, err = tri.findIntersectingTriangle(triangulate.NewSegment(geom.Line{s.GetEnd(), s.GetStart()}))
	if err != nil {
		return err
	}
	to := ct.qe.OPrev()

	quadedge.Connect(from, to)
	// since we aren't adding any vertices we don't need to modify the vertex
	// index.
	return nil
}

/*
createTriangle creates a triangle with vertices a, b and c. All the vertices
must already exist in the triangulator. Any existing edges that make up the triangle will not be recreated.

This method makes no effort to ensure the resulting changes are a valid
triangulation.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) createTriangle(a, b, c quadedge.Vertex) error {
	if err := tri.createSegment(triangulate.NewSegment(geom.Line{a, b})); err != nil {
		return err
	}

	if err := tri.createSegment(triangulate.NewSegment(geom.Line{b, c})); err != nil {
		return err
	}

	if err := tri.createSegment(triangulate.NewSegment(geom.Line{c, a})); err != nil {
		return err
	}

	return nil
}

/*
deleteEdge deletes the specified edge and updates all associated neighbors to
reflect the removal. The local vertex index is also updated to reflect the
deletion.

It is invalid to call this method on the last edge that links to a vertex.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) deleteEdge(e *quadedge.QuadEdge) error {

	toRemove := make(map[*quadedge.QuadEdge]bool, 4)

	eSym := e.Sym()
	eRot := e.Rot()
	eRotSym := e.Rot().Sym()

	// a set of all the edges that will be removed.
	toRemove[e] = true
	toRemove[eSym] = true
	toRemove[eRot] = true
	toRemove[eRotSym] = true

	// remove this edge from the vertex index.
	if err := tri.removeEdgesFromVertexIndex(toRemove, e.Orig()); err != nil {
		return err
	}
	if err := tri.removeEdgesFromVertexIndex(toRemove, e.Dest()); err != nil {
		return err
	}
	quadedge.Splice(e.OPrev(), e)
	quadedge.Splice(eSym.OPrev(), eSym)

	// TODO: this call is horribly inefficient and should be optimized.
	tri.subdiv.Delete(e)

	return nil
}

/*
findIntersectingTriangle finds the triangle that shares the vertex s.GetStart()
and intersects at least part of the edge that extends from s.GetStart().

Tolerance is not considered when determining if vertices are the same.

Returns a quadedge that has s.GetStart() as the origin and the right face is
the desired triangle. If the segment falls on an edge, the triangle to the
right of the segment is returned.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) findIntersectingTriangle(s triangulate.Segment) (*Triangle, error) {

	qe, err := tri.locateEdgeByVertex(s.GetStart())
	if err != nil {
		return nil, err
	}

	left := qe

	// walk around all the triangles that share qe.Orig()
	for true {
		if left.IsLive() == false {
			return nil, ErrUnexpectedDeadNode
		}
		// create the two quad edges around s
		right := left.OPrev()

		lc := s.GetEnd().Classify(left.Orig(), left.Dest())
		rc := s.GetEnd().Classify(right.Orig(), right.Dest())

		if (lc == quadedge.RIGHT && rc == quadedge.LEFT) || lc == quadedge.BETWEEN || lc == quadedge.DESTINATION || lc == quadedge.BEYOND {
			// if s is between the two edges, we found our triangle.
			return &Triangle{left}, nil
		} else if lc != quadedge.RIGHT && lc != quadedge.LEFT && rc != quadedge.LEFT && rc != quadedge.RIGHT {
			// if s falls on lc or rc, then throw an error (for now)
			// TODO: Handle this case
			return nil, ErrUnsupportedCoincidentEdges
		}
		left = right

		if left == qe {
			// if we've walked all the way around the vertex.
			return nil, fmt.Errorf("no intersecting triangle: %v", s)
		}
	}

	return nil, fmt.Errorf("no intersecting triangle: %v", s)
}

/*
GetEdges gets the edges of the computed triangulation as a MultiLineString.

returns the edges of the triangulation

If tri is nil a panic will occur.
*/
func (tri *Triangulator) GetEdges() geom.MultiLineString {
	return tri.builder.GetEdges()
}

/*
GetTriangles Gets the faces of the computed triangulation as a
MultiPolygon.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) GetTriangles() (geom.MultiPolygon, error) {
	return tri.builder.GetTriangles()
}

/*
InsertSegments inserts the line segments in the specified geometry and builds
a triangulation. The line segments are used as constraints in the
triangulation. If the geometry is made up solely of points, then no
constraints will be used.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) InsertSegments(g geom.Geometry) error {
	err := tri.insertSites(g)
	if err != nil {
		return err
	}

	err = tri.insertConstraints(g)
	if err != nil {
		return err
	}

	return nil
}

/*
insertSites inserts all of the vertices found in g into a Delaunay
triangulation. Other steps will modify the Delaunay Triangulation to create
the constrained Delaunay triangulation.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) insertSites(g geom.Geometry) error {
	tri.builder = triangulate.NewDelaunayTriangulationBuilder(tri.tolerance)
	err := tri.builder.SetSites(g)
	if err != nil {
		return err
	}
	tri.subdiv = tri.builder.GetSubdivision()

	// Add all the edges to a constant time lookup
	tri.vertexIndex = make(map[quadedge.Vertex]*quadedge.QuadEdge)
	edges := tri.subdiv.GetEdges()
	for i := range edges {
		e := edges[i]
		if _, ok := tri.vertexIndex[e.Orig()]; ok == false {
			tri.vertexIndex[e.Orig()] = e
		}
		if _, ok := tri.vertexIndex[e.Dest()]; ok == false {
			tri.vertexIndex[e.Dest()] = e.Sym()
		}
	}

	return nil
}

/*
insertConstraints modifies the triangulation by incrementally using the
line segements in g as constraints in the triangulation. After this step
the triangulation is no longer a proper Delaunay triangulation, but the
constraints are guaranteed. Some constraints may need to be split (think
about the case when two constraints intersect).

If tri is nil a panic will occur.
*/
func (tri *Triangulator) insertConstraints(g geom.Geometry) error {
	tri.constraints = make(map[triangulate.Segment]bool)

	lines, err := geom.ExtractLines(g)
	if err != nil {
		return fmt.Errorf("error adding constraint: %v", err)
	}
	constraints := make(map[triangulate.Segment]bool)
	for _, l := range lines {
		// make the line ordering consistent
		if !cmp.PointLess(l[0], l[1]) {
			l[0], l[1] = l[1], l[0]
		}

		seg := triangulate.NewSegment(l)
		// this maintains the constraints and de-dupes
		constraints[seg] = true
		tri.constraints[seg] = true
	}

	for seg := range constraints {
		if err := tri.insertEdgeCDT(&seg); err != nil {
			return fmt.Errorf("error adding constraint: %v", err)
		}
		if err = tri.Validate(); err != nil {
			return err
		}
	}

	return nil
}

/*
intersection calculates the intersection between two line segments. When the
rest of geom is ported over from spatial, this can be replaced with a more
generic call.

The tolerance here only acts by extending the lines by tolerance. E.g. if the
tolerance is 0.1 and you have two lines {{0, 0}, {1, 0}} and
{{0, 0.01}, {1, 0.01}} then these will not be marked as intersecting lines.

If tolerance is used to mark two lines as intersecting, you are still
guaranteed that the intersecting point will fall _on_ one of the lines, not in
the extended region of the line.

Taken from: https://stackoverflow.com/questions/563198/how-do-you-detect-where-two-line-segments-intersect

If tri is nil a panic will occur.
*/
func (tri *Triangulator) intersection(l1, l2 triangulate.Segment) (quadedge.Vertex, error) {
	p := l1.GetStart()
	r := l1.GetEnd().Sub(p)
	q := l2.GetStart()
	s := l2.GetEnd().Sub(q)

	rs := r.CrossProduct(s)

	if rs == 0 {
		return quadedge.Vertex{}, ErrLinesDoNotIntersect
	}
	t := q.Sub(p).CrossProduct(s.Divide(r.CrossProduct(s)))
	u := p.Sub(q).CrossProduct(r.Divide(s.CrossProduct(r)))

	// calculate the acceptable range of values for t
	ttolerance := tri.tolerance / r.Magn()
	tlow := -ttolerance
	thigh := 1 + ttolerance

	// calculate the acceptable range of values for u
	utolerance := tri.tolerance / s.Magn()
	ulow := -utolerance
	uhigh := 1 + utolerance

	if t < tlow || t > thigh || u < ulow || u > uhigh {
		return quadedge.Vertex{}, ErrLinesDoNotIntersect
	}
	// if t is just out of range, but within the acceptable tolerance, snap
	// it back to the beginning/end of the line.
	t = math.Min(1, math.Max(t, 0))

	return p.Sum(r.Times(t)), nil
}

/*
IsConstraint returns true if e is a constrained edge.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) IsConstraint(e *quadedge.QuadEdge) bool {

	_, ok := tri.constraints[triangulate.NewSegment(geom.Line{e.Orig(), e.Dest()})]
	if ok {
		return true
	}
	_, ok = tri.constraints[triangulate.NewSegment(geom.Line{e.Dest(), e.Orig()})]
	return ok
}

/*
insertEdgeCDT attempts to follow the pseudo code in Domiter.

Procedure InsertEdgeCDT(T:CDT, ab:Edge)

There are some deviations that are also mentioned inline in the comments

 - Some aparrent typos that are resolved to give consistent results
 - Modifications to work with the planar subdivision representation of
   a triangulation (QuadEdge)
 - Modification to support the case when two constrained edges intersect
   at more than the end points.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) insertEdgeCDT(ab *triangulate.Segment) error {

	qe, err := tri.LocateSegment(ab.GetStart(), ab.GetEnd())
	if qe != nil && err != nil {
		return fmt.Errorf("error inserting constraint: %v", err)
	}
	if qe != nil {
		// nothing to do, the edge already exists.
		return nil
	}

	// Precondition: a,b in T and ab not in T
	// Find the triangle t ∈ T that contains a and is cut by ab
	t, err := tri.findIntersectingTriangle(*ab)
	if err != nil {
		return err
	}

	removalList := make([]*quadedge.QuadEdge, 0)

	// PU:=EmptyList
	pu := make([]quadedge.Vertex, 0)
	// PL:=EmptyList
	pl := make([]quadedge.Vertex, 0)
	// v:=a
	v := ab.GetStart()
	b := ab.GetEnd()

	// While v not in t do -- should this be 'b not in t'!? -JRS
	for t.IntersectsPoint(b) == false {
		// tseq:=OpposedTriangle(t,v)
		tseq, err := t.opposedTriangle(v)
		if err != nil {
			return err
		}
		// vseq:=OpposesdVertex(tseq,t)
		vseq, err := tseq.opposedVertex(t)
		if err != nil {
			return err
		}

		shared, err := t.sharedEdge(tseq)
		if err != nil {
			return err
		}

		c := vseq.Classify(ab.GetStart(), ab.GetEnd())

		// should we remove the edge shared between t & tseq?
		flagEdgeForRemoval := false

		switch {

		case tri.subdiv.IsOnLine(ab.GetLineSegment(), shared.Orig()):
			// InsertEdgeCDT(T, vseqb)
			vb := triangulate.NewSegment(geom.Line{shared.Orig(), ab.GetEnd()})
			tri.insertEdgeCDT(&vb)
			// a:=vseq -- Should this be b:=vseq!? -JRS
			b = shared.Orig()
			*ab = triangulate.NewSegment(geom.Line{ab.GetStart(), b})

		case tri.subdiv.IsOnLine(ab.GetLineSegment(), shared.Dest()):
			// InsertEdgeCDT(T, vseqb)
			vb := triangulate.NewSegment(geom.Line{shared.Dest(), ab.GetEnd()})
			tri.insertEdgeCDT(&vb)
			// a:=vseq -- Should this be b:=vseq!? -JRS
			b = shared.Dest()
			*ab = triangulate.NewSegment(geom.Line{ab.GetStart(), b})

		// if the constrained edge is passing through another constrained edge
		case tri.IsConstraint(shared):
			// find the point of intersection
			iv, err := tri.intersection(*ab, triangulate.NewSegment(geom.Line{shared.Orig(), shared.Dest()}))
			if err != nil {
				return err
			}

			// split the constrained edge we interesect
			if err := tri.splitEdge(shared, iv); err != nil {
				return err
			}
			tri.deleteEdge(shared)
			tseq, err = t.opposedTriangle(v)
			if err != nil {
				return err
			}

			// create a new edge for the rest of this segment and recursively
			// insert the new edge.
			vb := triangulate.NewSegment(geom.Line{iv, ab.GetEnd()})
			tri.insertEdgeCDT(&vb)

			// the current insertion will stop at the interesction point
			b = iv
			*ab = triangulate.NewSegment(geom.Line{ab.GetStart(), iv})

		// If vseq above the edge ab then
		case c == quadedge.LEFT:
			// v:=Vertex shared by t and tseq above ab
			v = shared.Orig()
			pu = appendNonRepeat(pu, v)
			// AddList(PU ,vseq)
			pu = appendNonRepeat(pu, vseq)
			flagEdgeForRemoval = true

		// Else If vseq below the edge ab
		case c == quadedge.RIGHT:
			// v:=Vertex shared by t and tseq below ab
			v = shared.Dest()
			pl = appendNonRepeat(pl, v)
			// AddList(PL, vseq)
			pl = appendNonRepeat(pl, vseq)
			flagEdgeForRemoval = true

		case c == quadedge.DESTINATION:
			flagEdgeForRemoval = true

		default:
			return ErrInvalidPointClassification
		}

		if flagEdgeForRemoval {
			// "Remove t from T" -- We are just removing the edge intersected
			// by ab, which in effect removes the triangle.
			removalList = append(removalList, shared)
		}

		t = tseq
	}
	// EndWhile

	// remove the previously marked edges
	for i := range removalList {
		tri.deleteEdge(removalList[i])
	}

	// TriangulatePseudoPolygon(PU,ab,T)
	if err := tri.triangulatePseudoPolygon(pu, *ab); err != nil {
		return err
	}
	// TriangulatePseudoPolygon(PL,ab,T)
	if err := tri.triangulatePseudoPolygon(pl, *ab); err != nil {
		return err
	}

	if err := tri.Validate(); err != nil {
		return err
	}

	// Add edge ab to T
	if err := tri.createSegment(*ab); err != nil {
		return err
	}

	return nil
}

/*
locateEdgeByVertex finds a quad edge that has this vertex as Orig(). This will
not be a unique edge.

This is looking for an exact match and tolerance will not be considered.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) locateEdgeByVertex(v quadedge.Vertex) (*quadedge.QuadEdge, error) {
	qe := tri.vertexIndex[v]

	if qe == nil {
		return nil, quadedge.ErrLocateFailure
	}
	return qe, nil
}

/*
LocateSegment finds a quad edge that has this vertex as Orig(). This will
not be a unique edge.

This is looking for an exact match and tolerance will not be considered.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) LocateSegment(v1 quadedge.Vertex, v2 quadedge.Vertex) (*quadedge.QuadEdge, error) {
	qe := tri.vertexIndex[v1]

	if qe == nil {
		return nil, quadedge.ErrLocateFailure
	}

	start := qe
	for true {
		if qe == nil || qe.IsLive() == false {
			log.Printf("unexpected dead node: %v", qe)
			return nil, fmt.Errorf("nil or dead qe when locating segment %v %v", v1, v2)
		}
		if v2.Equals(qe.Dest()) {
			return qe, nil
		}

		qe = qe.ONext()
		if qe == start {
			return nil, quadedge.ErrLocateFailure
		}
	}

	return qe, nil
}

/*
removeConstraintEdge removes any constraints that share the same Orig() and
Dest() as the edge provided. If there are none, no changes are made.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) removeConstraintEdge(e *quadedge.QuadEdge) {
	delete(tri.constraints, triangulate.NewSegment(geom.Line{e.Orig(), e.Dest()}))
	delete(tri.constraints, triangulate.NewSegment(geom.Line{e.Dest(), e.Orig()}))
}

/*
removeEdgesFromVertexIndex will remove a set of QuadEdges from the vertex index
for the specified vertex. If the operation cannot be completed an error will be
returned and the index will not be modified.

The vertex index maps from a vertex to an arbitrary QuadEdges. This method is
helpful in modifying the index after an edge has been deleted.

toRemove - a set of QuadEdges that should be removed from the index. These
QuadEdges don't necessarily have to link to the provided vertex.
v - The vertex to modify in the index.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) removeEdgesFromVertexIndex(toRemove map[*quadedge.QuadEdge]bool, v quadedge.Vertex) error {
	ve := tri.vertexIndex[v]
	if toRemove[ve] {
		for testEdge := ve.ONext(); ; testEdge = testEdge.ONext() {
			if testEdge == ve {
				// if we made it all the way around the vertex without finding
				// a valid edge to reference from this vertex
				return ErrUnableToUpdateVertexIndex
			}
			if toRemove[testEdge] == false {
				tri.vertexIndex[v] = testEdge
				return nil
			}
		}
	}
	// this should happen if the vertex doesn't need to be updated.
	return nil
}

/*
splitEdge splits the given edge at the vertex v.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) splitEdge(e *quadedge.QuadEdge, v quadedge.Vertex) error {
	constraint := tri.IsConstraint(e)

	ePrev := e.OPrev()
	eSym := e.Sym()
	eSymPrev := eSym.OPrev()

	tri.removeConstraintEdge(e)

	e1 := tri.subdiv.MakeEdge(e.Orig(), v)
	e2 := tri.subdiv.MakeEdge(e.Dest(), v)

	if _, ok := tri.vertexIndex[v]; ok == false {
		tri.vertexIndex[v] = e1.Sym()
	}

	// splice e1 on
	quadedge.Splice(ePrev, e1)
	// splice e2 on
	quadedge.Splice(eSymPrev, e2)

	// splice e1 and e2 together
	quadedge.Splice(e1.Sym(), e2.Sym())

	if constraint {
		tri.constraints[triangulate.NewSegment(geom.Line{e1.Orig(), e1.Dest()})] = true
		tri.constraints[triangulate.NewSegment(geom.Line{e2.Dest(), e2.Orig()})] = true
	}

	// since we aren't adding any vertices we don't need to modify the vertex
	// index.
	return nil
}

/*
triangulatePseudoPolygon is taken from the pseudocode TriangulatePseudoPolygon
from Figure 10 in Domiter.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) triangulatePseudoPolygon(p []quadedge.Vertex, ab triangulate.Segment) error {
	a := ab.GetStart()
	b := ab.GetEnd()
	var c quadedge.Vertex
	// If P has more than one element then
	if len(p) > 1 {
		// c:=First vertex of P
		c = p[0]
		ci := 0
		// For each vertex v in P do
		for i, v := range p {
			// If v ∈ CircumCircle (a, b, c) then
			if quadedge.TrianglePredicate.IsInCircleRobust(a, b, c, v) {
				c = v
				ci = i
			}
		}
		// Divide P into PE and PD giving P=PE+c+PD
		pe := p[0:ci]
		pd := p[ci+1:]
		// TriangulatePseudoPolygon(PE, ac, T)
		if err := tri.triangulatePseudoPolygon(pe, triangulate.NewSegment(geom.Line{a, c})); err != nil {
			return err
		}
		// TriangulatePseudoPolygon(PD, cd, T) (cb instead of cd? -JRS)
		if err := tri.triangulatePseudoPolygon(pd, triangulate.NewSegment(geom.Line{c, b})); err != nil {
			return err
		}
	} else if len(p) == 1 {
		c = p[0]
	}

	// If P is not empty then
	if len(p) > 0 {
		// Add triangle with vertices a, b, c into T
		if err := tri.createTriangle(a, c, b); err != nil {
			return err
		}
	}

	return nil
}

/*
validate runs a number of self consistency checks against a triangulation and
reports the first error.

This is most useful when testing/debugging.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) Validate() error {
	if tri.validate == false {
		return nil
	}
	err := tri.subdiv.Validate()
	if err != nil {
		return err
	}
	return tri.validateVertexIndex()
}

/*
validateVertexIndex self consistency checks against a triangulation and the
subdiv and reports the first error.

If tri is nil a panic will occur.
*/
func (tri *Triangulator) validateVertexIndex() error {
	// collect a set of all edges
	edgeSet := make(map[*quadedge.QuadEdge]bool)
	vertexSet := make(map[quadedge.Vertex]bool)
	edges := tri.subdiv.GetEdges()
	for i := range edges {
		edgeSet[edges[i]] = true
		edgeSet[edges[i].Sym()] = true
		vertexSet[edges[i].Orig()] = true
		vertexSet[edges[i].Dest()] = true
	}

	// verify the vertex index points to appropriate edges and vertices
	for v, e := range tri.vertexIndex {
		if _, ok := vertexSet[v]; ok == false {
			return fmt.Errorf("vertex index contains an unexpected vertex: %v", v)
		}
		if _, ok := edgeSet[e]; ok == false {
			return fmt.Errorf("vertex index contains an unexpected edge: %v", e)
		}
		if v.Equals(e.Orig()) == false {
			return fmt.Errorf("vertex index points to an incorrect edge, expected %v got %v", e.Orig(), v)
		}
	}

	// verify all vertices are in the vertex index
	for v, _ := range vertexSet {
		if _, ok := tri.vertexIndex[v]; ok == false {
			return fmt.Errorf("vertex index is missing a vertex: %v", v)
		}
	}

	return nil
}
