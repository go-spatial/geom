
package quadedge

import (
	"fmt"

	"github.com/go-spatial/geom/encoding/wkt"
)

// DebugDumpEdges returns a string with the WKT representation of the
// edges. On error, an error string is returned.
//
// This is intended for debug purposes only.
func (qes *QuadEdgeSubdivision) DebugDumpEdges() string {
	edges := qes.GetEdgesAsMultiLineString()
	edgesWKT, err := wkt.Encode(edges)
	if err != nil {
		return fmt.Sprintf("error formatting as WKT: %v", err)
	}
	return edgesWKT
}

/*
Validate runs a self consistency checks and reports the first error.

This is not part of the original JTS code.
*/
func (qes *QuadEdgeSubdivision) Validate() error {
	// collect a set of all edges
	edgeSet := make(map[*QuadEdge]bool)
	edges := qes.GetEdges()
	for i := range edges {
		if _, ok := edgeSet[edges[i]]; ok == true {
			return fmt.Errorf("edge reported multiple times in subdiv: %v", edges[i])
		}
		if edges[i].IsLive() == false {
			return fmt.Errorf("a deleted edge is still in subdiv: %v", edges[i])
		}
		if edges[i].Sym().IsLive() == false {
			return fmt.Errorf("a deleted edge is still in subdiv: %v", edges[i].Sym())
		}
		edgeSet[edges[i]] = true
	}

	return qes.validateONext()
}

/*
validateONext validates that each QuadEdge's ONext() goes to the next edge that
shares an origin point in CCW order.

This is not part of the original JTS code.
*/
func (qes *QuadEdgeSubdivision) validateONext() error {

	edgeSet := make(map[*QuadEdge]bool)
	edges := qes.GetEdges()
	for _, e := range edges {
		if _, ok := edgeSet[e]; ok == false {
			// if we haven't checked this edge already
			n := e
			for true {
				ccw := n.ONext()
				if n.Orig().Equals(e.Orig()) == false {
					return fmt.Errorf("edge in ONext() doesn't share an origin: between %v and %v", e, n)
				}
				// this will only work if the angles between edges are < 180deg
				// if both edges are frame edges then the CCW rule may not 
				// be easily detectable. (think angles > 180deg)
				if (qes.isFrameEdge(n) == false || qes.isFrameEdge(ccw) == false) && n.Orig().IsCCW(n.Dest(), ccw.Dest()) == false {
					return fmt.Errorf("edges are not CCW, expected %v to be CCW of %v", ccw, n)
				}
				edgeSet[n] = true
				n = ccw
				if (n == e) {
					break
				}
			}
		}
	}

	return nil
}

