// Package quadedge describes a quadedge object used to build up the triangulation
// A quadedge is made up of four directional edges
//
//          DO
//          ^*
//          ||
//    O*----++---->D
//  L D<----++----*O   R
//          ||
//          *V
//          OD
//
//  O represents the Origin
//  D represents the Destination
//
package quadedge

import "sync/atomic"

// QuadEdge describes a quadedge object. Which is made up of four directional edges
type QuadEdge struct {
	initialized bool
	e           [4]Edge
}

// NewQEdge create a new quad edge object
func NewQEdge() *QuadEdge {
	var gidx = atomic.AddUint64(&glbIdx, 1)
	var qe QuadEdge

	qe.e[0].glbIdx, qe.e[1].glbIdx, qe.e[2].glbIdx, qe.e[3].glbIdx = gidx, gidx, gidx, gidx
	qe.e[0].num, qe.e[1].num, qe.e[2].num, qe.e[3].num = 0, 1, 2, 3
	qe.e[0].qe, qe.e[1].qe, qe.e[2].qe, qe.e[3].qe = &qe, &qe, &qe, &qe

	qe.e[0].next = &(qe.e[0])
	qe.e[1].next = &(qe.e[3])
	qe.e[2].next = &(qe.e[2])
	qe.e[3].next = &(qe.e[1])

	qe.initialized = true

	return &qe
}
