package subdivision

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar/triangulate/geometry"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar"
)

var (
	ErrCancel           = errors.New("canceled walk")
	ErrCoincidentEdges  = errors.New("coincident edges")
	ErrDidNotFindToFrom = errors.New("did not find to and from edge")
)

type VertexIndex map[geometry.Point]*quadedge.Edge

type Subdivision struct {
	startingEdge *quadedge.Edge
	ptcount      int
	frame        [3]geometry.Point
}

// New initialize a subdivision to the triangle defined by the points a,b,c.
func New(a, b, c geometry.Point) *Subdivision {
	ea := quadedge.New()
	ea.EndPoints(&a, &b)
	eb := quadedge.New()
	quadedge.Splice(ea.Sym(), eb)
	eb.EndPoints(&b, &c)

	ec := quadedge.New()
	ec.EndPoints(&c, &a)
	quadedge.Splice(eb.Sym(), ec)
	quadedge.Splice(ec.Sym(), ea)
	return &Subdivision{
		startingEdge: ea,
		ptcount:      3,
		frame:        [3]geometry.Point{a, b, c},
	}
}

func NewForPoints(ctx context.Context, points [][2]float64) *Subdivision {
	sort.Sort(cmp.ByXY(points))
	tri := geometry.TriangleContaining(points...)
	ttri := [3]geometry.Point{geometry.NewPoint(tri[0][0], tri[0][1]), geometry.NewPoint(tri[1][0], tri[1][1]), geometry.NewPoint(tri[2][0], tri[2][1])}
	sd := New(ttri[0], ttri[1], ttri[2])
	var oldPt geometry.Point
	for i, pt := range points {
		if ctx.Err() != nil {
			return nil
		}
		bfpt := geometry.NewPoint(pt[0], pt[1])
		if i != 0 && geometry.ArePointsEqual(oldPt, bfpt) {
			continue
		}
		oldPt = bfpt
		if !sd.InsertSite(bfpt) {
			log.Printf("Failed to insert point %v", bfpt)
		}
	}
	return sd
}

func ptEqual(x geometry.Point, a *geometry.Point) bool {
	if a == nil {
		return false
	}
	return geometry.ArePointsEqual(*a, x)
}

func testEdge(x geometry.Point, e *quadedge.Edge) (*quadedge.Edge, bool) {
	switch {
	case ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()):
		return e, true
	case quadedge.RightOf(x, e):
		return e.Sym(), false
	case !quadedge.RightOf(x, e.ONext()):
		return e.ONext(), false
	case !quadedge.RightOf(x, e.DPrev()):
		return e.DPrev(), false
	default:
		return e, true
	}
}

func locate(se *quadedge.Edge, x geometry.Point, limit int) (*quadedge.Edge, bool) {
	var (
		e     *quadedge.Edge
		ok    bool
		count int
	)
	for e, ok = testEdge(x, se); !ok; e, ok = testEdge(x, e) {
		if limit > 0 {

			count++
			if e == se || count > limit {
				log.Println("searching all edges for", x)
				e = nil

				WalkAllEdges(se, func(ee *quadedge.Edge) error {
					if _, ok = testEdge(x, ee); ok {
						e = ee
						return ErrCancel
					}
					return nil
				})
				log.Printf(
					"Got back to starting edge after %v iterations, only have %v points ",
					count,
					limit,
				)
				return e, false
			}
		}
	}
	return e, true

}

func (sd *Subdivision) VertexIndex() VertexIndex {
	return NewVertexIndex(sd.startingEdge)
}

// NewVertexIndex will return a new vertex index given a starting edge.
func NewVertexIndex(startingEdge *quadedge.Edge) VertexIndex {
	vx := make(VertexIndex)
	WalkAllEdges(startingEdge, func(e *quadedge.Edge) error {
		vx.Add(e)
		return nil
	})
	return vx
}
func (vx VertexIndex) Add(e *quadedge.Edge) {
	var (
		ok   bool
		orig = *e.Orig()
		dest = *e.Dest()
	)
	if _, ok = vx[orig]; !ok {
		vx[orig] = e
	}
	if _, ok = vx[dest]; !ok {
		vx[dest] = e.Sym()
	}
}

