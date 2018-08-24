package delaunay

import (
	"errors"
	"log"
	"sort"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

type ConstrainedBuilder struct {
	Builder
	constraints edgeMap
}

const TOLERANCE = 0.0001

func NewConstrained(tolerance float64, points []geom.Point, constraints []geom.Line) (cb ConstrainedBuilder) {

	sort.Sort(sort.Reverse(byLength(constraints)))
	// We need to normalize and unique the constraints
	cb.constraints = make(edgeMap, len(constraints))
	for i := range constraints {
		cb.constraints.AddEdge(constraints[i])
	}

	// Make a copy so we don't mess with the original.
	pts := make([]geom.Point, len(points))
	copy(pts, points)

	// Let's add any points from the constraints to the points that will make up the triangulation
	for i := range constraints {
		pts = append(pts, geom.Point(constraints[i][0]), geom.Point(constraints[i][1]))
	}
	uniquePoints := planar.SortUniquePoints(pts)
	cb.Builder.Tolerance = tolerance
	cb.Builder.siteCoords = make([]quadedge.Vertex, len(uniquePoints))

	// free up memory.
	for i := range uniquePoints {
		cb.Builder.siteCoords[i] = quadedge.Vertex(uniquePoints[i])
	}
	return cb
}

func (cb *ConstrainedBuilder) addConstraint(l geom.Line) error {
	if cb == nil {
		return errors.New("uninitilized ConstrainedBuilder")
	}
	if cb.constraints == nil {
		cb.constraints = make(edgeMap)
	}
	cb.constraints.AddEdge(l)
	lenSite := len(cb.siteCoords)
	pts := make([]geom.Point, lenSite+2)
	for i := range cb.Builder.siteCoords {
		pts[i] = geom.Point(cb.Builder.siteCoords[i])
	}
	pts[lenSite] = l[0]
	pts[lenSite+1] = l[1]

	uniquePoints := planar.SortUniquePoints(pts)

	cb.Builder.siteCoords = make([]quadedge.Vertex, len(uniquePoints))

	// free up memory.
	for i := range uniquePoints {
		cb.Builder.siteCoords[i] = quadedge.Vertex(uniquePoints[i])
	}

	return nil
}

type byFirstXY []geom.Line

func (l byFirstXY) Len() int      { return len(l) }
func (l byFirstXY) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byFirstXY) Less(i, j int) bool {
	if cmp.PointEqual(l[i][0], l[j][0]) {
		return cmp.PointLess(l[i][1], l[j][1])
	}
	return cmp.PointLess(l[i][0], l[j][0])
}

type byLength []geom.Line

func (l byLength) Len() int      { return len(l) }
func (l byLength) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byLength) Less(i, j int) bool {
	lilen := l[i].LenghtSquared()
	ljlen := l[j].LenghtSquared()
	if lilen == ljlen {
		if cmp.PointEqual(l[i][0], l[j][0]) {
			return cmp.PointLess(l[i][1], l[j][1])
		}
		return cmp.PointLess(l[i][0], l[j][0])
	}
	return lilen < ljlen
}

//appendNonRepeat only appends the provided value if it does not repeat the last
// value that was appended onto the array.
func appendNonRepeat(arr []geom.Point, v geom.Point) []geom.Point {
	if len(arr) == 0 || !cmp.PointerEqual(arr[len(arr)-1], v) {
		return append(arr, v)
	}
	return arr
}

// removeEdgeFromIndex will remove the the given edge from the given edge index
func removeEdgeFromIndex(index map[quadedge.Vertex]*quadedge.QuadEdge, e *quadedge.QuadEdge) {

	toRemove := [4]*quadedge.QuadEdge{e, e.Sym(), e.Rot(), e.Rot().Sym()}
	shouldRemove := func(e *quadedge.QuadEdge) bool {
		for i := range toRemove {
			if toRemove[i] == e {
				return true
			}
		}
		return false
	}

NextVertex:
	for _, v := range [...]quadedge.Vertex{e.Orig(), e.Dest()} {
		ve := index[v]
		if !shouldRemove(ve) {
			continue
		}
		for testEdge := ve.ONext(); ; testEdge = testEdge.ONext() {
			if testEdge == ve {
				// We made it all the way around the vertex without finding a valid edge to reference from this vertex.
				// All edges are gone?
				return
			}
			if !shouldRemove(testEdge) {
				index[v] = testEdge
				continue NextVertex
			}
		}
	} // for NextVertex
}

