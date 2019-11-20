package quadedge

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/winding"
)

func TestSplice(t *testing.T) {
	type tcase struct {
		Desc string
		a    *Edge
		b    *Edge
		err  ErrInvalid
	}
	order := winding.Order{
		YPositiveDown: true,
	}
	fn := func(tc tcase) (string, func(*testing.T)) {
		return tc.Desc, func(t *testing.T) {
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

			tc.a, _ = ResolveEdge(tc.a, *tc.b.Orig())
			tc.b, _ = ResolveEdge(tc.b, *tc.a.Orig())
			t.Logf("Splicing a:%v to b: %v", wkt.MustEncode(tc.a.AsLine()), wkt.MustEncode(tc.b.AsLine()))

			Splice(tc.a.Sym(), tc.b)

			if err := Validate(tc.a, order); err != nil {
				if verr, ok := err.(ErrInvalid); ok {
					for i, estr := range verr {
						t.Logf("%03v: %v", i, estr)
					}
					t.Logf("edges: %v", tc.a.DumpAllEdges())
				}
				t.Errorf("after splice validate on a: expected nil got %v", err)
			}
			if err := Validate(tc.b, order); err != nil {
				if verr, ok := err.(ErrInvalid); ok {
					for i, estr := range verr {
						t.Logf("%03v: %v", i, estr)
					}
					t.Logf("edges: %v", tc.b.DumpAllEdges())
				}
				t.Errorf("after splice validate on b: expected nil got %v", err)
			}
			t.Logf("a edges: %v", tc.a.DumpAllEdges())
			t.Logf("b edges: %v", tc.b.DumpAllEdges())

		}
	}

	tests := []tcase{
		{
			Desc: "Simple",
			a: BuildEdgeGraphAroundPoint(
				geom.Point{0, 0},
				geom.Point{0, -1},
				geom.Point{2, 0},
				geom.Point{0, 1},
				geom.Point{-1, 0},
			),
			b: BuildEdgeGraphAroundPoint(
				geom.Point{2, 0},
				geom.Point{2, -1},
				geom.Point{3, 0},
				geom.Point{2, 1},
			),
		},
	}
	for _, tc := range tests {
		t.Run(fn(tc))
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
		YPositiveDown: true,
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

			tc.a, _ = ResolveEdge(tc.a, *tc.b.Orig())
			tc.b, _ = ResolveEdge(tc.b, *tc.a.Orig())
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
		// Subtests
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
