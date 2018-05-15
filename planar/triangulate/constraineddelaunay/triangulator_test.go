
package constraineddelaunay

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkb"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar/triangulate"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

func TestFindContainingTriangle(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
    // A simple website for performing conversions:
    // https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB      string
		searchFrom		geom.Line
		expectedEdge  string
	}

	fn := func(t *testing.T, tc tcase) {
		bytes, err := hex.DecodeString(tc.inputWKB)
		if err != nil {
			t.Fatalf("error decoding hex string: %v", err)
			return
		}
		g, err := wkb.DecodeBytes(bytes)
		if err != nil {
			t.Fatalf("error decoding WKB: %v", err)
			return
		}

		uut := new(Triangulator)
		uut.tolerance = 1e-6
		uut.insertSites(g)

		tri, err := uut.findContainingTriangle(triangulate.NewSegment(tc.searchFrom))
		if err != nil {
			t.Fatalf("error, expected nil got %v", err)
			return
		}
		qeStr := fmt.Sprintf("%v -> %v", tri.qe.Orig(), tri.qe.Dest())
		if qeStr != tc.expectedEdge {
			t.Fatalf("error, expected %v got %v", tc.expectedEdge, qeStr)
		}

	}
	testcases := []tcase{
		{
			inputWKT:      `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:      `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom: geom.Line{{0,0}, {10, 10}},
			expectedEdge: `[0 0] -> [0 10]`,
		},
		{
			inputWKT:      `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:      `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom: geom.Line{{10,0}, {0, 20}},
			expectedEdge: `[10 0] -> [0 10]`,
		},
		{
			inputWKT:      `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:      `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom: geom.Line{{10,10}, {0, 0}},
			expectedEdge: `[10 10] -> [10 0]`,
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestDeleteEdge(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
	    // A simple website for performing conversions:
	    // https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB      string
		deleteMe		geom.Line
	}

	fn := func(t *testing.T, tc tcase) {
		bytes, err := hex.DecodeString(tc.inputWKB)
		if err != nil {
			t.Fatalf("error decoding hex string, expected nil got %v", err)
			return
		}
		g, err := wkb.DecodeBytes(bytes)
		if err != nil {
			t.Fatalf("error decoding WKB, expected nil got %v", err)
			return
		}

		uut := new(Triangulator)
		uut.tolerance = 1e-6
		uut.InsertSegments(g)
		e, err := uut.LocateSegment(quadedge.Vertex(tc.deleteMe[0]), quadedge.Vertex(tc.deleteMe[1]))
		if err != nil {
			t.Fatalf("error locating segment, expected nil got %v", err)
			return
		}

		err = uut.Validate()
		if err != nil {
			t.Errorf("error validating triangulation, expected nil got %v", err)
			return
		}

		uut.deleteEdge(e)
		err = uut.Validate()
		if err != nil {
			t.Errorf("error validating triangulation after delete, expected nil got %v", err)
			return
		}

		// this edge shouldn't exist anymore.
		_, err = uut.LocateSegment(quadedge.Vertex(tc.deleteMe[0]), quadedge.Vertex(tc.deleteMe[1]))
		if err == nil {
			t.Fatalf("error locating segment, expected %v got nil", quadedge.ErrLocateFailure)
			return
		}

	}
	testcases := []tcase{
		{
			inputWKT:      `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:      `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			deleteMe: geom.Line{{0,10}, {10, 0}},
		},
		{
			inputWKT:      `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:      `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			deleteMe: geom.Line{{0,10}, {10, 0}},
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

/*
TestTriangulation test cases test for small constrained triangulations and 
edge cases
*/
func TestTriangulation(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
    // A simple website for performing conversions:
    // https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB      string
		expectedEdges string
		expectedTris  string
	}

	fn := func(t *testing.T, tc tcase) {
		bytes, err := hex.DecodeString(tc.inputWKB)
		if err != nil {
			t.Fatalf("error decoding hex string: %v", err)
			return
		}
		g, err := wkb.DecodeBytes(bytes)
		if err != nil {
			t.Fatalf("error decoding WKB: %v", err)
			return
		}

		uut := new(Triangulator)
		uut.tolerance = 1e-6
		err = uut.InsertSegments(g)
		if err != nil {
			t.Fatalf("error inserting segments, expected nil got %v", err)
		}

		edges := uut.GetEdges()
		edgesWKT, err := wkt.Encode(edges)
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		if edgesWKT != tc.expectedEdges {
			t.Errorf("error, expected %v got %v", tc.expectedEdges, edgesWKT)
			return
		}

		tris, err := uut.GetTriangles()
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		trisWKT, err := wkt.Encode(tris)
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		if trisWKT != tc.expectedTris {
			t.Errorf("error, expected %v got %v", tc.expectedTris, trisWKT)
			return
		}
	}
	testcases := []tcase{
		{
			// should create a triangulation w/ a vertical line (2 5, 2 -5). 
			// The unconstrained version has a horizontal line
			inputWKT:      `LINESTRING(0 0, 2 5, 2 -5, 5 0)`,
			inputWKB:      `0102000000040000000000000000000000000000000000000000000000000000400000000000001440000000000000004000000000000014c000000000000014400000000000000000`,
			expectedEdges: `MULTILINESTRING ((2 5,5 0),(0 0,2 5),(0 0,2 -5),(2 -5,5 0),(2 -5,2 5))`,
			expectedTris: `MULTIPOLYGON (((0 0,2 -5,2 5,0 0)),((2 5,2 -5,5 0,2 5)))`,
		},
		{
			// a horizontal rectangle w/ one diagonal line. The diagonal line
			// should be maintained and the top/bottom re-triangulated.
			inputWKT:      `MULTILINESTRING ((0 0,0 1,1 1.1,2 1,2 0,1 -0.1,0 0),(0 0,2 1))`,
			inputWKB:      `010500000002000000010200000007000000000000000000000000000000000000000000000000000000000000000000f03f000000000000f03f9a9999999999f13f0000000000000040000000000000f03f00000000000000400000000000000000000000000000f03f9a9999999999b93f000000000000000000000000000000000102000000020000000000000000000000000000000000f03f00000000000000400000000000000000`,
			expectedEdges: `MULTILINESTRING ((1 1.1,2 1),(0 1,1 1.1),(0 0,0 1),(0 0,2 0),(2 0,2 1),(1 1.1,2 0),(0 1,2 0),(1 0.1,2 0),(0 1,1 0.1),(0 0,1 0.1))`,
			expectedTris: `MULTIPOLYGON (((0 1,0 0,1 0.1,0 1)),((0 1,1 0.1,2 0,0 1)),((0 1,2 0,1 1.1,0 1)),((1 1.1,2 0,2 1,1 1.1)),((0 0,2 0,1 0.1,0 0)))`,
		},
		{
			// a horizontal rectangle w/ one diagonal line. The diagonal line
			// should be maintained and the top/bottom re-triangulated.
			inputWKT:      `MULTILINESTRING((0 0,-0.1 0.5,0 1,0.5 1.2,1 1.3,1.5 1.2,2 1,2.1 0.5,2 0,1.5 -0.2,1 -0.3,0.5 -0.2,0 0),(-0.1 0.5,2.1 0.5))`,
			inputWKB:      `01050000000200000001020000000d000000000000000000000000000000000000009a9999999999b9bf000000000000e03f0000000000000000000000000000f03f000000000000e03f333333333333f33f000000000000f03fcdccccccccccf43f000000000000f83f333333333333f33f0000000000000040000000000000f03fcdcccccccccc0040000000000000e03f00000000000000400000000000000000000000000000f83f9a9999999999c9bf000000000000f03f333333333333d3bf000000000000e03f9a9999999999c9bf000000000000000000000000000000000102000000020000009a9999999999b9bf000000000000e03fcdcccccccccc0040000000000000e03f`,
			expectedEdges: `MULTILINESTRING ((1.5 1.2,2 1),(1 1.3,1.5 1.2),(0.5 1.2,1 1.3),(0 1,0.5 1.2),(-0.1 0.5,0 1),(-0.1 0.5,0 0),(0 0,0.5 -0.2),(0.5 -0.2,1 -0.3),(1 -0.3,1.5 -0.2),(1.5 -0.2,2 0),(2 0,2.1 0.5),(2 1,2.1 0.5),(1.5 1.2,2.1 0.5),(1 1.3,2.1 0.5),(-0.1 0.5,2.1 0.5),(-0.1 0.5,1 1.3),(-0.1 0.5,0.5 1.2),(1.5 -0.2,2.1 0.5),(-0.1 0.5,1.5 -0.2),(0.5 -0.2,1.5 -0.2),(-0.1 0.5,0.5 -0.2))`,
			expectedTris: `MULTIPOLYGON (((0 1,-0.1 0.5,0.5 1.2,0 1)),((0.5 1.2,-0.1 0.5,1 1.3,0.5 1.2)),((1 1.3,-0.1 0.5,2.1 0.5,1 1.3)),((1 1.3,2.1 0.5,1.5 1.2,1 1.3)),((1.5 1.2,2.1 0.5,2 1,1.5 1.2)),((1 -0.3,1.5 -0.2,0.5 -0.2,1 -0.3)),((0.5 -0.2,1.5 -0.2,-0.1 0.5,0.5 -0.2)),((0.5 -0.2,-0.1 0.5,0 0,0.5 -0.2)),((-0.1 0.5,1.5 -0.2,2.1 0.5,-0.1 0.5)),((2.1 0.5,1.5 -0.2,2 0,2.1 0.5)))`,
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

