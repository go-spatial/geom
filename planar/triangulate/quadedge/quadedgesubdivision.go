/*
Copyright (c) 2016 Vivid Solutions.

All rights reserved. This program and the accompanying materials
are made available under the terms of the Eclipse Public License v1.0
and Eclipse Distribution License v. 1.0 which accompanies this distribution.
The Eclipse Public License is available at http://www.eclipse.org/legal/epl-v10.html
and the Eclipse Distribution License is available at

http://www.eclipse.org/org/documents/edl-v10.php.
*/

package quadedge

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
)

var ErrLocateFailure = errors.New("failure locating edge")

/*
A class that contains the QuadEdges representing a planar subdivision that
models a triangulation. The subdivision is constructed using the quadedge
algebra defined in the class QuadEdge. All metric calculations are done in the
Vertex class. In addition to a triangulation, subdivisions support extraction
of Voronoi diagrams. This is easily accomplished, since the Voronoi diagram is
the dual of the Delaunay triangulation.

Subdivisions can be provided with a tolerance value. Inserted vertices which
are closer than this value to vertices already in the subdivision will be
ignored. Using a suitable tolerance value can prevent robustness failures
from happening during Delaunay triangulation.

Subdivisions maintain a frame triangle around the client-created
edges. The frame is used to provide a bounded "container" for all edges
within a TIN. Normally the frame edges, frame connecting edges, and frame
triangles are not included in client processing.

Author David Skea
Author Martin Davis
Ported to Go by Jason R. Surratt
*/
type QuadEdgeSubdivision struct {

	// used for edge extraction to ensure edge uniqueness
	visitedKey               int
	quadEdges                []*QuadEdge
	startingEdge             *QuadEdge
	tolerance                float64
	edgeCoincidenceTolerance float64
	frameVertex              [3]Vertex
	frameEnv                 geom.Extent
	locator                  QuadEdgeLocator
}

var EDGE_COINCIDENCE_TOL_FACTOR float64 = 1000

// 	/**
// 	 * Gets the edges for the triangle to the left of the given {@link QuadEdge}.
// 	 *
// 	 * @param startQE
// 	 * @param triEdge
// 	 *
// 	 * @throws IllegalArgumentException
// 	 *           if the edges do not form a triangle
// 	 */
// 	public static void getTriangleEdges(QuadEdge startQE, QuadEdge[] triEdge) {
// 		triEdge[0] = startQE;
// 		triEdge[1] = triEdge[0].lNext();
// 		triEdge[2] = triEdge[1].lNext();
// 		if (triEdge[2].lNext() != triEdge[0])
// 			throw new IllegalArgumentException("Edges do not form a triangle");
// 	}

/*
Creates a new instance of a quad-edge subdivision based on a frame triangle
that encloses a supplied bounding box. A new super-bounding box that
contains the triangle is computed and stored.

env - the bounding box to surround
tolerance - the tolerance value for determining if two sites are equal
*/
func NewQuadEdgeSubdivision(env geom.Extent, tolerance float64) *QuadEdgeSubdivision {
	var qes QuadEdgeSubdivision
	qes.tolerance = tolerance
	qes.edgeCoincidenceTolerance = tolerance / EDGE_COINCIDENCE_TOL_FACTOR

	qes.createFrame(env)
	qes.startingEdge = qes.initSubdiv()
	qes.locator = NewLastFoundQuadEdgeLocator(&qes)
	return &qes
}

func (qes *QuadEdgeSubdivision) createFrame(env geom.Extent) {
	deltaX := env.XSpan()
	deltaY := env.YSpan()
	offset := 0.0
	if deltaX > deltaY {
		offset = deltaX * 10.0
	} else {
		offset = deltaY * 10.0
	}

	qes.frameVertex[0] = Vertex{(env.MaxX() + env.MinX()) / 2.0, env.MaxY() + offset}
	qes.frameVertex[1] = Vertex{env.MinX() - offset, env.MinY() - offset}
	qes.frameVertex[2] = Vertex{env.MaxX() + offset, env.MinY() - offset}

	qes.frameEnv = *geom.NewExtent(qes.frameVertex[0], qes.frameVertex[1], qes.frameVertex[2])
}

