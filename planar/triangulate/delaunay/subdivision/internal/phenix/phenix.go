package phenix

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/subdivision"
	"github.com/go-spatial/geom/winding"
)

type LabeledPoint struct {
	Point *geom.Point
	Label string
}

type PointBag struct {
	TrianglePoints [3]geom.Point
	Points         []geom.Point
}

func (pb *PointBag) LabeledPoints(lbls ...string) []LabeledPoint {
	lbpts := make([]LabeledPoint, len(lbls))
	for i := range lbls {
		lbpts[i] = pb.LabeledPoint(lbls[i])
	}
	return lbpts
}

func (pb *PointBag) LabeledPoint(lbl string) LabeledPoint {
	switch lbl {
	case "t0":
		return LabeledPoint{Label: lbl, Point: &(pb.TrianglePoints[0])}
	case "t1":
		return LabeledPoint{Label: lbl, Point: &(pb.TrianglePoints[1])}
	case "t2":
		return LabeledPoint{Label: lbl, Point: &(pb.TrianglePoints[2])}
	}
	// assume 0 is p
	i, err := strconv.Atoi(lbl[1:])
	if err != nil {
		panic(fmt.Sprintf("expected int for '%v' %v", lbl, err))
	}
	return LabeledPoint{Label: lbl, Point: &(pb.Points[i])}
}

func (pb *PointBag) NewPhenix(name string, checkFn func(*PointBag) []Check) *Phenix {
	checks := checkFn(pb)
	return NewPhenix(name, pb.TrianglePoints, pb.Points, checks...)
}

func (pb *PointBag) NewEDefinition(desc string) (edef EDefinition) {
	// expect desc to be in one of two forms
	// t0 : t1 t2
	// or
	// e0 = t1.onext.onext.sym : t1 t2
	parts := strings.Split(desc, " ")
	// first part is always the name
	if len(parts) <= 0 {
		panic("expected proper desc")
	}
	edef.Name = parts[0]
	parts = parts[1:]
	if parts[0] == "=" {
		// we have a def
		defs := strings.Split(parts[1], ".")
		if len(defs) <= 0 {
			panic("expected defs to have something")
		}
		edef.BaseName = defs[0]
		for _, opt := range defs[1:] {
			switch opt {
			case "onext":
				edef.Ops = append(edef.Ops, quadedge.ONext)
			case "sym":
				edef.Ops = append(edef.Ops, quadedge.Sym)
			default:
				panic(fmt.Sprintf("Unsupported operation %v", opt))
			}
		}
		parts = parts[2:]
	}
	// assume parts[0] == ':'
	edef.Points = pb.LabeledPoints(parts[1:]...)
	return edef
}

func (pb *PointBag) NewCheck(debug bool, defs ...string) Check {
	edges := make([]EDefinition, len(defs))
	for i := range defs {
		edges[i] = pb.NewEDefinition(defs[i])
	}
	return NewCheck(debug, edges...)
}

func edgeCount(e *quadedge.Edge) (c int) {
	if e == nil {
		return 0
	}
	for ne := e.ONext(); ne != e; ne, c = ne.ONext(), c+1 {
	}
	return c + 1
}
func testEdgeONextOPrevDest1(t *testing.T, lbl string, e *quadedge.Edge, pts ...LabeledPoint) bool {
	t.Helper()
	printedEdge := false
	ln := len(pts) - 1
	ne, pe := e.ONext(), e.OPrev()
	for i := range pts {
		j := ln - i
		if !cmp.GeomPointEqual(*(ne.Dest()), *pts[i].Point) {
			t.Logf("edge %v", wkt.MustEncode(e.AsLine()))
			if !printedEdge {
				printedEdge = true
			}
			t.Errorf("edge %v onext [%v], expected %v [%v] got %v", lbl, i, *pts[i].Point, pts[i].Label, *(ne.Dest()))
			return false
		}
		if !cmp.GeomPointEqual(*(pe.Dest()), *pts[j].Point) {
			if !printedEdge {
				t.Logf("edge %v", wkt.MustEncode(e.AsLine()))
			}
			t.Errorf("edge %v oprev [%v], expected %v [%v] got %v", lbl, j, *pts[i].Point, pts[i].Label, *(pe.Dest()))
			return false
		}
		ne, pe = ne.ONext(), pe.OPrev()
	}
	return true
}

func logEdgeDest(t *testing.T, e *quadedge.Edge, total int) {
	//t.Helper()
	if e == nil {
		return
	}
	padding := int(math.Log10(float64(total)))
	t.Logf("%0*v: %v", padding, 0, wkt.MustEncode(*e.Dest()))
	for ne, c := e.ONext(), 1; ne != e; ne, c = ne.ONext(), c+1 {
		t.Logf("%0*v: %v", padding, c, wkt.MustEncode(*ne.Dest()))
	}
}
func checkEdge(t *testing.T, PrintOutEdge bool, lbl string, e *quadedge.Edge, pts []LabeledPoint) bool {
	//t.Helper()
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
	BaseName string
	Ops      []quadedge.Operation
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

	edef.edge = c.edgeForName(edef.BaseName).Apply(edef.Ops...)
	return edef.edge
}

