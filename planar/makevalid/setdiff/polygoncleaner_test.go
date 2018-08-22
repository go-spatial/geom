package setdiff

import (
	"encoding/hex"
	"log"
	"strconv"
	"testing"

	//"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkb"
	"github.com/go-spatial/geom/encoding/wkt"
)

/*
TestTriangulation test cases test for small constrained triangulations and
edge cases
*/
func TestPolygonMakeValid(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
		// A simple website for performing conversions:
		// https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB       string
		expectedWKT    string
		expectedInside string
	}

	// to change the flags on the default logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

		uut := new(PolygonCleaner)
		uut.tolerance = 1e-6
		vg, err := uut.MakeValid(g)
		if err != nil {
			t.Fatalf("error inserting segments, expected nil got %v", err)
		}

		s := uut.getLabelsAsString()
		if tc.expectedInside != `disabled` && s != tc.expectedInside {
			t.Errorf("error, expected %#v got %#v", tc.expectedInside, s)
			return
		}

		vgWKT, err := wkt.Encode(vg)
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		if vgWKT != tc.expectedWKT {
			t.Errorf("error, expected %v got %v", tc.expectedWKT, vgWKT)
			return
		}
	}
	testcases := []tcase{
		{
			// Simple triangle, no errors
			inputWKT:       `POLYGON((0 0,1 1,2 0,0 0))`,
			inputWKB:       `0103000000010000000400000000000000000000000000000000000000000000000000f03f000000000000f03f0000000000000040000000000000000000000000000000000000000000000000`,
			expectedWKT:    `POLYGON ((1 1,0 0,2 0,1 1))`,
			expectedInside: `inside: [[0 0],[1 1],[2 0]]`,
		},
		{
			// Simple four sided polygon, no errors
			inputWKT:       `POLYGON((0 0,1 1,3 0,1 -1,0 0))`,
			inputWKB:       `0103000000010000000500000000000000000000000000000000000000000000000000f03f000000000000f03f00000000000008400000000000000000000000000000f03f000000000000f0bf00000000000000000000000000000000`,
			expectedWKT:    `POLYGON ((1 1,0 0,1 -1,3 0,1 1))`,
			expectedInside: "inside: [[0 0],[1 1],[1 -1]]\ninside: [[1 -1],[1 1],[3 0]]",
		},
		{
			// Bow-tie with four sided concave polygon on left and triangle on
			// right. Should break into two polygons
			inputWKT:       `POLYGON ((0 0, 0.2 0.3, 0 1, 2 0, 2 1, 0 0))`,
			inputWKB:       `01030000000100000006000000000000000000000000000000000000009a9999999999c93f333333333333d33f0000000000000000000000000000f03f000000000000004000000000000000000000000000000040000000000000f03f00000000000000000000000000000000`,
			expectedWKT:    `MULTIPOLYGON (((0.2 0.3,0 0,1 0.5,0 1,0.2 0.3)),((2 1,1 0.5,2 0,2 1)))`,
			expectedInside: "inside: [[0 0],[0.2 0.3],[1 0.5]]\ninside: [[0 1],[1 0.5],[0.2 0.3]]\ninside: [[1 0.5],[2 1],[2 0]]",
		},
		{
			// Super simple triangle with a hole in the center. No errors.
			inputWKT:       `POLYGON((0 0,3 3,6 0,0 0),(2 1,3 2,4 1,2 1))`,
			inputWKB:       `0103000000020000000400000000000000000000000000000000000000000000000000084000000000000008400000000000001840000000000000000000000000000000000000000000000000040000000000000000000040000000000000f03f000000000000084000000000000000400000000000001040000000000000f03f0000000000000040000000000000f03f`,
			expectedWKT:    `POLYGON ((3 3,0 0,6 0,3 3),(2 1,3 2,4 1,2 1))`,
			expectedInside: "inside: [[0 0],[2 1],[4 1]]\ninside: [[0 0],[3 3],[2 1]]\ninside: [[0 0],[4 1],[6 0]]\ninside: [[2 1],[3 3],[3 2]]\ninside: [[3 2],[3 3],[4 1]]\ninside: [[3 3],[6 0],[4 1]]",
		},
		{
			// Fairly simple multipolygon w/ hole. No errors.
			inputWKT:    `MULTIPOLYGON (((40 40, 20 45, 45 30, 40 40)), ((20 35, 10 30, 10 10, 30 5, 45 20, 20 35), (30 20, 20 15, 20 25, 30 20)))`,
			inputWKB:    `01060000000200000001030000000100000004000000000000000000444000000000000044400000000000003440000000000080464000000000008046400000000000003e4000000000000044400000000000004440010300000002000000060000000000000000003440000000000080414000000000000024400000000000003e40000000000000244000000000000024400000000000003e4000000000000014400000000000804640000000000000344000000000000034400000000000804140040000000000000000003e40000000000000344000000000000034400000000000002e40000000000000344000000000000039400000000000003e400000000000003440`,
			expectedWKT: `MULTIPOLYGON (((10 30,10 10,30 5,45 20,20 35,10 30),(20 15,20 25,30 20,20 15)),((40 40,20 45,45 30,40 40)))`,
			// not being tested...
			expectedInside: `disabled`,
		},
		{
			// Two triangles w/ overlap. Should produce a single polygon
			inputWKT:       `MULTIPOLYGON(((0 2,2 0,0 -1,0 2)),((1 0,2 1,3 -1,1 0)))`,
			inputWKB:       `0106000000020000000103000000010000000400000000000000000000000000000000000040000000000000004000000000000000000000000000000000000000000000f0bf0000000000000000000000000000004001030000000100000004000000000000000000f03f00000000000000000000000000000040000000000000f03f0000000000000840000000000000f0bf000000000000f03f0000000000000000`,
			expectedWKT:    `POLYGON ((0 2,0 -1,1.5 -0.25,3 -1,2 1,1.5 0.5,0 2))`,
			expectedInside: "inside: [[0 -1],[0 2],[1 0]]\ninside: [[0 -1],[1 0],[1.5 -0.25]]\ninside: [[0 2],[1.5 0.5],[1 0]]\ninside: [[1 0],[1.5 0.5],[2 0]]\ninside: [[1 0],[2 0],[1.5 -0.25]]\ninside: [[1.5 -0.25],[2 0],[3 -1]]\ninside: [[1.5 0.5],[2 1],[2 0]]\ninside: [[2 0],[2 1],[3 -1]]",
		},
		// Re-enable after the similar test in triangulator_test (line ~370) is re-enabled.
		// {
		// 	// Overlapping multipolygon w/ hole. Should produce a single
		// 	// polygon with a smaller hole
		// 	inputWKT:    `MULTIPOLYGON(((40 40,20 45,28 10,40 40)),((20 35,10 30,10 10,30 5,45 20,20 35),(30 20,20 15,20 25,30 20)))`,
		// 	inputWKB:    `0106000000020000000103000000010000000400000000000000000044400000000000004440000000000000344000000000008046400000000000003c40000000000000244000000000000044400000000000004440010300000002000000060000000000000000003440000000000080414000000000000024400000000000003e40000000000000244000000000000024400000000000003e4000000000000014400000000000804640000000000000344000000000000034400000000000804140040000000000000000003e40000000000000344000000000000034400000000000002e40000000000000344000000000000039400000000000003e400000000000003440`,
		// 	expectedWKT: ``,
		// 	expectedInside: "disabled",
		// },
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