func (vx VertexIndex) Remove(e *quadedge.Edge) {
	// Don't think I need e.Rot() and e.Rot().Sym() in this list
	// as they are face of the quadedge.
	toRemove := [4]*quadedge.Edge{e, e.Sym(), e.Rot(), e.Rot().Sym()}
	shouldRemove := func(e *quadedge.Edge) bool {
		for i := range toRemove {
			if toRemove[i] == e {
				return true
			}
		}
		return false
	}

	for _, v := range [...]geometry.Point{*e.Orig(), *e.Dest()} {
		ve := vx[v]
		if ve == nil || !shouldRemove(ve) {
			continue
		}
		delete(vx, v)
		// See if the ccw edge is the same as us, if it's isn't
		// then use that as the edge for our lookup.
		if ve != ve.ONext() {
			vx[v] = ve.ONext()
		}
	}
}

// locate returns an edge e, s.t. either x is on e, or e is an edge of
// a triangle containing x. The search starts from startingEdge
// and proceeds in the general direction of x. Based on the
// pseudocode in Guibas and Stolfi (1985) p.121
func (sd *Subdivision) locate(x geometry.Point) (*quadedge.Edge, bool) {
	return locate(sd.startingEdge, x, sd.ptcount*2)
}

func (sd *Subdivision) FindEdge(vertexIndex VertexIndex, start, end geometry.Point) *quadedge.Edge {
	if vertexIndex == nil {
		vertexIndex = sd.VertexIndex()
	}
	return vertexIndex[start].FindONextDest(end)
}

// InsertSite will insert a new point into a subdivision representing a Delaunay
// triangulation, and fixes the affected edges so that the result
// is  still a Delaunay triangulation. This is based on the pseudocode
// from Guibas and Stolfi (1985) p.120, with slight modificatons and a bug fix.
func (sd *Subdivision) InsertSite(x geometry.Point) bool {
	sd.ptcount++
	e, got := sd.locate(x)
	if !got {
		// Did not find the edge using normal walk
		return false
	}

	if ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()) {
		// Point is already in subdivision
		return true
	}

	if quadedge.OnEdge(x, e) {
		e = e.OPrev()
		// Check to see if this point is still alreayd there.
		if ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()) {
			// Point is already in subdivision
			return true
		}
		quadedge.Delete(e.ONext())
	}

	// Connect the new point to the vertices of the containing
	// triangle (or quadrilaterial, if the new point fell on an
	// existing edge.)
	base := quadedge.NewWithEndPoints(e.Orig(), &x)
	quadedge.Splice(base, e)
	sd.startingEdge = base

	base = quadedge.Connect(e, base.Sym())
	e = base.OPrev()
	for e.LNext() != sd.startingEdge {
		base = quadedge.Connect(e, base.Sym())
		e = base.OPrev()
	}

	// Examine suspect edges to ensure that the Delaunay condition
	// is satisfied.
	for {
		t := e.OPrev()
		switch {
		case quadedge.RightOf(*t.Dest(), e) &&
			geometry.InCircle(*e.Orig(), *t.Dest(), *e.Dest(), x):
			quadedge.Swap(e)
			e = e.OPrev()

		case e.ONext() == sd.startingEdge: // no more suspect edges
			return true
		default: // pop a suspect edge
			e = e.ONext().LPrev()
		}
	}
	return true
}