func (qes *QuadEdgeSubdivision) initSubdiv() *QuadEdge {
	// build initial subdivision from frame
	ea := qes.MakeEdge(qes.frameVertex[0], qes.frameVertex[1])
	eb := qes.MakeEdge(qes.frameVertex[1], qes.frameVertex[2])
	Splice(ea.Sym(), eb)
	ec := qes.MakeEdge(qes.frameVertex[2], qes.frameVertex[0])
	Splice(eb.Sym(), ec)
	Splice(ec.Sym(), ea)
	return ea
}

/*
Gets the vertex-equality tolerance value
used in this subdivision

return the tolerance value
*/
func (qes *QuadEdgeSubdivision) GetTolerance() float64 {
	return qes.tolerance
}

/*
Gets the envelope of the Subdivision (including the frame).

@return the envelope
*/
func (qes *QuadEdgeSubdivision) GetEnvelope() geom.Extent {
	// returns a deep copy to avoid modification by caller
	return qes.frameEnv.Extent()
}

/*
GetEdges gets the collection of base {@link QuadEdge}s (one for every pair of
vertices which is connected).

return a collection of QuadEdges
*/
func (qes *QuadEdgeSubdivision) GetEdges() []*QuadEdge {
	return qes.quadEdges
}

// 	/**
// 	 * Sets the {@link QuadEdgeLocator} to use for locating containing triangles
// 	 * in this subdivision.
// 	 *
// 	 * @param locator
// 	 *          a QuadEdgeLocator
// 	 */
// 	public void setLocator(QuadEdgeLocator locator) {
// 		this.locator = locator;
// 	}

/*
MakeEdge creates a new quadedge, recording it in the edges list.

return a new quadedge
*/
func (qes *QuadEdgeSubdivision) MakeEdge(o Vertex, d Vertex) *QuadEdge {
	q := MakeEdge(o, d)
	qes.quadEdges = append(qes.quadEdges, q)
	return q
}

/*
Connect creates a new QuadEdge connecting the destination of a to the origin
of b, in such a way that all three have the same left face after the connection
is complete. The quadedge is recorded in the edges list.

@param a
@param b
@return a quadedge
*/
func (qes *QuadEdgeSubdivision) Connect(a *QuadEdge, b *QuadEdge) *QuadEdge {
	q := Connect(a, b)
	qes.quadEdges = append(qes.quadEdges, q)
	return q
}

/*
Deletes a quadedge from the subdivision. Linked quadedges are updated to
reflect the deletion.

e - the quadedge to delete
*/
func (qes *QuadEdgeSubdivision) Delete(e *QuadEdge) {
	Splice(e, e.OPrev())
	Splice(e.Sym(), e.Sym().OPrev())

	eSym := e.Sym()
	eRot := e.Rot()
	eRotSym := e.Rot().Sym()

	e.Delete()
	eSym.Delete()
	eRot.Delete()
	eRotSym.Delete()

	// this is inefficient on an array, but this method should be called
	// infrequently
	newArray := make([]*QuadEdge, 0, len(qes.quadEdges))
	for _, ele := range qes.quadEdges {
		if ele.IsLive() {
			newArray = append(newArray, ele)

			if ele.next.IsLive() == false {
				log.Fatal("a dead edge is still linked: %v", ele)
			}
		}
	}
	qes.quadEdges = newArray

	if qes.startingEdge.IsLive() == false {
		if len(qes.quadEdges) > 0 {
			qes.startingEdge = qes.quadEdges[0]
		} else {
			qes.startingEdge = nil
		}
	}
}

