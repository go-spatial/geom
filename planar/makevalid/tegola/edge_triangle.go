package tegola

import (
	"context"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/constraineddelaunay"
)

func sortedEdge(pt1, pt2 [2]float64) [2][2]float64 {
	if cmp.PointLess(pt1, pt2) {
		return [2][2]float64{pt1, pt2}
	}
	return [2][2]float64{pt2, pt1}
}

func edgeMapFromTriangles(triangles ...triangle) map[[2][2]float64][]int {
	// an edge can have at most two triangles associated with it.
	em := make(map[[2][2]float64][]int, 2*len(triangles))
	for i, tri := range triangles {
		for _, edg := range tri.SortedEdges() {
			if _, ok := em[edg]; !ok {
				em[edg] = make([]int, 0, 2)
			}
			em[edg] = append(em[edg], i)
		}
	}
	return em
}

func triangulateGeometry(g geom.Geometry) (geom.MultiPolygon, error) {
	uut := new(constraineddelaunay.Triangulator)
	uut.InsertSegments(g)

	if debug {
		err := uut.Validate()
		if err != nil {
			log.Printf("Triangulator is not validate for the given segments %v : %v", g, err)
			return nil, err
		}
	}
	// TODO(gdey): We need to insure that GetTriangles does not dup the first point to the
	//              last point.
	return uut.GetTriangles()
}

func newEdgeIndexTriangles(ctx context.Context, hm planar.HitMapper, g geom.Geometry) (*edgeIndexTriangles, error) {

	var geomTriangles geom.MultiPolygon
	var err error

	if geomTriangles, err = triangulateGeometry(g); err != nil {
		return nil, err
	}

	triangles := make([]triangle, 0, len(geomTriangles))

	for _, ply := range geomTriangles {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		tri := newTriangleFromPolygon(ply)
		if hm.LabelFor(tri.Center()) == planar.Outside {
			continue
		}
		triangles = append(triangles, tri)
	}
	if debug {
		log.Printf("Got the following triangles: %v", triangles)
	}

	t := edgeIndexTriangles{
		triangles: make([]triangle, len(triangles)),
		edgeMap:   edgeMapFromTriangles(triangles...),
	}
	copy(t.triangles, triangles)
	return &t, nil
}

type edgeIndexTriangles struct {
	triangles []triangle
	edgeMap   map[[2][2]float64][]int
}

func (eit *edgeIndexTriangles) MultiPolygon(ctx context.Context) (mplyg [][][][2]float64) {
	if eit == nil {
		return mplyg
	}
	seen := make(map[int]bool, len(eit.triangles))
	for i := range eit.triangles {
		if ctx.Err() != nil {
			return nil
		}
		if seen[i] {
			continue
		}
		seen[i] = true
		plyg := eit.PolygonForTriangle(ctx, i, seen)
		if len(plyg) > 0 {
			mplyg = append(mplyg, plyg)
		}
	}
	return mplyg
}

func (eit *edgeIndexTriangles) indexForEdge(p1, p2 [2]float64, defaultIdx int, seen map[int]bool) (idx int, ok bool) {
	for _, idx := range eit.edgeMap[sortedEdge(p1, p2)] {
		if seen[idx] || idx == defaultIdx {
			continue
		}
		return idx, true
	}
	return defaultIdx, false
}

// ringForTriangle will walk the set of triangles starting at the given triangle index. As it walks the triangles it will
// mark them as seen on the seen map. The function will return the outside ring of the walk
func (eit *edgeIndexTriangles) ringForTriangle(ctx context.Context, idx int, seen map[int]bool) (rng [][2]float64) {

	var ok bool

	if debug {
		log.Printf("Getting ring for Triangle %v", idx)
	}

	seen[idx] = true

	// This tracks the start of the ring.
	// The segment we are adding a point will be between the endpoint and the begining of the ring.
	// This track the original begining of the ring.
	var headIdx int

	rng = append(rng, eit.triangles[idx][:]...)
	cidxs := []int{idx, idx, idx}
	cidx := cidxs[len(cidxs)-1]

	if debug {
		log.Printf("Triangles: Starting at %v", idx)
		for i := range eit.triangles {
			log.Printf("Triangle(%v) %v", i, eit.triangles[i])
		}
		log.Printf("EdgeMap")
		for k, v := range eit.edgeMap {
			log.Printf("Edge %v  => %v", k, v)
		}
	}

RING_LOOP:
	for {
		// A few sanity checks, were we cancelled, or reached the end of our walk.
		if ctx.Err() != nil || // We were told to come home.
			headIdx >= len(rng) || len(cidxs) == 0 { // we returned home.
			return rng
		}

		if debug {
			log.Printf("headIdx: %v -- len(rng): %v", headIdx, len(rng))
			log.Printf("ring: %v | %v", rng[:headIdx], rng[headIdx:])
			log.Printf("cidxs: %v", cidxs)
		}

		if cidx, ok = eit.indexForEdge(rng[0], rng[len(rng)-1], cidxs[len(cidxs)-1], seen); !ok {
			// We don't have a neighbor to walk to here. Let's move back one and see if there is a path we need to go down.
			headIdx += 1
			lpt := rng[len(rng)-1]
			copy(rng[1:], rng)
			rng[0] = lpt
			cidxs = cidxs[:len(cidxs)-1]
			continue
		}

		if cidx == idx {
			// We go back to our starting triangle. We need to stop.
			return rng
		}

		if debug {
			log.Printf("Check to see if we have seen the triangle we are going to jump to.")
		}

		// Check to see if we have reached the triangle before.
		for i, pcidx := range cidxs {
			if pcidx != cidx {
				continue
			}
			if debug {
				log.Printf("We have encountered idx (%v) before at %v", cidx, i)
			}
			// need to move all the points over
			tlen := len(rng) - (i + 1)
			tpts := make([][2]float64, tlen)
			copy(tpts, rng[i+1:])
			copy(rng[tlen:], rng[:i+1])
			copy(rng, tpts)
			headIdx += tlen

			cidxs = cidxs[:i+1]
			continue RING_LOOP
		}

		rng = append(rng, eit.triangles[cidx].ThirdPoint(rng[0], rng[len(rng)-1]))

		cidxs[len(cidxs)-1] = cidx
		cidxs = append(cidxs, cidx)
		seen[cidx] = true

	} // for loop
	return rng
}

