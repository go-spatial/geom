package quadedge

import (
	"fmt"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/winding"
)

func testEachSoloSym(e *Edge, dests ...geom.Point) (errs []error) {
	// we will test each  point and then make sure that the sym is only one edge
	for i, pt := range dests {
		deste := e.FindONextDest(pt)
		if deste == nil {
			errs = append(errs, fmt.Errorf("error pt[%v] %v not in edge %v", i, wkt.MustEncode(pt), wkt.MustEncode(e.AsLine())))
			continue
		}
		symdest := deste.Sym()
		if deste == nil {
			errs = append(errs, fmt.Errorf("error pt[%v] %v, edge sym nil", i, wkt.MustEncode(pt)))
			continue
		}
		if !cmp.GeomPointEqual(*symdest.Orig(), pt) {
			errs = append(errs, fmt.Errorf("error pt[%v] %v not origin of edge sym %v", i, wkt.MustEncode(pt), wkt.MustEncode(symdest.AsLine())))
			continue
		}
		// let's make sure that the sym only has one edge in it.
		if symdest.ONext() != symdest {
			errs = append(errs, fmt.Errorf("error pt[%v] %v edge sym onext not same %p vs %p ", i, wkt.MustEncode(pt), symdest, symdest.ONext()))
			continue
		}
		if symdest.OPrev() != symdest {
			errs = append(errs, fmt.Errorf("error pt[%v] %v edge sym oprev not same %p vs %p ", i, wkt.MustEncode(pt), symdest, symdest.OPrev()))
			continue
		}
	}
	return errs
}