/*
Locates an edge of a triangle which contains a location specified by a Vertex
v. The edge returned has the property that either v is on e, or e is an edge
of a triangle containing v. The search starts from startEdge amd proceeds on
the general direction of v.

This locate algorithm relies on the subdivision being Delaunay. For
non-Delaunay subdivisions, this may loop for ever.

v - the location to search for
startEdge - an edge of the subdivision to start searching at
return a QuadEdge which contains v, or is on the edge of a triangle containing
v

If the location algorithm fails to converge in a reasonable number of
iterations a ErrLocateFailure will be returned.
*/
func (qes *QuadEdgeSubdivision) LocateFromEdge(v Vertex, startEdge *QuadEdge) (*QuadEdge, error) {
	iter := 0
	maxIter := len(qes.quadEdges)

	e := startEdge

	for {
		iter++

		/*
			So far it has always been the case that failure to locate indicates an
			invalid subdivision. So just fail completely. (An alternative would be
			to perform an exhaustive search for the containing triangle, but this
			would mask errors in the subdivision topology)

			This can also happen if two vertices are located very close together,
			since the orientation predicates may experience precision failures.
		*/
		if iter > maxIter {
			return nil, ErrLocateFailure
			// String msg = "Locate failed to converge (at edge: " + e + ").
			// Possible causes include invalid Subdivision topology or very close
			// sites";
			// System.err.println(msg);
			// dumpTriangles();
		}

		if v.Equals(e.Orig()) || v.Equals(e.Dest()) {
			break
		} else if v.RightOf(*e) {
			e = e.Sym()
		} else if !v.RightOf(*e.ONext()) {
			e = e.ONext()
		} else if !v.RightOf(*e.DPrev()) {
			e = e.DPrev()
		} else {
			// on edge or in triangle containing edge
			break
		}
	}
	// System.out.println("Locate count: " + iter);
	return e, nil
}

/*
Finds a quadedge of a triangle containing a location
specified by a {@link Vertex}, if one exists.

v - the vertex to locate
Return a quadedge on the edge of a triangle which touches or contains the
location or nil if no such triangle exists
*/
func (qes *QuadEdgeSubdivision) Locate(v Vertex) (*QuadEdge, error) {
	return qes.locator.Locate(v)
}

/*
Locates the edge between the given vertices, if it exists in the
subdivision.

p0 a coordinate
p1 another coordinate
Return the edge joining the coordinates, if present or null if no such edge 
exists
*/
func (qes *QuadEdgeSubdivision) LocateSegment(p0 Vertex, p1 Vertex) (*QuadEdge, error) {
	// find an edge containing one of the points
	e, err := qes.locator.Locate(p0);
	if err != nil || e == nil {
		return nil, err
	}

	// normalize so that p0 is origin of base edge
	base := e;
	if (e.Dest().EqualsTolerance(p0, qes.tolerance)) {
		base = e.Sym();
	}
	// check all edges around origin of base edge
	locEdge := base;
	done := false
	for !done {
		if locEdge.Dest().EqualsTolerance(p1, qes.tolerance) {
			return locEdge, nil
		}
		locEdge = locEdge.ONext();

		if locEdge == base {
			done = true
		}
	}
	return nil, nil
}

/**
 * Inserts a new site into the Subdivision, connecting it to the vertices of
 * the containing triangle (or quadrilateral, if the split point falls on an
 * existing edge).
 * <p>
 * This method does NOT maintain the Delaunay condition. If desired, this must
 * be checked and enforced by the caller.
 * <p>
 * This method does NOT check if the inserted vertex falls on an edge. This
 * must be checked by the caller, since this situation may cause erroneous
 * triangulation
 *
 * @param v
 *          the vertex to insert
 * @return a new quad edge terminating in v
 */
// public QuadEdge insertSite(Vertex v) {
// 	QuadEdge e = locate(v);

// 	if ((v.equals(e.orig(), tolerance)) || (v.equals(e.dest(), tolerance))) {
// 		return e; // point already in subdivision.
// 	}

// 	// Connect the new point to the vertices of the containing
// 	// triangle (or quadrilateral, if the new point fell on an
// 	// existing edge.)
// 	QuadEdge base = makeEdge(e.orig(), v);
// 	QuadEdge.splice(base, e);
// 	QuadEdge startEdge = base;
// 	do {
// 		base = connect(e, base.sym());
// 		e = base.oPrev();
// 	} while (e.lNext() != startEdge);