func (sd *Subdivision) InsertConstraint(ctx context.Context, vertexIndex VertexIndex, start, end geometry.Point) (err error) {

	if debug {

		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)

	}
	defer func() {
		if err != nil && err != ErrCoincidentEdges {
			//DumpSubdivision(sd)
			fmt.Printf("starting point %#v\n", start)
			fmt.Printf("end point %#v\n", end)
		}
	}()

	var (
		pu []geometry.Point
		pl []geometry.Point
	)

	if vertexIndex == nil {
		vertexIndex = sd.VertexIndex()
	}

	startingEdge, ok := vertexIndex[start]
	if !ok {
		// start is not in our subdivision
		return errors.New("Invalid starting vertex.")
	}

	if e := startingEdge.FindONextDest(end); e != nil {
		// Nothing to do, edge already in the subdivision.
		return nil
	}

	removalList, err := IntersectingEdges(ctx, startingEdge, end)
	if err != nil && err != ErrCoincidentEdges {
		return err
	}

	pu = append(pu, start)
	pl = append(pl, start)

	for _, e := range removalList {
		if IsHardFrameEdge(sd.frame, e) {
			continue
		}
		for _, spoint := range [2]geometry.Point{*e.Orig(), *e.Dest()} {
			switch c := Classify(spoint, start, end); c {
			case LEFT:
				pl = geometry.AppendNonRepeat(pl, spoint)
			case RIGHT:
				pu = geometry.AppendNonRepeat(pu, spoint)
			default:
				/*
					if debug {
						log.Printf("Classification: %v -- %v, %v, %v",c,spoint, start,end)
						// should not come here.
						return ErrAssumptionFailed()
					}
				*/
				continue
			}
		}
		vertexIndex.Remove(e)
		quadedge.Delete(e)
	}

	pl = geometry.AppendNonRepeat(pl, end)
	pu = geometry.AppendNonRepeat(pu, end)

	for _, pts := range [2][]geometry.Point{pu, pl} {
		if len(pts) == 2 {
			// just a shared line, no points to triangulate.
			continue
		}

		edges, err := triangulatePseudoPolygon(pts)
		if err != nil {
			log.Println("triangulate pseudo polygon fail.")
			return err
		}

		var redoedges []int
		for i, edge := range edges {

			// First we need to check that the edge does not intersect other edges, this can happen if
			// the polygon we are  triangulating happens to be concave. In which case it is possible
			// a triangle outside of the "ok" region, and we should ignore those edges

			// Original code think this is a bug: intersectList, _ := intersectingEdges(startingEdge,end)
			{
				/*
					startingEdge := vertexIndex[edge[0]]

					//intersectList, err := IntersectingEdges(ctx, startingEdge, edge[1])
					intersectList, err := IntersectingEdges(ctx, startingEdge, end)
					if err != nil && err != ErrCoincidentEdges {
						log.Println("failed to insert edge check")
						return err
					}
					// filter out intersects only at the end points.
					count := 0
					for _, iln := range intersectList {
						if geometry.ArePointsEqual(*iln.Orig(), edge[0]) ||
							geometry.ArePointsEqual(*iln.Dest(), edge[0]) ||
							geometry.ArePointsEqual(*iln.Orig(), edge[1]) ||
							geometry.ArePointsEqual(*iln.Dest(), edge[1]) {
							continue
						}
						count++
					}
					if count > 0 {
						if debug {
							debugger.Record(ctx,
								geometry.UnwrapPoint(edge[0]),
								"intersecting line:startPoint",
								"Start Point",
							)
							debugger.Record(ctx,
								geometry.UnwrapPoint(edge[1]),
								"intersecting line:endPoint",
								"End Point",
							)
							debugger.Record(ctx,
								startingEdge.AsGeomLine(),
								"intersecting line:startingedge",
								"StartingEdge %v", startingEdge.AsGeomLine(),
							)
							l := geom.Line{geometry.UnwrapPoint(*startingEdge.Orig()), geometry.UnwrapPoint(edge[1])}
							debugger.Record(ctx,
								l,
								"intersecting line:intersecting",
								"should not fine any intersects with this line. %v ", l,
							)
							for i, il := range intersectList {
								debugger.Record(ctx,
									il.AsGeomLine(),
									"intersecting line:intersected",
									"line %v of %v -- %v", i, len(intersectList), il.AsGeomLine(),
								)
							}

						}
						log.Println("number of intersectlist found", count)
						//return errors.New("Should not get here.")
						continue
					}
				*/
			}

			if err = sd.insertEdge(vertexIndex, edge[0], edge[1]); err != nil {
				if err == ErrDidNotFindToFrom {
					// let's reque this edge
					redoedges = append(redoedges, i)
					continue
				}
				log.Println("Failed to insert edge.")
				return err
			}
		}
		for _, i := range redoedges {
			if err = sd.insertEdge(vertexIndex, edges[i][0], edges[i][1]); err != nil {
				log.Println("Redo Failed to insert edge.", len(redoedges))

				//ignore
				//	return err
			}
		}
	}

	return nil
}

func selectCorrectEdges(from, to *quadedge.Edge) (cfrom, cto *quadedge.Edge) {
	orig := *from.Orig()
	dest := *to.Orig()
	cfrom, cto = from, to
	log.Printf("curr RightOf(dest)? %v", quadedge.RightOf(dest, cfrom))
	log.Printf("destedge.Sym RightOf(orig)? %v", quadedge.RightOf(orig, cto))
	if !quadedge.RightOf(dest, cfrom) {
		cfrom = cfrom.OPrev()
	}
	if !quadedge.RightOf(orig, cto) {
		cto = cto.OPrev()
	}
	return cfrom, cto
}