func intersectedEdges(startingQE *quadedge.QuadEdge, end quadedge.Vertex) (intersected []*quadedge.QuadEdge, err error) {

	start := startingQE.Orig()
	// Need to figure out all the edges that would intersect with this new edge.
	// First we need to find an intersecting triangle.
	// startingQE is a good stating point for searching for the triangle.
	// Precondition: a,b ∈ T and ab ∉ T
	// Find the triangle t ∈ T that contains a and is cut by ab
	t, err := quadedge.FindIntersectingTriangle(startingQE, end)

	// TODO(gdey): ConincidentEdges
	if err != nil {
		if debug {
			log.Println("returning error:", err)
		}
		return nil, err
	}

	runCount := 0

	var tseq quadedge.Triangle
	var vseq quadedge.Vertex
	var shared *quadedge.QuadEdge
	currentVertex := start

	// While b not in t do
	for !t.IntersectsPoint(end) {
		if debug {
			log.Printf("t(%v)\n%v", runCount, wkt.MustEncode(t.Points()))
		}

		if tseq, err = t.OpposedTriangle(currentVertex); err != nil {
			return nil, err
		}

		if debug {
			log.Printf("tseq(%v)\n%v", runCount, wkt.MustEncode(tseq.Points()))
		}

		// Find the shared edge between the two triangles
		if shared, err = t.SharedEdge(tseq); err != nil {
			return nil, err
		}

		if vseq, err = tseq.OpposedVertex(t); err != nil {
			return nil, err
		}

		switch vseq.Classify(start, end) {
		case quadedge.LEFT:
			currentVertex = shared.Orig()
		case quadedge.RIGHT:
			currentVertex = shared.Dest()
		}

		intersected = append(intersected, shared)
		t = tseq
		if debug {
			runCount++
		}
	} // for !t.IntersectPoint
	return intersected, nil
}

