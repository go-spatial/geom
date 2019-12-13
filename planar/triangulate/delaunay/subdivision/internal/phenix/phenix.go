package phenix

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/quadedge"
)

type LabeledPoint struct {
	Point *geom.Point
	Label string
}

func edgeCount(e *quadedge.Edge) (c int) {
	if e == nil {
		return 0
	}
	for ne := e.ONext(); ne != e; ne, c = ne.ONext(), c+1 {
	}
	return c + 1
}
func testEdgeONextOPrevDest1(t *testing.T, lbl string, e *quadedge.Edge, pts ...geom.Point) bool {
	printedEdge := false
	ln := len(pts) - 1
	ne, pe := e.ONext(), e.OPrev()
	for i := range pts {
		j := ln - i
		if !cmp.GeomPointEqual(*(ne.Dest()), pts[i]) {
			t.Logf("edge %v", wkt.MustEncode(e.AsLine()))
			if !printedEdge {
				printedEdge = true
			}
			t.Errorf("edge %v onext [%v], expected %v got %v", lbl, i, pts[i], *(ne.Dest()))
			return false
		}
		if !cmp.GeomPointEqual(*(pe.Dest()), pts[j]) {
			if !printedEdge {
				t.Logf("edge %v", wkt.MustEncode(e.AsLine()))
			}
			t.Errorf("edge %v oprev [%v], expected %v got %v", lbl, j, pts[j], *(pe.Dest()))
			return false
		}
		ne, pe = ne.ONext(), pe.OPrev()
	}
	return true
}

func logEdgeDest(t *testing.T, e *quadedge.Edge, total int) {
	if e == nil {
		return
	}
	padding := int(math.Log10(float64(total)))
	t.Logf("%0*v: %v", padding, 0, wkt.MustEncode(*e.Dest()))
	for ne, c := e.ONext(), 1; ne != e; ne, c = ne.ONext(), c+1 {
		t.Logf("%0*v: %v", padding, c, wkt.MustEncode(*ne.Dest()))
	}
}
func checkEdge(t *testing.T, lbl string, e *quadedge.Edge, pts []LabeledPoint) bool {
	if e == nil {
		t.Errorf("edge %v, expected edge got nil", lbl)
		return false
	}

	if PrintOutEdge {
		fmt.Printf(`
edge %v
%v
%v

`,
			lbl,
			wkt.MustEncode(*e.Orig()),
			wkt.MustEncode(e.AsLine()),
		)
		for ne, c := e.ONext(), 1; ne != e; ne, c = ne.ONext(), c+1 {
			fmt.Printf("%v\n", wkt.MustEncode(ne.AsLine()))
		}
		fmt.Printf("\n")
	}

	if count := edgeCount(e); count != len(pts)+1 {
		t.Errorf("edge %v :vertex %v edges, expected %v, got %v", lbl, wkt.MustEncode(*e.Orig()), len(pts)+1, count)
		logEdgeDest(t, e, count)
		return false
	}

	return testEdgeONextOPrevDest1(t, lbl, e, pts...)

}

type EDefinition struct {
	Name     string
	Points   []LabeledPoint
	edge     *quadedge.Edge
	baseName string
	ops      []quadedge.Operation
}

type Check struct {
	parent    *Phenix
	edgeMap   map[string]*EDefinition
	edgeNames []string
	Debug     bool
}

func (c *Check) edgeForName(name string) *quadedge.Edge {
	if c == nil {
		return nil
	}
	edef := c.edgeMap[name]
	if edef == nil {
		return nil
	}
	if edef.edge != nil {
		return edef.edge
	}

	edef.edge = c.edgeForName(edef.baseName).Apply(edef.ops...)
	return edef.edge
}

func (c *Check) edges(t *testing.T) bool {
	for i, name := range c.edgeNames {
		edef := c.edgeMap[name]
		if !checkEdge(t, c.edgeForName(name), edef.Points) {
			return false
		}
	}
	return true
}

type Phenix struct {
	Name         string
	Line         string
	Filename     string
	TriangleEdge [3]*quadedge.Edge
	Points       []geom.Point
	Checks       []Check
}

func (p *Phenix) Name() string {

}
func (p *Phenix) Test(t *testing.T) {

}