func resolveEdge(gse *quadedge.Edge, dest geometry.Point) *quadedge.Edge {

	// There aren't any other edges on this vertex.
	if gse == gse.ONext() {
		return gse
	}

	var lre *quadedge.Edge
	se := gse
	curr := se
	for {
		if quadedge.RightOf(dest, curr) {
			if lre == nil {
				// reset our starting edge.
				se = curr
			}
			lre = curr
			curr = curr.ONext()
			if curr == se {
				break
			}
			continue
		}
		// not right of
		if lre == nil {
			// We have not spotted an element right of us
			curr = curr.ONext()
			if curr == se {
				break
			}
			continue
		}
		return lre
	}
	if lre != nil {
		return lre
	}
	return se
}

func findImmediateRightOfEdges(se *quadedge.Edge, dest geometry.Point) (*quadedge.Edge, *quadedge.Edge) {



	// We want the edge immediately left of the dest.

	orig := *se.Orig()
	log.Printf("Looking for orig fo %v to dest of %v", orig, dest)
	curr := se
	for {
		log.Printf("top level looking at: %p (%v -> %v)", curr, *curr.Orig(), *curr.Dest())
		if geometry.ArePointsEqual(*curr.Dest(), dest) {
			// edge already in the system.
			log.Printf("Edge already in system: %p", curr)
			return curr, nil

		}

		// Need to see if the dest Next has the dest.
		for destedge := curr.Sym().ONext(); destedge != curr.Sym(); destedge = destedge.ONext() {
			log.Printf("\t looking at: %p (%v -> %v)", destedge, *destedge.Orig(), *destedge.Dest())
			if geometry.ArePointsEqual(*destedge.Dest(), dest) {
				// found what we are looking for.
				log.Printf("Found the dest! %v -- %p %p", dest, curr, destedge.Sym())

				return selectCorrectEdges(curr, destedge.Sym())
			}
			//log.Println("Next:", *destedge.Orig(), *curr.Sym().Orig(), *curr.Sym().Dest())

		}
		curr = curr.ONext()
		if curr == se {
			break
		}
	}
	return nil, nil
}

func (sd *Subdivision) insertEdge(vertexIndex VertexIndex, start, end geometry.Point) error {
	if vertexIndex == nil {
		vertexIndex = sd.VertexIndex()
	}
	startingedge, ok := vertexIndex[start]
	if !ok {
		// start is not in our subdivision
		return errors.New("Invalid starting vertex.")
	}

	from  := resolveEdge(startingedge,end) 
	// need to check to see if the dest of from.ONext() is the same as end
	if geometry.ArePointsEqual(*from.ONext().Dest(), end) {
		// already in the system.
		return nil
	}
	startingedge, ok = vertexIndex[end]
	if !ok {
		// end is not in our subdivision
		return errors.New("Invalid end vertex.")
	}

	to := resolveEdge(startingedge, start)



	/*
	log.Println("Looking for to and from.")
	// Now let's find the edge that would be ccw to end
	from, to := findImmediateRightOfEdges(startingedge.ONext(), end)
	log.Printf("found for to and from? %p, %p", from, to)
	if from == nil {
		// The nodes are too far away or the line we are trying to
		// insert crosses and already existing line
		return ErrDidNotFindToFrom
	}
	if to == nil {
		// already in the system
		return nil
	}


		ct, err := FindIntersectingTriangle(edge, end)
		if err != nil && err != ErrCoincidentEdges {
			return err
		}
		if ct == nil {
			return errors.New("did not find an intersecting triangle. assumptions broken.")
		}

		from := ct.StartingEdge().Sym()

		symEdge, ok := vertexIndex[end]
		if !ok || symEdge == nil {
			return errors.New("Invalid ending vertex.")
		}

		ct, err = FindIntersectingTriangle(symEdge, start)
		if err != nil && err != ErrCoincidentEdges {
			return err
		}
		if ct == nil {
			return errors.New("sym did not find an intersecting triangle. assumptions broken.")
		}

		to := ct.StartingEdge().OPrev()
	*/
	newEdge := quadedge.Connect(from, to)
	log.Printf("Added edge %p", newEdge)
	vertexIndex.Add(newEdge)
	return nil
}

