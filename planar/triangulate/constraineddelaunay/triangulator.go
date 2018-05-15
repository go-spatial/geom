package constraineddelaunay

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar/triangulate"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

/*

TODO:

* Start w/ basic constraint implementation (no intersections)
  // Start at the origin point
  // Search around the origin point for the containing triangle. To determine containment, check two line segments for intersection, if they intersect, it is contained.
  // assert for now if the intersected line segment is a constraint
  // assert for now if we fall on an existing edge
  // Create methods OpposedTriangle and OpposedVertex (?)
  // Loop through finding opposed triangles removing edges, keep a set of all edges that will need to be revisited.

* When constraints are introduced:
  + if there is an intersection between line segments
  	+ Calculate the intersection and divide each line segment to use that point
  	+ The inserted segment should then be inserted again as two different segments, recurse
* Edge conditions:
  + a constrained edge that lies on top of an existing edge
  + a constrained edge that lies on top of another constrained edge

*/

var ErrInvalidPointClassification = errors.New("invalid point classification")
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
	subdiv     *quadedge.QuadEdgeSubdivision
	tolerance float64
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
	if len(arr) == 0 || arr[len(arr) - 1].Equals(v) == false {
		arr = append(arr, v)
	}
	return arr
}

/*
createSegment creates a segment with vertices a & b, if it doesn't already 
exist. All the vertices must already exist in the triangulator.
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

	ct, err := tri.findContainingTriangle(s)
	if err != nil {
		return err
	}
	from := ct.qe.Sym()

	ct, err = tri.findContainingTriangle(triangulate.NewSegment(geom.Line{s.GetEnd(), s.GetStart()}))
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
*/
func (tri *Triangulator) createTriangle(a, b, c quadedge.Vertex) error {
	log.Printf("a: %v b: %v c: %v", a, b, c)
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
*/
func (tri *Triangulator) deleteEdge(e *quadedge.QuadEdge) {

	toRemove := make(map[*quadedge.QuadEdge]bool, 4)

	eSym := e.Sym()
	eRot := e.Rot()
	eRotSym := e.Rot().Sym()

	// a set of all the edges that will be removed.
	toRemove[e] = true
	toRemove[eSym] = true
	toRemove[eRot] = true
	toRemove[eRotSym] = true

	updateVertexIndex := func(v quadedge.Vertex) {
		ve := tri.vertexIndex[v]
		if toRemove[ve] {
			log.Printf("Removing from vertex index: %v", ve)
			for testEdge := ve.ONext(); ; testEdge = testEdge.ONext() {
				if testEdge == ve {
					log.Fatal("unable to update vertex index")
				}
				if toRemove[testEdge] == false {
					log.Printf("Replacing %v with %v", ve, testEdge)
					tri.vertexIndex[v] = testEdge
					break
				}
			}
		}
	}

	// remove this edge from the vertex index.
	updateVertexIndex(e.Orig())
	updateVertexIndex(e.Dest())
	quadedge.Splice(e.OPrev(), e)
	quadedge.Splice(eSym.OPrev(), eSym)

	tri.subdiv.Delete(e)
}

/*
findContainingTriangle finds the triangle that contains the vertex s.GetStart()
and contains at least part of the edge that extends from s.GetStart().

Returns a quadedge that has s.GetStart() as the origin and the right face is 
the desired triangle.
*/
func (tri *Triangulator) findContainingTriangle(s triangulate.Segment) (*Triangle, error) {

	qe, err := tri.locateEdgeByVertex(s.GetStart())
	if err != nil {
		return nil, err
	}

	left := qe

	// walk around all the triangles that share qe.Orig()
	for true {
		if left.IsLive() == false {
			log.Fatalf("unexpected dead node: %v", left)
		}
		// create the two quad edges around s
		right := left.OPrev()

		lc := s.GetEnd().Classify(left.Orig(), left.Dest())
		rc := s.GetEnd().Classify(right.Orig(), right.Dest())
		
		if lc == quadedge.RIGHT && rc == quadedge.LEFT {
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
			return nil, fmt.Errorf("no containing triangle: %v", s)
		}
	}

	return nil, fmt.Errorf("no containing triangle: %v", s)
}

/*
GetEdges gets the edges of the computed triangulation as a MultiLineString.

returns the edges of the triangulation
*/
func (tri *Triangulator) GetEdges() geom.MultiLineString {
	return tri.builder.GetEdges()
}

/*
GetTriangles Gets the faces of the computed triangulation as a
MultiPolygon.
*/
func (tri *Triangulator) GetTriangles() (geom.MultiPolygon, error) {
	return tri.builder.GetTriangles()
}

/*
InsertSegments inserts the line segments in the specified geometry and builds
a triangulation. The line segments are used as constraints in the 
triangulation. If the geometry is made up solely of points, then no 
constraints will be used.
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
	for i := range(edges) {
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

func (tri *Triangulator) insertConstraints(g geom.Geometry) error {
	tri.constraints = make(map[triangulate.Segment]bool)

	lines, err := geom.ExtractLines(g)
	if err != nil {
		return fmt.Errorf("error adding constraint: %v", err)
	}
	for _, l := range(lines) {
		// make the line ordering consistent
		if !cmp.PointLess(l[0], l[1]) {
			l[0], l[1] = l[1], l[0]
		}

		seg := triangulate.NewSegment(l)
		// this maintains the constraints and de-dupes
		tri.constraints[seg] = true
	}

	log.Printf("tri.constraints: %v", tri.constraints)
	for seg := range tri.constraints {
		qe, err := tri.LocateSegment(seg.GetStart(), seg.GetEnd())
		if qe != nil && err != nil {
			return fmt.Errorf("error adding constraint: %v", err)
		}

		if qe == nil {
			err := tri.insertEdgeCDT(&seg)
			if err != nil {
				return fmt.Errorf("error adding constraint: %v", err)
			}
		}
		if err = tri.Validate(); err != nil {
			log.Fatalf("validate failed: %v", err)
		}
	}

	return nil
}

func (tri *Triangulator) IsConstraint(s triangulate.Segment) bool {
	_, ok := tri.constraints[s]
	return ok
}

// Procedure InsertEdgeCDT(T:CDT, ab:Edge)
func (tri *Triangulator) insertEdgeCDT(ab *triangulate.Segment) error {
	log.Printf("ab: %v", *ab)
	// Precondition: a,b in T and ab not in T
	// Find the triangle t ∈ T that contains a and is cut by ab
	at, err := tri.findContainingTriangle(*ab)
	if err != nil {
		return err
	}
	be, err := tri.locateEdgeByVertex(ab.GetEnd())
	if err != nil {
		return err
	}
	log.Printf("be: %v", be)
	log.Printf("at: %v", at)
	t := at

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
		log.Printf("t: %v", t)
		log.Printf("v: %v", v)
		log.Printf("tseq: %v", tseq)
		log.Printf("vseq: %v", vseq)
		shared, err := t.sharedEdge(tseq)
		if err != nil {
			return err
		}
		log.Printf("shared: %v", shared)

		c := vseq.Classify(ab.GetStart(), ab.GetEnd())

		switch c {
		// If vseq above the edge ab then
		case quadedge.LEFT:
			// v:=Vertex shared by t and tseq above ab
			v = shared.Orig()
			pu = appendNonRepeat(pu, v)
			// AddList(PU ,vseq)
			pu = appendNonRepeat(pu, vseq)
		// Else If vseq below the edge ab
		case quadedge.RIGHT:
			// v:=Vertex shared by t and tseq below ab
			v = shared.Dest()
			pl = appendNonRepeat(pl, v)
			// AddList(PL, vseq)
			pl = appendNonRepeat(pl, vseq)
		// NOTE: You may be able to use this same mechanism to handle edges that overlap/intersect
		// Else vseq on the edge ab
		case quadedge.BETWEEN:
			// InsertEdgeCDT(T, vseqb)
			// a:=vseq
			// break
		case quadedge.DESTINATION:
			// nothing left to do
		default:
			log.Printf("c: %v", c)
			return ErrInvalidPointClassification
		}

		// "Remove t from T" -- We are just removing the edge intersected by 
		// ab, which in effect removes the triangle.
		removalList = append(removalList, shared)

		t = tseq
	}
	// EndWhile

	// remove the previously marked edges
	// TODO Inefficient
	for i := range(removalList) {
		tri.deleteEdge(removalList[i])
	}
	if err := tri.Validate(); err != nil {
		log.Fatalf("validate failed: %v", err)
	}

	// TriangulatePseudoPolygon(PU,ab,T)
	log.Printf("pu: %v", pu)
	tri.triangulatePseudoPolygon(pu, *ab)
	// TriangulatePseudoPolygon(PL,ab,T)
	log.Printf("pl: %v", pl)
	tri.triangulatePseudoPolygon(pl, *ab)

	// Reconstitute the triangle adjacencies of T
	// bt, err := tri.findContainingTriangle(triangulate.NewSegment(geom.Line{ab.GetEnd(), ab.GetStart()}))
	// if err != nil {
	// 	return err
	// }

	// // Add edge ab to T
	// log.Printf("at.qe.Sym(): %v", at.qe.Sym())
	// log.Printf("bt.qe.OPrev(): %v", bt.qe.OPrev())
	// quadedge.Connect(at.qe.Sym(), bt.qe.OPrev())
	tri.createSegment(*ab)

	return nil
}

/*
locateEdgeByVertex finds a quad edge that has this vertex as Orig(). This will 
not be a unique edge.

This is looking for an exact match and tolerance will not be considered.
*/
func (tri *Triangulator) locateEdgeByVertex(v quadedge.Vertex) (*quadedge.QuadEdge, error) {
	qe := tri.vertexIndex[v]

	if qe == nil {
		return nil, quadedge.ErrLocateFailure
	}
	return qe, nil
}

/*
locateEdgeByVertex finds a quad edge that has this vertex as Orig(). This will 
not be a unique edge.

This is looking for an exact match and tolerance will not be considered.
*/
func (tri *Triangulator) LocateSegment(v1 quadedge.Vertex, v2 quadedge.Vertex) (*quadedge.QuadEdge, error) {
	qe := tri.vertexIndex[v1]

	if qe == nil {
		return nil, quadedge.ErrLocateFailure
	}
	if err := tri.Validate(); err != nil {
		log.Fatalf("validate failed: %v", err)
	}

	start := qe
	for true {
		if qe == nil || qe.IsLive() == false {
			log.Fatalf("unexpected dead node: %v", qe)
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


// TriangulatePseudoPolygon
// Pseudocode taken from Figure 10
// http://old.cescg.org/CESCG-2004/web/Domiter-Vid/CDT.pdf
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
*/
func (tri *Triangulator) Validate() error {
	err := tri.subdiv.Validate()
	if err != nil {
		return err
	}
	return tri.validateVertexIndex()
}

/*
validateVertexIndex self consistency checks against a triangulation and the 
subdiv and reports the first error.
*/
func (tri *Triangulator) validateVertexIndex() error {
	// collect a set of all edges
	edgeSet := make(map[*quadedge.QuadEdge]bool)
	vertexSet := make(map[quadedge.Vertex]bool)
	edges := tri.subdiv.GetEdges()
	for i := range(edges) {
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