// 	return startEdge;
// }

/*
isFrameEdge tests whether a QuadEdge is an edge incident on a frame triangle
vertex.

e - the edge to test
return true if the edge is connected to the frame triangle
*/
func (qes *QuadEdgeSubdivision) isFrameEdge(e *QuadEdge) bool {
	if qes.isFrameVertex(e.Orig()) || qes.isFrameVertex(e.Dest()) {
		return true
	}
	return false
}

// 	/**
// 	 * Tests whether a QuadEdge is an edge on the border of the frame facets and
// 	 * the internal facets. E.g. an edge which does not itself touch a frame
// 	 * vertex, but which touches an edge which does.
// 	 *
// 	 * @param e
// 	 *          the edge to test
// 	 * @return true if the edge is on the border of the frame
// 	 */
// 	public boolean isFrameBorderEdge(QuadEdge e) {
// 		// MD debugging
// 		QuadEdge[] leftTri = new QuadEdge[3];
// 		getTriangleEdges(e, leftTri);
// 		// System.out.println(new QuadEdgeTriangle(leftTri).toString());
// 		QuadEdge[] rightTri = new QuadEdge[3];
// 		getTriangleEdges(e.sym(), rightTri);
// 		// System.out.println(new QuadEdgeTriangle(rightTri).toString());

// 		// check other vertex of triangle to left of edge
// 		Vertex vLeftTriOther = e.lNext().dest();
// 		if (isFrameVertex(vLeftTriOther))
// 			return true;
// 		// check other vertex of triangle to right of edge
// 		Vertex vRightTriOther = e.sym().lNext().dest();
// 		if (isFrameVertex(vRightTriOther))
// 			return true;

// 		return false;
// 	}

/*
isFrameVertex tests whether a vertex is a vertex of the outer triangle.

v - the vertex to test
returns true if the vertex is an outer triangle vertex
*/
func (qes *QuadEdgeSubdivision) isFrameVertex(v Vertex) bool {
	if v.Equals(qes.frameVertex[0]) {
		return true
	}
	if v.Equals(qes.frameVertex[1]) {
		return true
	}
	if v.Equals(qes.frameVertex[2]) {
		return true
	}
	return false
}

// 	private LineSegment seg = new LineSegment();

/*
IsOnEdge Tests whether a point lies on a QuadEdge, up to a tolerance
determined by the subdivision tolerance.

Returns true if the vertex lies on the edge
*/
func (qes *QuadEdgeSubdivision) IsOnEdge(e *QuadEdge, p geom.Pointer) bool {
	dist := planar.DistanceToLineSegment(p, e.Orig(), e.Dest())

	// heuristic (hack?)
	return dist < qes.edgeCoincidenceTolerance
}

/*
IsOnSegment Tests whether a point lies on a segment, up to a tolerance
determined by the subdivision tolerance.

Returns true if the vertex lies on the edge
*/
func (qes *QuadEdgeSubdivision) IsOnLine(l geom.Line, p geom.Pointer) bool {
	dist := planar.DistanceToLineSegment(p, geom.Point(l[0]), geom.Point(l[1]))

	// heuristic (hack?)
	return dist < qes.edgeCoincidenceTolerance
}

/*
IsVertexOfEdge tests whether a {@link Vertex} is the start or end vertex of a
QuadEdge, up to the subdivision tolerance distance.

Returns true if the vertex is a endpoint of the edge
*/
func (qes *QuadEdgeSubdivision) IsVertexOfEdge(e *QuadEdge, v Vertex) bool {
	if (v.EqualsTolerance(e.Orig(), qes.tolerance)) || (v.EqualsTolerance(e.Dest(), qes.tolerance)) {
		return true
	}
	return false
}