func (cb *ConstrainedBuilder) insertConstraint(constraint geom.Line, i int) error {

	debugEdge := 472

	// Let's build our lookup table
	vertexIndex := cb.subdiv.VertexIndex()

	if debug {
		log.Printf("Constraint(%v)\n%v", i, wkt.MustEncode(constraint))
	}
	startingQE := vertexIndex[quadedge.Vertex(constraint[0])]
	startVertex, endVertex := quadedge.Vertex(constraint[0]), quadedge.Vertex(constraint[1])

	// check to see if the constraint edge already exists in the triangulation.
	found, err := quadedge.LocateSegment(startingQE, endVertex)
	if err != nil {
		return err
	}
	if found != nil {
		// nothing to change; the edge already exists.
		if debug && i == debugEdge {
			log.Println("Edge already exists.", found)
			log.Printf("Triangulation:\n%v\n", cb.subdiv.DebugDumpEdges())

		}
		return nil
	}

	// We did not find an edge.

	removalList, err := intersectedEdges(startingQE, endVertex)
	if err != nil {
		if debug {
			log.Println("returning error:", err)
		}
		return err
	}

	var (
		pu = []geom.Point{geom.Point(startVertex)}
		pl = []geom.Point{geom.Point(startVertex)}
	)

	for _, e := range removalList {

		if cb.subdiv.IsFrameEdge(e) {
			continue
		}

		for i, sharedVertex := range [2]quadedge.Vertex{e.Orig(), e.Dest()} {
			classification := sharedVertex.Classify(startVertex, endVertex)
			switch classification {
			case quadedge.LEFT:
				pl = appendNonRepeat(pl, geom.Point(sharedVertex))
			case quadedge.RIGHT:
				pu = appendNonRepeat(pu, geom.Point(sharedVertex))
			default:
				if debug {
					log.Printf("[%v -- %v] Skipping adding Vertex(%v): %v", classification, e, i, sharedVertex)
					for i := range removalList {
						log.Printf("RemovalList(%v): %v", i, removalList[i])
					}
					log.Printf("Triangulation:\n%v\n", cb.subdiv.DebugDumpEdges())
					log.Printf("Constraint(%v):\n%v\n", i, wkt.MustEncode(constraint))
					log.Printf("Constraints:\n%v\n", wkt.MustEncode(cb.constraints.Edges()))
					log.Printf("Raw Constraints:\n%#v\n", cb.constraints.Edges())
					panic("Die!!!")
				}
			}
		}

		removeEdgeFromIndex(vertexIndex, e)
		// TODO: this call is horribly inefficent and should be optimized
		cb.subdiv.Delete(e)
	}
	pu = appendNonRepeat(pu, geom.Point(endVertex))
	pl = appendNonRepeat(pl, geom.Point(endVertex))

	if debug {
		log.Printf("Constrained edge %v : \n%v", i, wkt.MustEncode(constraint))
		log.Printf("Main Polygon edges after removing intersecting edges\n%v", cb.subdiv.DebugDumpEdges())
	}

	pupllabel := [2]string{"pu", "pl"}
	for pi, pts := range [2][]geom.Point{pu, pl} {
		if debug {
			log.Printf("Triangulating(%v) \n %v", pupllabel[pi], wkt.MustEncode(pts))
		}
		if len(pts) == 2 {
			// there weren't any point to triangulate just he shared line.
			continue
		}
		// Now we need to triangulate the upper and lower pseudoPolygons.
		edges, err := triangulatePseudoPolygon(pts)
		if err != nil {
			log.Printf("Origianl Point List %v(%v) : \n%v", pupllabel[pi], len(pts), wkt.MustEncode(pts))
			log.Printf("Got an error: %v", err)
			log.Printf("Constraint(%v):\n%v\n", i, wkt.MustEncode(constraint))
			log.Printf("Triangulation After Removal:\n%v\n", cb.subdiv.DebugDumpEdges())
			log.Printf("Constraints:\n%v\n", wkt.MustEncode(cb.constraints.Edges()))
			log.Printf("Raw Constraints:\n%#v\n", cb.constraints.Edges())
			return err
		}
		for i, edge := range edges {
			// First we need to check that the edge does not intersect other edges, this can happen if the polygon
			// we were triangulating happen to be concave. In which case it is possible to have a triangle outside
			// of the "ok" region, and we should ignore those edges.
			intersectList, _ := intersectedEdges(startingQE, endVertex)
			if len(intersectList) > 0 {
				continue
			}

			if err := cb.subdiv.InsertEdge(vertexIndex, quadedge.Vertex(edge[0]), quadedge.Vertex(edge[1])); err != nil {
				return err
			}

			if debug {
				log.Printf("Main Polygon edges after pts[%v] edge insert(%v)\n%v\n%v", pi, i, wkt.MustEncode(edge), cb.subdiv.DebugDumpEdges())
			}
		}
	}

	return nil
}

func (cb *ConstrainedBuilder) Triangles(withFrame bool) (tris []geom.Triangle, err error) {

	if debug {
		log.Println("Building Constraint Triangles")
	}
	if err = cb.Builder.initSubdiv(); err != nil {
		if debug {
			log.Println("returning error:", err)
		}
		return nil, err
	}

	// Let's print out the triangulated polygon.
	if debug {
		log.Printf("Initial triangulation\n%v", cb.subdiv.DebugDumpEdges())
	}

	for i, constraint := range cb.constraints.Edges() {
		if err := cb.insertConstraint(constraint, i); err != nil {
			return nil, err
		}
	} // for range constraints

	cb.subdiv.VisitTriangles(func(triEdges []*quadedge.QuadEdge) {
		var triangle geom.Triangle
		if len(triEdges) != 3 {
			edges := cb.constraints.Edges()
			log.Printf("Something weird!")
			var pts []geom.Point
			for i := range triEdges {
				v := triEdges[i].Orig()
				pts = append(pts, geom.Point(v))
			}
			log.Printf("Points:\n%v", wkt.MustEncode(pts))
			log.Printf("Current triangulation edges\n%v", cb.subdiv.DebugDumpEdges())
			log.Printf("Constraints: %#v", edges)
			log.Printf("Constraints:\n%v", wkt.MustEncode(edges))
			return

		}

		for i := 0; i < 3; i++ {
			v := triEdges[i].Orig()
			triangle[i] = [2]float64(v)
		}
		tris = append(tris, triangle)
	}, withFrame)
	return tris, nil
}