// WalkAllEdges will call the provided function for each edge in the subdivision. The walk will
// be terminated if the function returns an error or ErrCancel. ErrCancel will not result in
// an error be returned by main function, otherwise the error will be passed on.
func (sd *Subdivision) WalkAllEdges(fn func(e *quadedge.Edge) error) error {

	if sd == nil || sd.startingEdge == nil {
		return nil
	}
	return WalkAllEdges(sd.startingEdge, fn)
}

func (sd *Subdivision) Triangles(includeFrame bool) (triangles [][3]geometry.Point, err error) {

	err = WalkAllTriangleEdges(
		sd.startingEdge,
		func(edges []*quadedge.Edge) error {
			if len(edges) != 3 {
				// skip this edge
				for i, e := range edges {
					log.Printf("got the following edge%v : %v", i,
						wkt.MustEncode(
							geom.Line{
								geometry.UnwrapPoint(*e.Orig()),
								geometry.UnwrapPoint(*e.Dest()),
							},
						),
					)
				}
				return nil
				//	return errors.New("Something Strange!")
			}

			pts := [3]geometry.Point{*edges[0].Orig(), *edges[1].Orig(), *edges[2].Orig()}

			// Do we want to skip because the points are part of the frame and
			// we have been requested not to include triangles attached to the frame.
			if IsFramePoint(sd.frame, pts[:]...) && !includeFrame {
				return nil
			}

			triangles = append(triangles, pts)
			return nil
		},
	)
	return triangles, err
}

func WalkAllEdges(se *quadedge.Edge, fn func(e *quadedge.Edge) error) error {
	if se == nil {
		return nil
	}
	var (
		toProcess quadedge.Stack
		visited   = make(map[*quadedge.Edge]bool)
	)
	toProcess.Push(se)
	for toProcess.Length() > 0 {
		e := toProcess.Pop()
		if visited[e] {
			continue
		}

		if err := fn(e); err != nil {
			if err == ErrCancel {
				return nil
			}
			return err
		}

		sym := e.Sym()

		toProcess.Push(e.ONext())
		toProcess.Push(sym.ONext())

		visited[e] = true
		visited[sym] = true
	}
	return nil
}

// IsFrameEdge indicates if the edge is part of the given frame.
func IsFrameEdge(frame [3]geometry.Point, es ...*quadedge.Edge) bool {
	for _, e := range es {
		o, d := *e.Orig(), *e.Dest()
		of := geometry.ArePointsEqual(o, frame[0]) || geometry.ArePointsEqual(o, frame[1]) || geometry.ArePointsEqual(o, frame[2])
		df := geometry.ArePointsEqual(d, frame[0]) || geometry.ArePointsEqual(d, frame[1]) || geometry.ArePointsEqual(d, frame[2])
		if of || df {
			return true
		}
	}
	return false
}

// IsFrameEdge indicates if the edge is part of the given frame where both vertexs are part of the frame.
func IsHardFrameEdge(frame [3]geometry.Point, e *quadedge.Edge) bool {
	o, d := *e.Orig(), *e.Dest()
	of := geometry.ArePointsEqual(o, frame[0]) || geometry.ArePointsEqual(o, frame[1]) || geometry.ArePointsEqual(o, frame[2])
	df := geometry.ArePointsEqual(d, frame[0]) || geometry.ArePointsEqual(d, frame[1]) || geometry.ArePointsEqual(d, frame[2])
	return of && df
}

func IsFramePoint(frame [3]geometry.Point, pts ...geometry.Point) bool {
	for _, pt := range pts {
		if geometry.ArePointsEqual(pt, frame[0]) ||
			geometry.ArePointsEqual(pt, frame[1]) ||
			geometry.ArePointsEqual(pt, frame[2]) {
			return true
		}
	}
	return false

}

func constructTriangleEdges(
	e *quadedge.Edge,
	toProcess *quadedge.Stack,
	visited map[*quadedge.Edge]bool,
	fn func(edges []*quadedge.Edge) error,
) error {

	if visited[e] {
		return nil
	}

	curr := e
	var triedges []*quadedge.Edge
	for backToStart := false; !backToStart; backToStart = curr == e {

		// Collect edge
		triedges = append(triedges, curr)

		sym := curr.Sym()
		if !visited[sym] {
			toProcess.Push(sym)
		}

		// mark edge as visted
		visited[curr] = true

		// Move the ccw edge
		curr = curr.LNext()
	}
	return fn(triedges)
}

