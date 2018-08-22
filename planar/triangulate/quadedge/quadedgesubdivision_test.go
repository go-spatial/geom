package quadedge

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
)

func TestNewQuadEdgeSubdivision(t *testing.T) {
	type tcase struct {
		env              geom.Extent
		tolerance        float64
		frameVertex      string
		expectedEnvelope geom.Extent
	}

	fn := func(t *testing.T, tc tcase) {
		uut := NewQuadEdgeSubdivision(tc.env, tc.tolerance)

		if fmt.Sprint(uut.frameVertex) != tc.frameVertex {
			t.Errorf("error, expected %v got %v", tc.frameVertex, uut.frameVertex)
		}
		if uut.GetTolerance() != tc.tolerance {
			t.Errorf("error, expected %v got %v", tc.tolerance, uut.tolerance)
		}
		if uut.GetEnvelope() != tc.expectedEnvelope {
			t.Errorf("error, expected %v got %v", tc.env, uut.GetEnvelope())
		}
	}
	testcases := []tcase{
		{
			env:              geom.Extent{0, 0, 20, 10},
			tolerance:        0.01,
			frameVertex:      "[[10 210] [-200 -200] [220 -200]]",
			expectedEnvelope: geom.Extent{-200, -200, 220, 210},
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestQuadEdgeSubdivisionDelete(t *testing.T) {
	type tcase struct {
		// insert this site
		a, b, c Vertex
		// expecting this neighbor
		expected string
	}

	fn := func(t *testing.T, tc tcase) {
		uut := NewQuadEdgeSubdivision(geom.Extent{0, 0, 20, 10}, 0.01)

		// this should implicitly connect c to a
		e1 := uut.MakeEdge(tc.a, tc.b)
		uut.Connect(uut.startingEdge, e1)
		e2 := uut.MakeEdge(tc.b, tc.c)
		uut.Connect(e2, e1)
		uut.Validate()

		uut.Delete(e2)
		uut.Validate()
		uut.DebugDumpEdges()

		edges := uut.GetEdgesAsMultiLineString()
		edgesWKT, err := wkt.Encode(edges)
		if err != nil {
			t.Fatalf("expected nil got %v", err)
		}

		qe, err := uut.LocateSegment(tc.a, tc.b)
		if err != nil {
			t.Errorf("expected nil got %v", err)
		}

		_, err = uut.LocateSegment(Vertex{100, 1000}, tc.b)
		if err == nil {
			t.Errorf("expected %v got %v", ErrLocateFailure{}, err)
		}

		if qe.Orig().Equals(tc.a) == false || qe.Dest().Equals(tc.b) == false {
			t.Errorf("expected true got false")
		}

		if edgesWKT != tc.expected {
			t.Errorf("expected %v got %v", tc.expected, edgesWKT)
		}
	}
	testcases := []tcase{
		{Vertex{0, 0}, Vertex{5, 5}, Vertex{0, 5}, "MULTILINESTRING ((0 0,5 5),(0 0,0 5))"},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestQuadEdgeSubdivisionGetEdges(t *testing.T) {
	type tcase struct {
		// insert this site
		a, b, c Vertex
		// expecting this neighbor
		expected string
	}

	fn := func(t *testing.T, tc tcase) {
		uut := NewQuadEdgeSubdivision(geom.Extent{0, 0, 20, 10}, 0.01)

		// this should implicitly connect c to a
		e1 := uut.MakeEdge(tc.a, tc.b)
		uut.Connect(uut.startingEdge, e1)
		e2 := uut.MakeEdge(tc.b, tc.c)
		uut.Connect(e2, e1)

		edges := uut.GetEdgesAsMultiLineString()
		edgesWKT, err := wkt.Encode(edges)
		if err != nil {
			t.Fatalf("expected nil got %v", err)
		}

		if edgesWKT != tc.expected {
			t.Errorf("expected %v got %v", tc.expected, edgesWKT)
		}

		// This process does not form a proper triangle.
		tris, err := uut.GetTriangles()
		if err != nil {
			t.Fatalf("expected nil got %v", err)
		}
		trisWKT, err := wkt.Encode(tris)
		if err != nil {
			t.Fatalf("expected nil got %v", err)
		}

		if trisWKT != "MULTIPOLYGON EMPTY" {
			t.Errorf("expected %v got %v", tc.expected, trisWKT)
		}
	}
	testcases := []tcase{
		{Vertex{0, 0}, Vertex{5, 5}, Vertex{0, 5}, "MULTILINESTRING ((0 0,5 5),(0 0,0 5),(0 5,5 5))"},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestQuadEdgeSubdivisionIsOnEdge(t *testing.T) {
	type tcase struct {
		// edge to test
		e1, e2 Vertex
		// point to test
		p        Vertex
		expected bool
	}

	fn := func(t *testing.T, tc tcase) {
		uut := NewQuadEdgeSubdivision(geom.Extent{0, 0, 20, 10}, 0.01)

		// this should implicitly connect c to a
		e1 := uut.MakeEdge(tc.e1, tc.e2)

		onEdge := uut.IsOnEdge(e1, tc.p)

		if onEdge != tc.expected {
			t.Fatalf("expected %v got %v", tc.expected, onEdge)
		}

		onLine := uut.IsOnLine(geom.Line{tc.e1, tc.e2}, tc.p)

		if onLine != tc.expected {
			t.Fatalf("expected %v got %v", tc.expected, onEdge)
		}
	}
	testcases := []tcase{
		{Vertex{0, 0}, Vertex{5, 5}, Vertex{3, 3}, true},
		// a small deviation should still be considered on the edge
		{Vertex{0, 0}, Vertex{5, 5}, Vertex{3 + 1e-5, 3}, true},
		// a slightly larger deviation is not on the edge.
		{Vertex{0, 0}, Vertex{5, 5}, Vertex{3 + 1e-4, 3}, false},
		{Vertex{2, 3}, Vertex{2, 5}, Vertex{2, 3}, true},
		{Vertex{2, 3}, Vertex{2, 5}, Vertex{2, 2}, false},
		{Vertex{2, 3}, Vertex{6, 3}, Vertex{6, 3}, true},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestQuadEdgeSubdivisionIsVertexOfEdge(t *testing.T) {
	type tcase struct {
		// edge to test
		e1, e2 Vertex
		// point to test
		p        Vertex
		expected bool
	}

	fn := func(t *testing.T, tc tcase) {
		uut := NewQuadEdgeSubdivision(geom.Extent{0, 0, 20, 10}, 0.01)

		// this should implicitly connect c to a
		e1 := uut.MakeEdge(tc.e1, tc.e2)

		onEdge := uut.IsVertexOfEdge(e1, tc.p)

		if onEdge != tc.expected {
			t.Fatalf("expected %v got %v", tc.expected, onEdge)
		}
	}
	testcases := []tcase{
		{Vertex{0, 0}, Vertex{5, 5}, Vertex{3, 3}, false},
		{Vertex{2, 3}, Vertex{2, 5}, Vertex{2, 3}, true},
		{Vertex{2, 3}, Vertex{2, 5}, Vertex{2, 2}, false},
		{Vertex{2, 3}, Vertex{6, 3}, Vertex{6, 3}, true},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestQuadEdgeSubdivisionLocate(t *testing.T) {
	type tcase struct {
		// search from this vertex
		from Vertex
		// expecting this neighbor
		expected string
	}

	fn := func(t *testing.T, tc tcase) {
		uut := NewQuadEdgeSubdivision(geom.Extent{0, 0, 20, 10}, 0.01)

		r, err := uut.Locate(tc.from)
		if err != nil {
			t.Fatalf("expected nil got %v", err)
		}
		fmt.Sprintf("%v %v", r.Orig(), r.Dest())
		// if r != tc.expected {
		// 	t.Errorf("expected %v got %v", tc.expected, r)
		// }
	}
	testcases := []tcase{
		{Vertex{0, 0}, "[10 210] [-200 -200]"},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