func TestSplice(t *testing.T) {
	var (
		// Assume for these tests we are on a coordinate system
		// with y-positive up
		order winding.Order

		err error
	)
	ringA := BuildEdgeGraphAroundPoint(
		geom.Point{0, 0},
		geom.Point{0, -1},
		geom.Point{2, 0},
		geom.Point{0, 1},
		geom.Point{-1, 0},
	)

	ringA = ringA.FindONextDest(geom.Point{2, 0})
	if ringA == nil {
		// Test setup failure
		panic("Failed to find (2,0) edge")
	}

	// Make sure ringA and ringB are valid
	if err = Validate(ringA, order); err != nil {
		er1 := err.(ErrInvalid)
		t.Errorf("ringA valid, expected nil got %v", err)
		for i := range er1 {
			t.Logf("ringA %v : %v", i, er1[i])
		}
	}

	if errs := testEachSoloSym(ringA,
		geom.Point{0, -1},
		geom.Point{2, 0},
		geom.Point{0, 1},
		geom.Point{-1, 0},
	); len(errs) != 0 {
		t.Errorf("ringA not build correctly")
		for i := range errs {
			t.Logf("err: %v", errs[i])
		}
	}

	// test splice with nil, should not modify anything
	Splice(ringA, nil)

	// Make sure ringA and ringB are valid
	if err = Validate(ringA, order); err != nil {
		er1 := err.(ErrInvalid)
		t.Errorf("ringA valid, expected nil got %v", err)
		for i := range er1 {
			t.Logf("ringA %v : %v", i, er1[i])
		}
	}

	if errs := testEachSoloSym(ringA,
		geom.Point{0, -1},
		geom.Point{2, 0},
		geom.Point{0, 1},
		geom.Point{-1, 0},
	); len(errs) != 0 {
		t.Errorf("ringA after nil splice not build correctly")
		for i := range errs {
			t.Logf("err: %v", errs[i])
		}
	}

	ringB := BuildEdgeGraphAroundPoint(
		geom.Point{2, 0},
		geom.Point{2, -1},
		geom.Point{3, 0},
		geom.Point{2, 1},
	)

	ringB = ringB.FindONextDest(geom.Point{2, 1})
	if ringB == nil {
		// Test setup failure
		panic("Failed to find (2,1) edge")
	}

	if err = Validate(ringB, order); err != nil {
		er1 := err.(ErrInvalid)
		t.Errorf("ringB valid, expected nil got %v", err)
		for i := range er1 {
			t.Logf("ringB %v : %v", i, er1[i])
		}

	}

	if errs := testEachSoloSym(ringB,
		geom.Point{2, -1},
		geom.Point{3, 0},
		geom.Point{2, 1},
	); len(errs) != 0 {
		t.Errorf("ringA not build correctly")
		for i := range errs {
			t.Logf("err: %v", errs[i])
		}
	}

	t.Logf("ringA(%p) edges: %v", ringA, ringA.DumpAllEdges())
	t.Logf("ringA.Sym(%p) edges: %v", ringA.Sym(), ringA.DumpAllEdges())
	t.Logf("ringB(%p) edges: %v", ringB, ringB.DumpAllEdges())
	Splice(ringA.Sym(), ringB)

	// Make sure ringA and ringB are valid
	if err = Validate(ringA, order); err != nil {
		er1 := err.(ErrInvalid)
		t.Errorf("ringA valid, expected nil got %v", err)
		t.Logf("ringA : %v", ringA.DumpAllEdges())
		for i := range er1 {
			t.Logf("ringA %v : %v", i, er1[i])
		}
	}

	if err = Validate(ringB, order); err != nil {
		er1 := err.(ErrInvalid)
		t.Errorf("ringB valid, expected nil got %v", err)
		t.Logf("ringB : %v", ringB.DumpAllEdges())
		for i := range er1 {
			t.Logf("ringB %v : %v", i, er1[i])
		}
	}

	if errs := testEachSoloSym(ringA,
		geom.Point{0, -1},
		geom.Point{0, 1},
		geom.Point{-1, 0},
	); len(errs) != 0 {
		t.Errorf("ringA after splice not build correctly")
		for i := range errs {
			t.Logf("err: %v", errs[i])
		}
	}
	if errs := testEachSoloSym(ringB,
		geom.Point{2, -1},
		geom.Point{3, 0},
		geom.Point{2, 1},
	); len(errs) != 0 {
		t.Errorf("ringB after splice not build correctly")
		for i := range errs {
			t.Logf("err: %v", errs[i])
		}
	}

	// Should be able to find on B point 0,0 now
	edgeB := ringB.FindONextDest(geom.Point{0, 0})
	if edgeB == nil {
		t.Errorf("find, expected edgeB got nil")
	}
	edgeA := ringA.FindONextDest(geom.Point{2, 0})
	if edgeB == nil {
		t.Errorf("find, expected edgeA got nil")
	}
	if edgeB != edgeA.Sym() {
		t.Errorf("edgeA.Sym != edgeB")
	}
	if edgeA != edgeB.Sym() {
		t.Errorf("edgeB.Sym != edgeA")
	}

	t.Logf("ringA.   (%p) edges: %v", ringA, ringA.DumpAllEdges())
	t.Logf("ringB.   (%p) edges: %v", ringB, ringB.DumpAllEdges())

	// Test that slice properly splits the quad-edge
	Splice(ringA.Sym(), ringB)

	// Make sure ringA and ringB are valid
	if err = Validate(ringA, order); err != nil {
		er1 := err.(ErrInvalid)
		t.Errorf("ringA valid, expected nil got %v", err)
		t.Logf("ringA : %v", ringA.DumpAllEdges())
		for i := range er1 {
			t.Logf("ringA %v : %v", i, er1[i])
		}

	}

	if errs := testEachSoloSym(ringA,
		geom.Point{0, -1},
		geom.Point{2, 0},
		geom.Point{0, 1},
		geom.Point{-1, 0},
	); len(errs) != 0 {
		t.Errorf("ringA not build correctly")
		for i := range errs {
			t.Logf("err: %v", errs[i])
		}
	}

	if err = Validate(ringB, order); err != nil {
		er1 := err.(ErrInvalid)
		t.Errorf("ringB valid, expected nil got %v", err)
		t.Logf("ringB : %v", ringB.DumpAllEdges())
		for i := range er1 {
			t.Logf("ringB %v : %v", i, er1[i])
		}
	}

	if errs := testEachSoloSym(ringB,
		geom.Point{2, -1},
		geom.Point{3, 0},
		geom.Point{2, 1},
	); len(errs) != 0 {
		t.Errorf("ringB not build correctly")
		for i := range errs {
			t.Logf("err: %v", errs[i])
		}
	}

	// Shouldn't be able to find on B point 0,0 now
	edgeB = ringB.FindONextDest(geom.Point{0, 0})
	if edgeB != nil {
		t.Errorf("find, expected nil got %v", wkt.MustEncode(edgeA.AsLine()))
		t.Logf("edgeA.   (%p) edges: %v", edgeA, edgeA.DumpAllEdges())
		t.Logf("ringA.   (%p) edges: %v", ringA, ringA.DumpAllEdges())
		t.Logf("ringB.   (%p) edges: %v", ringB, ringB.DumpAllEdges())
		t.Logf("edgeB.   (%p) edges: %v", edgeB, edgeB.DumpAllEdges())
		t.Logf("edgeA.Sym(%p) edges: %v", edgeA.Sym(), edgeA.Sym().DumpAllEdges())
	}

}