// WalkAllTriangleEdges will walk the subdivision starting from the starting edge (se) and return
// sets of edges that make make a triangle for each face.
func WalkAllTriangleEdges(se *quadedge.Edge, fn func(edges []*quadedge.Edge) error) error {
	if se == nil {
		return nil
	}
	var (
		toProcess quadedge.Stack
		visited   = make(map[*quadedge.Edge]bool)
	)
	toProcess.Push(se)
	for toProcess.Length() > 0 {
		e := toProcess.Pop()
		if visited[e] {
			continue
		}
		err := constructTriangleEdges(e, &toProcess, visited, fn)
		if err != nil {
			if err == ErrCancel {
				return nil
			}
			return err
		}
	}
	return nil
}

func FindIntersectingTriangle(startingEdge *quadedge.Edge, end geometry.Point) (*Triangle, error) {
	var (
		left  = startingEdge
		right *quadedge.Edge
	)

	for {
		right = left.OPrev()

		lc := Classify(end, *left.Orig(), *left.Dest())
		rc := Classify(end, *right.Orig(), *right.Dest())

		if (lc == RIGHT && rc == LEFT) ||
			lc == BETWEEN ||
			lc == DESTINATION ||
			lc == BEYOND {
			return &Triangle{left}, nil
		}

		if lc != RIGHT && lc != LEFT &&
			rc != RIGHT && rc != LEFT {
			return &Triangle{left}, ErrCoincidentEdges
		}
		left = right
		if left == startingEdge {
			// We have walked all around the vertex.
			break
		}

	}
	return nil, nil
}

func IntersectingEdges(ctx context.Context, startingEdge *quadedge.Edge, end geometry.Point) (intersected []*quadedge.Edge, err error) {

	if debug {

		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)

	}

	var (
		start        = startingEdge.Orig()
		tseq         *Triangle
		pseq         geometry.Point
		shared       *quadedge.Edge
		currentPoint = start
	)

	line := geom.Line{geometry.UnwrapPoint(*start), geometry.UnwrapPoint(end)}

	t, err := FindIntersectingTriangle(startingEdge, end)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Println("First Triangle: ", t.AsGeom())
		debugger.Record(ctx,
			t.AsGeom(),
			"FindIntersectingEdges:Triangle:0",
			"First triangle.",
		)
	}

	for !t.IntersectsPoint(end) {
		if tseq, err = t.OppositeTriangle(*currentPoint); err != nil {
			if debug {
				debugger.Record(ctx,
					tseq.AsGeom(),
					"FindIntersectingEdges:Triangle:Opposite",
					"Opposite triangle.",
				)
			}
			return nil, err
		}
		if debug {
			debugger.Record(ctx,
				tseq.AsGeom(),
				"FindIntersectingEdges:Triangle:Opposite",
				"Opposite triangle.",
			)
		}
		shared = t.SharedEdge(*tseq)
		if shared == nil {
			// Should I panic? This is weird.
			return nil, errors.New("did not find shared edge with Opposite Triangle.")
		}
		pseq = *tseq.OppositeVertex(*t)
		switch Classify(pseq, *start, end) {
		case LEFT:
			currentPoint = shared.Orig()
		case RIGHT:
			currentPoint = shared.Dest()
		}
		if _, ok := planar.SegmentIntersect(line, *shared.AsGeomLine()); ok {
			intersected = append(intersected, shared)
		}
		t = tseq
	}
	return intersected, nil

}

func (sd *Subdivision) IsValid(ctx context.Context) bool {
	count := 0
	if debug {

		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)

	}
	_ = sd.WalkAllEdges(func(e *quadedge.Edge) error {
		l := e.AsGeomLine()
		l2 := l.LenghtSquared()
		if l2 == 0 {
			count++
			if debug {
				debugger.Record(ctx,
					l,
					"ZeroLenght:Edge",
					"Line (%p) %v -- %v ", e, l2, l,
				)
			}
		}
		return nil
	})
	log.Println("Count", count)
	return count == 0
}