//   /**
//    * Gets the unique {@link Vertex}es in the subdivision,
//    * including the frame vertices if desired.
//    *
// 	 * @param includeFrame
// 	 *          true if the frame vertices should be included
//    * @return a collection of the subdivision vertices
//    *
//    * @see #getVertexUniqueEdges
//    */
//   public Collection getVertices(boolean includeFrame)
//   {
//     Set vertices = new HashSet();
//     for (Iterator i = quadEdges.iterator(); i.hasNext();) {
//       QuadEdge qe = (QuadEdge) i.next();
//       Vertex v = qe.orig();
//       //System.out.println(v);
//       if (includeFrame || ! isFrameVertex(v))
//         vertices.add(v);

//       /**
//        * Inspect the sym edge as well, since it is
//        * possible that a vertex is only at the
//        * dest of all tracked quadedges.
//        */
//       Vertex vd = qe.dest();
//       //System.out.println(vd);
//       if (includeFrame || ! isFrameVertex(vd))
//         vertices.add(vd);
//     }
//     return vertices;
//   }

//   /**
//    * Gets a collection of {@link QuadEdge}s whose origin
//    * vertices are a unique set which includes
//    * all vertices in the subdivision.
//    * The frame vertices can be included if required.
//    * <p>
//    * This is useful for algorithms which require traversing the
//    * subdivision starting at all vertices.
//    * Returning a quadedge for each vertex
//    * is more efficient than
//    * the alternative of finding the actual vertices
//    * using {@link #getVertices} and then locating
//    * quadedges attached to them.
//    *
//    * @param includeFrame true if the frame vertices should be included
//    * @return a collection of QuadEdge with the vertices of the subdivision as their origins
//    */
//   public List getVertexUniqueEdges(boolean includeFrame)
//   {
//   	List edges = new ArrayList();
//     Set visitedVertices = new HashSet();
//     for (Iterator i = quadEdges.iterator(); i.hasNext();) {
//       QuadEdge qe = (QuadEdge) i.next();
//       Vertex v = qe.orig();
//       //System.out.println(v);
//       if (! visitedVertices.contains(v)) {
//       	visitedVertices.add(v);
//         if (includeFrame || ! isFrameVertex(v)) {
//         	edges.add(qe);
//         }
//       }

//       /**
//        * Inspect the sym edge as well, since it is
//        * possible that a vertex is only at the
//        * dest of all tracked quadedges.
//        */
//       QuadEdge qd = qe.sym();
//       Vertex vd = qd.orig();
//       //System.out.println(vd);
//       if (! visitedVertices.contains(vd)) {
//       	visitedVertices.add(vd);
//         if (includeFrame || ! isFrameVertex(vd)) {
//         	edges.add(qd);
//         }
//       }
//     }
//     return edges;
//   }

type edgeStack []*QuadEdge
type edgeSet map[*QuadEdge]bool

func (es *edgeStack) push(edge *QuadEdge) {
	*es = append(*es, edge)
}

func (es *edgeStack) pop() *QuadEdge {
	if len(*es) == 0 {
		return nil
	}
	result := (*es)[len(*es)-1]
	*es = (*es)[:len(*es)-1]
	return result
}

/*
contains returns true if edge is in the map.

This just isn't natural for me yet...
if _, ok := es[edge]; ok {
*/
func (es *edgeSet) contains(edge *QuadEdge) bool {
	_, ok := (*es)[edge]
	return ok
}

/**
 * Gets all primary quadedges in the subdivision.
* A primary edge is a {@link QuadEdge}
 * which occupies the 0'th position in its array of associated quadedges.
 * These provide the unique geometric edges of the triangulation.
 *
 * @param includeFrame true if the frame edges are to be included
 * @return a List of QuadEdges
*/
func (qes *QuadEdgeSubdivision) GetPrimaryEdges(includeFrame bool) []*QuadEdge {
	qes.visitedKey++

	var edges []*QuadEdge
	if qes.startingEdge == nil {
		return edges
	}
	var stack edgeStack
	stack.push(qes.startingEdge)

	visitedEdges := make(edgeSet)

	for len(stack) > 0 {
		edge := stack.pop()

		if !visitedEdges.contains(edge) {
			priQE := edge.GetPrimary()

			if includeFrame || !qes.isFrameEdge(priQE) {
				edges = append(edges, priQE)
			}

			stack.push(edge.ONext())
			stack.push(edge.Sym().ONext())

			visitedEdges[edge] = true
			visitedEdges[edge.Sym()] = true
		}
	}
	return edges
}

