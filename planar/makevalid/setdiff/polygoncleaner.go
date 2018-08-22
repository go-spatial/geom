package setdiff

import (
	"errors"
	"log"
	"sort"
	"strings"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar/triangulate/constraineddelaunay"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

var ErrUnsupportedGeometryType = errors.New("unsupported geometry type")

// these errors should only be returned if there is an internal bug/error
var ErrNoExternalTriangle = errors.New("no external triangle found")
var ErrExtractingGeomFailed = errors.New("extracting geometry failed")

/*
X Maintain the source of each constraint when building the triangulation. Maybe with the data pointer? A single constraint may come from multiple sources.
X Build a boolean OddEven inside/outside map for each input linear ring
X Determine the final triangle class (inside/outside) using boolean logic on the inside/outside map. This should enable intersection, union, etc.
X As listed in "2012 - Automatically repairing invalid polygons with a constrained triangulation", use a stack to push/pop line strings each time an intersection is found.
  + The first/last line string is the exterior, all others are interiors.
  + If a linestring is mirrored it is degenerate and can be erased
  + If the linestring is closed, create a ring out of it.

* Get all tests working and commit
* Add a function for adding sites when constrained edges interesect.
  + Run before insertEdgeCDT
  + Traverse the triangulation in a similar fashion to insertEdgeCDT
  + If an intersection is found, insert a new site, then recrusively continue searching for more intersections
  + Inserting a new site should just greedily add the best edges around the site. Don't be too concerned with maintaining the Delaunay rules, because we can't
*/

type ringReference struct {
	exterior   bool
	linearRing geom.LineString
}

/*
Polygon provides methods for cleaning polygons and multipolygons. Collections
of polygons and multipolygons will also work.

Outer rings will be counter-clockwise and inner rings will be clockwise.

[2] describes the SetDiff method for labeling triangles. [1] describes the
method for converting triangles into linear rings and geometries (section 3.5)

1. Ohori, Ken Arroyo, Hugo Ledoux, and Martijn Meijers. "Validation and automatic repair of planar partitions using a constrained triangulation." Photogrammetrie-Fernerkundung-Geoinformation 2012, no. 5 (2012): 613-630.
http://www.dgpf.de/pfg/2012/pfg2012_5_Arroyo-Ohori.pdf

2. Ledoux, Hugo, Ken Arroyo Ohori, and Martijn Meijers. "A triangulation-based approach to automatically repair GIS polygons." Computers & Geosciences 66 (2014): 121-131.
https://pdfs.semanticscholar.org/d9ec/b32a7844b436fcd4757958e5eeca9563fcd2.pdf

*/
type PolygonCleaner struct {
	builder   *constraineddelaunay.Triangulator
	subdiv    *quadedge.QuadEdgeSubdivision
	tolerance float64
	// run validation after many modification operations. This is expensive,
	// but very useful when debugging.
	validate  bool
	rings     []geom.Geometry
	ringRefs  []*ringReference
	triLabels []map[constraineddelaunay.Triangle]bool
	// the triangles that have been extracted into geometries.
	extractedTris map[constraineddelaunay.Triangle]bool
	// the triangles that have been not been extracted into geometries and are
	// labeled as inside.
	unextractedTris []constraineddelaunay.Triangle
}

func (pc *PolygonCleaner) addGeometry(g geom.Geometry) error {
	switch gg := g.(type) {
	default:
		return geom.ErrUnknownGeometry{g}
	case geom.Pointer:
		return ErrUnsupportedGeometryType
	case geom.MultiPointer:
		return ErrUnsupportedGeometryType
	case geom.LineStringer:
		return ErrUnsupportedGeometryType
	case geom.MultiLineStringer:
		return ErrUnsupportedGeometryType
	case geom.Polygoner:
		exterior := true
		for _, ls := range gg.LinearRings() {
			pc.addRing(exterior, geom.LineString(ls))
			exterior = false
		}
		return nil
	case geom.MultiPolygoner:
		for _, p := range gg.Polygons() {
			if err := pc.addGeometry(geom.Polygon(p)); err != nil {
				return err
			}
		}
		return nil
	case geom.Collectioner:
		for _, child := range gg.Geometries() {
			if err := pc.addGeometry(child); err != nil {
				return err
			}
		}
		return nil
	}
}

/*
If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) addRing(exterior bool, r geom.LineString) {
	if len(r) > 1 {
		rr := &ringReference{exterior, r}
		pc.ringRefs = append(pc.ringRefs, rr)

		// make sure the line string has the same first/last point
		if cmp.PointEqual(r[0], r[len(r)-1]) == false {
			r = append(r, r[0])
		}

		pc.rings = append(pc.rings, geom.LineString(r))
		pc.triLabels = append(pc.triLabels, make(map[constraineddelaunay.Triangle]bool, 0))
	}
}

/*
If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) findUnextractedTri() *constraineddelaunay.Triangle {
	// put the unextracted triangles into an array and sort based on centroid
	if len(pc.extractedTris) == 0 {
		pc.unextractedTris = make([]constraineddelaunay.Triangle, 0)
		for k := range pc.triLabels[0] {
			if pc.isInside(k) {
				pc.unextractedTris = append(pc.unextractedTris, k)
				if debug {
					log.Printf("k: %v", k.String())
				}
			}
		}

		sort.Sort(constraineddelaunay.TriangleByCentroid(pc.unextractedTris))
	}

	// remove the last entry. if it hasn't been extracted return it
	for len(pc.unextractedTris) > 0 {
		t := pc.unextractedTris[0]
		pc.unextractedTris = pc.unextractedTris[1:len(pc.unextractedTris)]
		if debug {
			log.Printf("%v", pc.unextractedTris)
		}
		if len(pc.extractedTris) == 0 || pc.extractedTris[t] == false {
			return &t
		}
	}

	return nil
}

type edgeString []*quadedge.QuadEdge
type edgeStack []edgeString

/*
Returns true if qe connects to the end of the edgeString or the edgeString is
empty.
*/
func (es edgeString) connects(qe *quadedge.QuadEdge) bool {
	if len(es) == 0 || qe.Orig().Equals(es[len(es)-1].Dest()) {
		return true
	}

	return false
}

func (es edgeString) isClosed() bool {
	if es == nil || len(es) <= 1 {
		return false
	}

	if es[0].Orig().Equals(es[len(es)-1].Dest()) {
		return true
	}

	return false
}

func (es edgeString) toLinearRing() geom.LineString {
	if es == nil || len(es) <= 1 {
		return geom.LineString{}
	}

	result := geom.LineString{}
	for i := range es {
		result = append(result, es[i].Orig())
	}
	result = append(result, es[0].Orig())

	return result
}

/*
peeks the last edgeString in the stack

If es is nil a panic will occur.
*/
func (es *edgeStack) peek() *edgeString {
	if len(*es) == 0 {
		(*es).push(edgeString{})
	}
	return &((*es)[len(*es)-1])
}

/*
push pushes an edge onto the edgeStack

If es is nil a panic will occur.
*/
func (es *edgeStack) push(s edgeString) {
	*es = append(*es, s)
}

/*
pop pops an edge off the edgeStack

If es is nil a panic will occur.
*/
func (es *edgeStack) pop() edgeString {
	if len(*es) == 0 {
		return nil
	}
	result := (*es)[len(*es)-1]
	*es = (*es)[:len(*es)-1]
	return result
}

/*
If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) followEdges(qe *quadedge.QuadEdge, es *edgeStack, result *geom.Polygon) error {

	tri := constraineddelaunay.Triangle{qe}
	tri = *tri.Normalize()

	if pc.extractedTris[tri] {
		return nil
	}

	if debug {
		log.Printf("tri: %v", tri.String())
	}
	pc.extractedTris[tri] = true

	e := qe
	for {
		if debug {
			log.Printf("e: %v", e)
		}
		if pc.isRingEdge(e) {
			if es.peek().connects(e.Sym()) == false {
				if debug {
					log.Printf("No Connection: %v %v", es, e)
				}
				// this doesn't connect so create a new edge string
				es.push(edgeString{})
			}
			top := es.peek()
			*top = append(*top, e.Sym())
			if debug {
				log.Printf("top: %v e: %v", top, e)
			}

			if top.isClosed() {
				*result = append(*result, es.pop().toLinearRing())
				if debug {
					log.Printf("result: %v", result)
				}
			}
		} else {
			if err := pc.followEdges(e.Sym(), es, result); err != nil {
				return err
			}
			if debug {
				log.Printf("es: %v", es)
			}
		}
		e = e.RNext()

		if e == qe {
			break
		}
	}
	if debug {
		log.Printf("es: %v", es)
	}

	return nil
	// push the first edge onto the stack

	// The first/last line string is the exterior, all others are interiors.

	// follow the edge
	// if we found an intersection, push it on to the stack and follow
	// if we are back at the beginning, record it and pop the stack

	// If a linestring is mirrored it is degenerate and can be erased

	// If the linestring is closed, create a ring out of it.

}

/*

As listed in [1], use a stack to push/pop line strings each time an
intersection is found.
  + The first/last line string is the exterior, all others are interiors.
  + If a linestring is mirrored it is degenerate and can be erased
  + If the linestring is closed, create a ring out of it.

If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) extractGeometry() (geom.Geometry, error) {
	pc.extractedTris = make(map[constraineddelaunay.Triangle]bool, 0)

	result := geom.MultiPolygon{}

	// while we are still finding more geometries
	for {
		// find a starting triangle
		startingTri := pc.findUnextractedTri()
		if debug {
			log.Printf("startingTri: %v", startingTri)
		}

		// if we didn't find a starting triangle, exit
		if startingTri == nil {
			break
		}

		poly := geom.Polygon{}
		es := edgeStack{}
		if err := pc.followEdges(startingTri.Qe, &es, &poly); err != nil {
			return nil, err
		}
		if len(es) > 0 {
			if debug {
				log.Printf("Expected es to be empty after run: %v", es)
			}
			return nil, ErrExtractingGeomFailed
		}

		result = append(result, poly)

	}
	if debug {
		log.Printf("result: %v", result)
	}

	// As listed in [1], use a stack to push/pop line strings each time an
	// intersection is found.

	if len(result) == 0 {
		return nil, nil
	}
	if len(result) == 1 {
		return geom.Polygon(result[0]), nil
	}
	return result, nil
}

func (pc *PolygonCleaner) getLabelsAsString() string {
	if len(pc.triLabels) == 0 {
		return ""
	}
	// TODO: Update to use the ring hierarchy when determining inside/outside
	result := []string{}

	l := pc.triLabels[0]
	for k := range l {
		if pc.isInside(k) {
			result = append(result, "inside: "+k.String())
		}
	}

	// make sure the ordering is consistent for testing
	sort.Strings(result)

	return strings.Join(result, "\n")
}

/*
Returns true if the triangle is inside, otherwise false.

This must be called after all triangled have been labeled for each ring.

tri must be normalized before being called.

If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) isInside(tri constraineddelaunay.Triangle) bool {
	inside := false

	for i := range pc.ringRefs {
		v, ok := pc.triLabels[i][tri]
		if !ok {
			if debug {
				log.Printf("subdiv: %v", pc.subdiv.DebugDumpEdges())
				log.Printf("tri: %v lables: %v", tri, pc.triLabels[i])
			}
			// did you call labelTriangles first?
			log.Fatalf("triangle was not labeled properly")
		}

		if pc.ringRefs[i].exterior == true {
			// each time we hit an exterior this is a new polygon. If the
			// state is inside at this point we consider the label to be
			// inside.
			if inside == true {
				return true
			}
			inside = v
		} else {
			if v {
				inside = false
			}
		}
	}

	return inside
}

/*
Returns true if there are an odd number of references to rr in qe.

If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) isLegitEdge(ri int, qe *quadedge.QuadEdge) bool {
	if qe.GetData() == nil {
		return false
	}
	arr, ok := qe.GetData().([]interface{})
	// this should never happen
	if !ok {
		log.Fatalf("could not cast data to array of interfaces")
	}

	count := 0
	for i := range arr {
		rr, ok := arr[i].(*ringReference)
		// this should never happen
		if !ok {
			log.Fatalf("could not data element to ringReference")
		}

		if rr == pc.ringRefs[ri] {
			count++
		}
	}

	// if there are an even number of edges they cancel out.
	return count%2 == 1
}

/*
Returns true if the finalized triangle labels on either side of an edge differ.

If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) isRingEdge(qe *quadedge.QuadEdge) bool {
	t1 := constraineddelaunay.Triangle{qe}
	t1 = *t1.Normalize()
	t2 := constraineddelaunay.Triangle{qe.Sym()}
	t2 = *t2.Normalize()
	return pc.isInside(t1) != pc.isInside(t2)
}

/*
If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) labelTriangles(inside bool, ri int, tri *constraineddelaunay.Triangle) error {

	if debug && inside {
		log.Printf("inside! %v", tri)
	}

	pc.triLabels[ri][*tri] = inside
	if debug {
		log.Printf("tri: %v", tri)
	}

	qe := tri.GetStartEdge()
	if debug {
		log.Printf("Starting qe: %v", qe)
	}
	// go through each edge
	for i := 0; i < 3; i++ {
		if pc.subdiv.IsFrameVertex(qe.Orig()) == false || pc.subdiv.IsFrameVertex(qe.Dest()) == false {
			// if the edge is not a frame edge, evaluate the triangle

			neighbor := &constraineddelaunay.Triangle{qe.Sym()}
			neighbor = neighbor.Normalize()

			if _, ok := pc.triLabels[ri][*neighbor]; ok == false {
				// if we haven't already visited this triangle

				if debug {
					log.Printf("neighbor: %v", neighbor)
				}
				if debug && pc.isLegitEdge(ri, qe) != pc.isLegitEdge(ri, qe.Sym()) {
					log.Printf("WHOOPS!")
				}
				if pc.isLegitEdge(ri, qe) {
					if debug {
						log.Printf("legit edge: %v", qe)
						log.Printf("inside: %v", inside)
					}
					// if the edge contains an odd number of references to the
					// ring then invert the value of inside for the recursive
					// call
					pc.labelTriangles(!inside, ri, neighbor)
				} else {
					pc.labelTriangles(inside, ri, neighbor)
				}
			}
		}
		qe = qe.RNext()
	}

	return nil
}

/*
If pc is nil a panic will occur.
*/
func (pc *PolygonCleaner) MakeValid(g geom.Geometry) (geom.Geometry, error) {
	pc.builder = new(constraineddelaunay.Triangulator)
	pc.rings = make([]geom.Geometry, 0)
	pc.ringRefs = make([]*ringReference, 0)

	if err := pc.addGeometry(g); err != nil {
		return nil, err
	}

	tmp := make([]interface{}, len(pc.ringRefs))
	for i := range pc.ringRefs {
		tmp[i] = pc.ringRefs[i]
	}
	if debug {
		log.Printf("pc.rings: %v", pc.rings)
	}
	if err := pc.builder.InsertGeometries(pc.rings, tmp); err != nil {
		return nil, err
	}

	pc.subdiv = pc.builder.GetSubdivision()
	if debug {
		log.Print(pc.subdiv.DebugDumpEdges())
	}

	start := pc.builder.GetExteriorTriangle().Normalize()
	if debug {
		log.Print(start.Qe)
	}
	if start == nil {
		return nil, ErrNoExternalTriangle
	}

	for i := range pc.rings {
		if err := pc.labelTriangles(false, i, start); err != nil {
			return nil, err
		}
	}

	result, err := pc.extractGeometry()
	if err != nil {
		return nil, err
	}
	if debug {
		log.Printf("debug: %v", pc.subdiv.DebugDumpEdges())
	}

	// TODO
	return result, nil
}