// polygonForRing returns a polygon for the given ring, this will destroy the ring.
func (eit *edgeIndexTriangles) polygonForRing(ctx context.Context, rng [][2]float64) (plyg [][][2]float64) {
	if debug {
		log.Printf("Turning ring into polygon.")
	}
	if len(rng) == 0 {
		return nil
	}

	// normalize ring
	cmp.RotateToLeftMostPoint(rng)
	if len(rng) <= 3 {
		return [][][2]float64{rng}
	}

	pIdx := func(i int) int {
		if i == 0 {
			return len(rng) - 1
		}
		return i - 1
	}
	nIdx := func(i int) int {
		if i == len(rng)-1 {
			return 0
		}
		return i + 1
	}

	// Allocate space for the initial ring.
	plyg = make([][][2]float64, 1)

	// Remove bubbles. There are two types of bubbles we have to look for.
	// 1. ab … bc, in which case we need to hold on to b.
	//    It is possible that b is absolutely not necessary. It could lie on the line between a and c, in which case
	//    we should remove the extra point.
	// 2. ab … ba, which case we do not need to have b in the ring.

	// let's build an index of where the points that we are walking are. That way when we encounter the same
	// point we are able to “jump” to that point.
	ptIndex := map[[2]float64]int{}
	var (
		ok              bool
		pidx, idx, nidx int
	)

	// Let's walk the points
	for i := 0; i < len(rng); i++ {
		// Context has been cancelled.
		if ctx.Err() != nil {
			return nil
		}
		pt := rng[i]

		// check to see if we have already seen this point.
		if idx, ok = ptIndex[pt]; !ok {
			ptIndex[rng[i]] = i
			continue
		}

		if debug {
			log.Printf("Ring %v", rng)
			log.Printf("Found Bubble at %v", i)
		}

		// We need to figure out which type of bubble this is.
		pidx, nidx = pIdx(idx), nIdx(i)

		if debug {
			log.Printf("Check to see what kind of bubble: %v : %v -- %v : %v ", pidx, idx, i, nidx)
			log.Printf("Prev Pt: %v    Next Pt: %v", rng[pidx], rng[nidx])

		}

		// ab…ba ring. So we need to remove all the way to a.
		if nidx != pidx && cmp.PointEqual(rng[pidx], rng[nidx]) {
			if debug {
				log.Printf("bubble type ab … ba")
				log.Printf("pidx: %v idx: %v i: %v nidx: %v", pidx, idx, i, nidx)
			}

			// Cuts are in counter-clockwise order.
			for j := idx; j <= i; j++ {
				if debug {
					log.Printf("Adding Point(%v) to cut %v", j, rng[j])
				}
				delete(ptIndex, rng[j])
			}
			// These may have wrapped; so don't range with them.
			delete(ptIndex, rng[nidx])
			delete(ptIndex, rng[pidx])

			sliver := cut(&rng, pidx, nidx)
			// remove ther start ab
			sliver = sliver[2:]
			{ // make a copy to free up memory.
				ps := make([][2]float64, len(sliver))
				copy(ps, sliver)
				cmp.RotateToLeftMostPoint(ps)
				plyg = append(plyg, ps)
			}

			i = pidx
			if i >= len(rng) {
				i = 0
			}
			ptIndex[rng[i]] = i
			continue
		}

		// ab … bc
		if debug {
			log.Printf("bubble type ab…bc")
		}
		// Do a quick check to see if b is on ac
		removeB := planar.IsPointOnLine(rng[i], rng[pidx], rng[nidx])

		// Cuts are in counter-clockwise order.
		for j := idx; j <= i-1; j++ {
			delete(ptIndex, rng[j])
		}

		sliver := cut(&rng, idx, i)
		cmp.RotateToLeftMostPoint(sliver)
		plyg = append(plyg, sliver)

		if removeB {
			cut(&rng, idx, idx+1)
		}

		i = idx
		ptIndex[rng[i]] = i
	}
	plyg[0] = make([][2]float64, len(rng))
	copy(plyg[0], rng)
	return plyg
}

func (eit *edgeIndexTriangles) PolygonForTriangle(ctx context.Context, idx int, seen map[int]bool) (plyg [][][2]float64) {

	if debug {
		log.Printf("Polygon For Triangle: Started at %v", idx)
	}
	// Get the external ring for the given triangle.
	rng := eit.ringForTriangle(ctx, idx, seen)
	if debug {
		log.Printf("Got back ring: %v", rng)
	}
	plyg = eit.polygonForRing(ctx, rng)

	return plyg
}