//   /**
//    * A TriangleVisitor which computes and sets the
//    * circumcentre as the origin of the dual
//    * edges originating in each triangle.
//    *
//    * @author mbdavis
//    *
//    */
// 	private static class TriangleCircumcentreVisitor implements TriangleVisitor
// 	{
// 		public TriangleCircumcentreVisitor() {
// 		}

// 		public void visit(QuadEdge[] triEdges)
// 		{
// 			Coordinate a = triEdges[0].orig().getCoordinate();
// 			Coordinate b = triEdges[1].orig().getCoordinate();
// 			Coordinate c = triEdges[2].orig().getCoordinate();

// 			// TODO: choose the most accurate circumcentre based on the edges
//       Coordinate cc = Triangle.circumcentre(a, b, c);
// 			Vertex ccVertex = new Vertex(cc);
// 			// save the circumcentre as the origin for the dual edges originating in this triangle
// 			for (int i = 0; i < 3; i++) {
// 				triEdges[i].rot().setOrig(ccVertex);
// 			}
// 		}
// 	}

// 	/*****************************************************************************
// 	 * Visitors
// 	 ****************************************************************************/

func (qes *QuadEdgeSubdivision) visitTriangles(triVisitor func(triEdges []*QuadEdge), includeFrame bool) {
	qes.visitedKey++

	// visited flag is used to record visited edges of triangles
	// setVisitedAll(false);
	var stack *edgeStack = new(edgeStack)
	if qes.startingEdge != nil {
		stack.push(qes.startingEdge)
	}

	visitedEdges := make(edgeSet)

	for len(*stack) > 0 {
		edge := stack.pop()
		if !visitedEdges.contains(edge) {
			triEdges := qes.fetchTriangleToVisit(edge, stack, includeFrame, visitedEdges)
			if triEdges != nil {
				triVisitor(triEdges)
			}
		} 
	}
}

/*
Stores the edges for a visited triangle. Also pushes sym (neighbour) edges
on stack to visit later.

@param edge
@param edgeStack
@param includeFrame
@return the visited triangle edges
or null if the triangle should not be visited (for instance, if it is
        outer)
*/
func (qes *QuadEdgeSubdivision) fetchTriangleToVisit(edge *QuadEdge, stack *edgeStack, includeFrame bool, visitedEdges edgeSet) []*QuadEdge {
	triEdges := make([]*QuadEdge, 0, 3)
	curr := edge
	var isFrame bool
	for true {
		triEdges = append(triEdges, curr)

		if curr.IsLive() == false {
			log.Fatal("traversing dead edge")
		}
		if qes.isFrameEdge(curr) {
			isFrame = true
		}

		// push sym edges to visit next
		sym := curr.Sym()
		if !visitedEdges.contains(sym) {
			stack.push(sym)
		}

		// mark this edge as visited
		visitedEdges[curr] = true

		curr = curr.LNext()

		if curr == edge {
			break
		}
	}

	if isFrame && !includeFrame {
		return nil
	}
	return triEdges
}

// 	/**
// 	 * Gets a list of the triangles
// 	 * in the subdivision, specified as
// 	 * an array of the primary quadedges around the triangle.
// 	 *
// 	 * @param includeFrame
// 	 *          true if the frame triangles should be included
// 	 * @return a List of QuadEdge[3] arrays
// 	 */
// 	public List getTriangleEdges(boolean includeFrame) {
// 		TriangleEdgesListVisitor visitor = new TriangleEdgesListVisitor();
// 		visitTriangles(visitor, includeFrame);
// 		return visitor.getTriangleEdges();
// 	}