func (c *Check) edges(t *testing.T) bool {
	t.Helper()
	for _, name := range c.edgeNames {
		edef := c.edgeMap[name]
		if !checkEdge(t, c.Debug, name, c.edgeForName(name), edef.Points) {
			return false
		}
	}
	return true
}

func (c *Check) Test(t *testing.T) bool {

	t.Helper()
	for _, name := range c.edgeNames {
		edef := c.edgeMap[name]
		edge := c.edgeForName(name)
		cont := checkEdge(t, c.Debug, name, edge, edef.Points)
		if !cont {
			return false
		}
	}
	return true
}

func NewCheck(debug bool, edges ...EDefinition) Check {
	edgMap := make(map[string]*EDefinition, len(edges))
	edgNames := make([]string, 0, len(edges))
	for i := range edges {
		edgMap[edges[i].Name] = &edges[i]
		edgNames = append(edgNames, edges[i].Name)
	}
	return Check{
		Debug:     debug,
		edgeMap:   edgMap,
		edgeNames: edgNames,
	}
}

type Phenix struct {
	Name             string
	Line             int
	Filename         string
	Subdivision      *subdivision.Subdivision
	TriangleEdge     [3]*quadedge.Edge
	TrianglePoints   [3]geom.Point
	Points           []geom.Point
	Checks           []Check
	Order            winding.Order
	VerifyCheckCount bool
}

func genAndTestNewSD(t *testing.T, order winding.Order, trianglePoint [3]geom.Point) (*subdivision.Subdivision, [3]*quadedge.Edge) {
	t.Helper()
	var triangleEdge [3]*quadedge.Edge
	sd := subdivision.New(order, trianglePoint[0], trianglePoint[1], trianglePoint[2])
	se := sd.StartingEdge()
	if !cmp.GeomPointEqual(*(se.Orig()), trianglePoint[0]) {
		se = se.FindONextDest(trianglePoint[0]).Sym()
	}

	triangleEdge[0] = se.FindONextDest(trianglePoint[2])
	triangleEdge[1] = triangleEdge[0].ONext().Sym()
	triangleEdge[2] = triangleEdge[0].Sym()

	for i := range trianglePoint {
		// verify that point edge is origin is correct
		if !cmp.GeomPointEqual(*(triangleEdge[i].Orig()), trianglePoint[i]) {
			t.Errorf("new edge %v origin, expected %v got %v", i, wkt.MustEncode(trianglePoint[i]), wkt.MustEncode(*(triangleEdge[i].Orig())))
			return nil, triangleEdge
		}
		// Let's verify that all vertexs have only two edges
		for i := range triangleEdge {
			if count := edgeCount(triangleEdge[i]); count != 2 {
				t.Errorf("vertex %v edges, expected 2, got %v", i, count)
			}
		}
	}
	return sd, triangleEdge
}

func (p *Phenix) Test(t *testing.T) {
	t.Helper()
	if p == nil {
		return
	}

	p.Subdivision, p.TriangleEdge = genAndTestNewSD(t, p.Order, p.TrianglePoints)
	if p.Subdivision == nil {
		return
	}
	for i := range p.TriangleEdge {
		t.Logf("%v: %v", i, p.TriangleEdge[i].DumpAllEdges())
	}

	for i := range p.Points {
		t.Logf("Inserting Point p%d : %v", i, p.Points[i])
		if !p.Subdivision.InsertSite(p.Points[i]) {
			t.Log("failed to insert point")
		}
		if i < len(p.Checks) {
			// Fill in the t0,t1,t2 edges for each check
			if p.Checks[i].edgeMap["t0"] == nil {
				panic(
					fmt.Sprintf("Test at %v:%v did not define check for t0", p.Filename, p.Line),
				)
			}
			if p.Checks[i].edgeMap["t1"] == nil {
				panic(
					fmt.Sprintf("Test at %v:%v did not define check for t1", p.Filename, p.Line),
				)
			}
			if p.Checks[i].edgeMap["t2"] == nil {
				panic(
					fmt.Sprintf("Test at %v:%v did not define check for t2", p.Filename, p.Line),
				)
			}
			p.Checks[i].edgeMap["t0"].edge = p.TriangleEdge[0]
			p.Checks[i].edgeMap["t1"].edge = p.TriangleEdge[1]
			p.Checks[i].edgeMap["t2"].edge = p.TriangleEdge[2]
			if !p.Checks[i].Test(t) {
				t.Log(p.Subdivision.StartingEdge().DumpAllEdges())
				return
			}
		} else if p.VerifyCheckCount {
			t.Errorf("number of checks, expected %v, got %v", len(p.Points), len(p.Checks))
			return
		}
	}
}

func NewPhenix(name string, trianglePoints [3]geom.Point, points []geom.Point, checks ...Check) *Phenix {
	var (
		ok   bool
		file string
		line int
	)

	// Get the caller info
	if _, file, line, ok = runtime.Caller(1); ok {
		// let's get the short version of the file.
		file = filepath.Base(file)
	} else {
		file = "???"
		line = 0
	}
	p := &Phenix{
		Name:           name,
		Filename:       file,
		Line:           line,
		Points:         points,
		TrianglePoints: trianglePoints,
		Checks:         checks,
	}
	for i := range p.Checks {
		p.Checks[i].parent = p
	}
	return p

}