func TestConnect(t *testing.T) {
	type tcase struct {
		Name string
		a    *Edge
		b    *Edge
		err  ErrInvalid
	}

	order := winding.Order{
		YPositiveDown: false,
	}

	fn := func(tc tcase) (string, func(*testing.T)) {
		return tc.Name, func(t *testing.T) {
			var err error
			if err = Validate(tc.a, order); err != nil {
				if verr, ok := err.(ErrInvalid); ok {
					for i, estr := range verr {
						t.Logf("%03v: %v", i, estr)
					}
				}
				t.Errorf("validate on a: expected nil got %v", err)
				return
			}
			if err = Validate(tc.b, order); err != nil {
				if verr, ok := err.(ErrInvalid); ok {
					for i, estr := range verr {
						t.Logf("%03v: %v", i, estr)
					}
				}
				t.Errorf("validate on b: expected nil got %v", err)
				return
			}

			aOrig := *tc.a.Orig()
			bOrig := *tc.b.Orig()
			tc.a, _ = ResolveEdge(order, tc.a, bOrig)
			tc.b, _ = ResolveEdge(order, tc.b, aOrig)
			t.Logf("Connecting a:%v to b: %v", wkt.MustEncode(tc.a.AsLine()), wkt.MustEncode(tc.b.AsLine()))

			e := Connect(tc.a, tc.b, order)

			if err = Validate(tc.a, order); err != nil {
				if verr, ok := err.(ErrInvalid); ok {
					for i, estr := range verr {
						t.Logf("%03v: %v", i, estr)
					}
				}
				t.Errorf("validate on a: expected nil got %v", err)
				return
			}
			if err = Validate(tc.b, order); err != nil {
				if verr, ok := err.(ErrInvalid); ok {
					for i, estr := range verr {
						t.Logf("%03v: %v", i, estr)
					}
				}
				t.Errorf("validate on b: expected nil got %v", err)
				return
			}
			if err = Validate(e, order); err != nil {
				if verr, ok := err.(ErrInvalid); ok {
					for i, estr := range verr {
						t.Logf("%03v: %v", i, estr)
					}
				}
				t.Errorf("validate on e: expected nil got %v", err)
				return
			}
			t.Logf("a edges: %v", tc.a.DumpAllEdges())
			t.Logf("b edges: %v", tc.b.DumpAllEdges())
			t.Logf("e edges: %v", e.DumpAllEdges())

		}
	}

	tests := []tcase{
		{
			Name: "Simple",
			a: BuildEdgeGraphAroundPoint(
				geom.Point{0, 0},
				geom.Point{0, 1},
				geom.Point{-1, 0},
				geom.Point{0, -1},
			),
			b: BuildEdgeGraphAroundPoint(
				geom.Point{2, 0},
				geom.Point{2, 1},
				geom.Point{2, -1},
				geom.Point{3, 0},
			),
		},
	}

	for i := range tests {
		t.Run(fn(tests[i]))
	}
}