// 	private static class TriangleEdgesListVisitor implements TriangleVisitor {
// 		private List triList = new ArrayList();

// 		public void visit(QuadEdge[] triEdges) {
// 			triList.add(triEdges);
// 		}

// 		public List getTriangleEdges() {
// 			return triList;
// 		}
// 	}

// 	/**
// 	 * Gets a list of the triangles in the subdivision,
// 	 * specified as an array of the triangle {@link Vertex}es.
// 	 *
// 	 * @param includeFrame
// 	 *          true if the frame triangles should be included
// 	 * @return a List of Vertex[3] arrays
// 	 */
// 	public List getTriangleVertices(boolean includeFrame) {
// 		TriangleVertexListVisitor visitor = new TriangleVertexListVisitor();
// 		visitTriangles(visitor, includeFrame);
// 		return visitor.getTriangleVertices();
// 	}

// 	private static class TriangleVertexListVisitor implements TriangleVisitor {
// 		private List triList = new ArrayList();

// 		public void visit(QuadEdge[] triEdges) {
// 			triList.add(new Vertex[] { triEdges[0].orig(), triEdges[1].orig(),
// 					triEdges[2].orig() });
// 		}

// 		public List getTriangleVertices() {
// 			return triList;
// 		}
// 	}

/**
Gets the coordinates for each triangle in the subdivision as an array.

@param includeFrame
         true if the frame triangles should be included
@return a list of Coordinate[4] representing each triangle
*/
func (qes *QuadEdgeSubdivision) GetTriangleCoordinates(includeFrame bool) ([]geom.Polygon, error) {
	var visitor TriangleCoordinatesVisitor
	qes.visitTriangles(visitor.visit, includeFrame)
	if visitor.err != nil {
		return nil, visitor.err
	}
	return visitor.getTriangles(), nil
}

type TriangleCoordinatesVisitor struct {
	triCoords []geom.Polygon
	err       error
}

func (tcv *TriangleCoordinatesVisitor) visit(triEdges []*QuadEdge) {
	if tcv.err != nil {
		return
	}

	var triangle geom.Polygon
	triangle = append(triangle, [][2]float64{})
	for i := 0; i < 3; i++ {
		v := triEdges[i].Orig()
		triangle[0] = append(triangle[0], [2]float64(v))
	}
	if len(triangle[0]) > 0 {
		// close the ring
		triangle[0] = append(triangle[0], triangle[0][0])
		if len(triangle[0]) != 4 {
			//checkTriangleSize(pts);
			tcv.err = fmt.Errorf("invalid triangle: %v", triangle)

			return
		}

		tcv.triCoords = append(tcv.triCoords, triangle)
	}
}

func (tcv *TriangleCoordinatesVisitor) getTriangles() []geom.Polygon {
	return tcv.triCoords
}

// private void checkTriangleSize(Coordinate[] pts)
// {
// 	String loc = "";
// 	if (pts.length >= 2)
// 		loc = WKTWriter.toLineString(pts[0], pts[1]);
// 	else {
// 		if (pts.length >= 1)
// 			loc = WKTWriter.toPoint(pts[0]);
// 	}
// 	// Assert.isTrue(pts.length == 4, "Too few points for visited triangle at " + loc);
// 	//com.vividsolutions.jts.util.Debug.println("too few points for triangle at " + loc);
// }

/*
GetEdgesAsMultiLineString gets the geometry for the edges in the subdivision
as a MultiLineString containing 2-point lines.

returns a MultiLineString
*/
func (qes *QuadEdgeSubdivision) GetEdgesAsMultiLineString() geom.MultiLineString {
	quadEdges := qes.GetPrimaryEdges(false)
	var ms geom.MultiLineString
	for _, qe := range quadEdges {
		var ls [][2]float64
		ls = append(ls, qe.Orig().XY(), qe.Dest().XY())
		ms = append(ms, ls)
	}
	return ms
}

/*
GetTriangles gets the geometry for the triangles in a triangulated subdivision
as a MultiPolygon of triangular Polygons.

Unlike JTS, this method returns a MultiPolygon. I found not all viewers like
displaying collections. -JRS

Returns a MultiPolygon of triangular Polygons
*/
func (qes *QuadEdgeSubdivision) GetTriangles() (geom.MultiPolygon, error) {
	tris, err := qes.GetTriangleCoordinates(false)
	if err != nil {
		return nil, err
	}
	var gc geom.MultiPolygon
	for i := 0; i < len(tris); i++ {
		gc = append(gc, tris[i])
	}
	return gc, nil
}

// 	/**
// 	 * Gets the cells in the Voronoi diagram for this triangulation.
// 	 * The cells are returned as a {@link GeometryCollection} of {@link Polygon}s
//    * <p>
//    * The userData of each polygon is set to be the {@link Coordinate}
//    * of the cell site.  This allows easily associating external
//    * data associated with the sites to the cells.
// 	 *
// 	 * @param geomFact a geometry factory
// 	 * @return a GeometryCollection of Polygons
// 	 */
//   public Geometry getVoronoiDiagram(GeometryFactory geomFact)
//   {
//     List vorCells = getVoronoiCellPolygons(geomFact);
//     return geomFact.createGeometryCollection(GeometryFactory.toGeometryArray(vorCells));
//   }

// 	/**
// 	 * Gets a List of {@link Polygon}s for the Voronoi cells
// 	 * of this triangulation.
//    * <p>
//    * The userData of each polygon is set to be the {@link Coordinate}
//    * of the cell site.  This allows easily associating external
//    * data associated with the sites to the cells.
// 	 *
// 	 * @param geomFact a geometry factory
// 	 * @return a List of Polygons
// 	 */
//   public List getVoronoiCellPolygons(GeometryFactory geomFact)
//   {
//   	/*
//   	 * Compute circumcentres of triangles as vertices for dual edges.
//   	 * Precomputing the circumcentres is more efficient,
//   	 * and more importantly ensures that the computed centres
//   	 * are consistent across the Voronoi cells.
//   	 */
//   	visitTriangles(new TriangleCircumcentreVisitor(), true);

//     List cells = new ArrayList();
//     Collection edges = getVertexUniqueEdges(false);
//     for (Iterator i = edges.iterator(); i.hasNext(); ) {
//     	QuadEdge qe = (QuadEdge) i.next();
//       cells.add(getVoronoiCellPolygon(qe, geomFact));
//     }
//     return cells;
//   }

//   /**
//    * Gets the Voronoi cell around a site specified
//    * by the origin of a QuadEdge.
//    * <p>
//    * The userData of the polygon is set to be the {@link Coordinate}
//    * of the site.  This allows attaching external
//    * data associated with the site to this cell polygon.
//    *
//    * @param qe a quadedge originating at the cell site
//    * @param geomFact a factory for building the polygon
//    * @return a polygon indicating the cell extent
//    */
//   public Polygon getVoronoiCellPolygon(QuadEdge qe, GeometryFactory geomFact)
//   {
//     List cellPts = new ArrayList();
//     QuadEdge startQE = qe;
//     do {
// //    	Coordinate cc = circumcentre(qe);
//     	// use previously computed circumcentre
//     	Coordinate cc = qe.rot().orig().getCoordinate();
//       cellPts.add(cc);

//       // move to next triangle CW around vertex
//       qe = qe.oPrev();
//     } while (qe != startQE);

//     CoordinateList coordList = new CoordinateList();
//     coordList.addAll(cellPts, false);
//     coordList.closeRing();

//     if (coordList.size() < 4) {
//       System.out.println(coordList);
//       coordList.add(coordList.get(coordList.size()-1), true);
//     }

//     Coordinate[] pts = coordList.toCoordinateArray();
//     Polygon cellPoly = geomFact.createPolygon(geomFact.createLinearRing(pts));

//     Vertex v = startQE.orig();
//     cellPoly.setUserData(v.getCoordinate());
//     return cellPoly;
//   }

// }



