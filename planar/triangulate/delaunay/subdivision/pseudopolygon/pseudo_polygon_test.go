package pseudopolygon

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/go-spatial/geom/internal/test/must"
	"github.com/go-spatial/geom/winding"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
)

var showanswer bool

func init() {
	showanswer, _ = strconv.ParseBool(strings.TrimSpace(os.Getenv("SHOWANSWER")))
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func TestTriangulate(t *testing.T) {
	type tcase struct {
		Desc   string
		points []geom.Point

		edges []geom.Line
		err   error
	}
	order := winding.Order{
		YPositiveDown: true,
	}
	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			edges, err := Triangulate(tc.points, order)

			if tc.err != nil {
				if tc.err != err {
					t.Errorf("error, expected %v got %v", tc.err, err)
				}
				return
			}

			if err != nil {
				t.Errorf("error, expected %v got %v", tc.err, err)
				return
			}

			if !reflect.DeepEqual(tc.edges, edges) {
				t.Errorf("edges,\n\t expected %v\n\t got      %v", wkt.MustEncode(tc.edges), wkt.MustEncode(edges))
				if showanswer {
					t.Errorf("edges:\n%#v", edges)
				}
			}

		}
	}
	tests := [...]tcase{
		{
			points: []geom.Point{},
			err:    ErrInvalidPseudoPolygonSize,
		},
		{
			points: []geom.Point{{0, 0}},
			err:    ErrInvalidPseudoPolygonSize,
		},
		{
			points: []geom.Point{{0, 0}, {1, 1}},
			edges:  []geom.Line{{{0, 0}, {1, 1}}},
		},
		{ // simple triangle
			points: []geom.Point{{10, 10}, {10, 20}, {20, 20}},
			edges: []geom.Line{
				{{10, 10}, {10, 20}},
				{{10, 20}, {20, 20}},
				{{20, 20}, {10, 10}},
			},
		},
		{
			Desc:   "all points on horizontal line",
			points: []geom.Point{{-10, 0}, {-5, 0}, {-1, 0}, {0, 0}, {5, 0}, {10, 0}},
			err:    ErrAllPointsColinear,
		},
		{
			Desc:   "From Parse MultiLines #4",
			points: []geom.Point{{10, 20}, {20, 20}, {20, 10}, {20, 0}, {10, 0}, {0, 0}, {0, 10}, {0, 20}},
			edges: must.ParseLines([]byte(
				`MULTILINESTRING ((0 10,0 20),(0 20,10 20),(10 20,0 10),(10 20,20 20),(20 20,20 10),(20 10,10 20),(20 10,20 0),(20 0,10 0),(10 0,20 10),(10 0,0 0),(0 0,0 10),(0 10,10 0),(10 0,20 10),(0 10,10 0),(0 10,20 10),(20 10,10 20),(10 20,0 10))`,
			)),
		},
		{
			points: []geom.Point{
				{10, 0},
				{0, 0},
				{0, 10},
				{0, 20},
			},
			edges: []geom.Line{
				{{0, 20}, {10, 0}},
				{{0, 10}, {0, 20}},
				{{0, 10}, {10, 0}},
				{{10, 0}, {0, 0}},
				{{0, 0}, {0, 10}},
			},
		},
		{
			points: []geom.Point{
				{458, 1228}, {457, 1225}, {449, 1196}, {456, 1225}, {457, 1232},
			},
			edges: []geom.Line{

				{{457, 1232}, {458, 1228}},
				{{458, 1228}, {457, 1225}},
				{{457, 1225}, {457, 1232}},
				{{457, 1232}, {457, 1225}},
				{{456, 1225}, {457, 1232}},
				{{456, 1225}, {457, 1225}},
				{{457, 1225}, {449, 1196}},
				{{449, 1196}, {456, 1225}},
			},
		},
		{ //7
			Desc: "multiple duplicated points",
			points: []geom.Point{
				{3940, 471}, {3936, 479}, {3941, 478}, {3936, 479}, {3937, 484}, {3936, 479}, {3936, 483}, {3936, 479}, {3932, 480},
			},
			edges: []geom.Line{
				/*
					// This is if you comment out the guard for lines
					// MULTILINESTRING ((3940 471,3936 479),(3936 479,3932 480),(3932 480,3940 471))

					{{3940,471},{3936,479}},
					{{3936,479},{3932,480}},
					{{3932,480},{3940,471}},

				*/

				// MULTILINESTRING ((3936 479,3941 478),(3936 479,3937 484),(3936 479,3936 483),(3940 471,3936 479),(3936 479,3932 480),(3932 480,3940 471))
				{{3936, 479}, {3941, 478}},
				{{3936, 479}, {3937, 484}},
				{{3936, 479}, {3936, 483}},
				{{3940, 471}, {3936, 479}},
				{{3936, 479}, {3932, 480}},
				{{3932, 480}, {3940, 471}},
			},
		},
		{
			//  test case for a set of points that are no colinear but
			// the closest points to the start points are
			//                * c
			//               / \
			//              /   \
			//             /     \
			//            /       \
			//           *----*----*
			//           d    a     b
			//
			Desc: "non-colinear w end start colinear",
			points: []geom.Point{
				{0, 0},
				{2, 0},
				{0, -3},
				{-2, 0},
			},
			edges: []geom.Line{
				{{-2, 0}, {0, 0}},
				{{0, -3}, {-2, 0}},
				{{0, -3}, {0, 0}},
				{{0, 0}, {2, 0}},
				{{2, 0}, {0, -3}},
			},
		},
		{
			//  test case for a set of points that are no colinear but
			// the closest points to the start points are
			//                * d
			//               / \
			//              /   \
			//             /     \
			//            /       \
			//           *-*--*-*--*
			//           e f  a b   c
			//
			Desc: "non-colinear w end start colinear",
			points: []geom.Point{
				{0, 0},
				{1, 0},
				{2, 0},
				{0, -4},
				{-2, 0},
				{-1, 0},
			},
			edges: []geom.Line{
				{{0, -4}, {-1, 0}},
				{{-1, 0}, {0, 0}},
				{{0, 0}, {0, -4}},
				{{0, 0}, {1, 0}},
				{{0, -4}, {0, 0}},
				{{0, -4}, {1, 0}},
				{{1, 0}, {2, 0}},
				{{2, 0}, {0, -4}},
				{{0, -4}, {-2, 0}},
				{{-2, 0}, {-1.0}},
				{{-1, 0}, {0, -4}},
			},
		},
		{
			// simple test case with repeated points (b == e)
			//
			//   a *       * c
			//     | \   / |
			//     |  \b/  |
			//     |   *   |
			//     |  /e\  |
			//     | /   \ |
			//   f *      \* d
			//
			Desc: "simple dup pt",
			points: []geom.Point{
				{-2, -2}, {0, 0}, {2, -2},
				{2, 2}, {0, 0}, {-2, 2},
			},
			edges: []geom.Line{
				{{0, 0}, {2, -2}},   // bc
				{{2, -2}, {2, 2}},   // cd
				{{2, 2}, {0, 0}},    // db
				{{-2, -2}, {0, 0}},  // ab
				{{0, 0}, {-2, 2}},   // bf
				{{-2, 2}, {-2, -2}}, // fa
			},
		},
		{
			Desc: "clockwise triangle",
			points: []geom.Point{
				{10, 0}, {0, 0}, {0, 10}, {0, 20},
			},
			edges: []geom.Line{
				{{0, 20}, {10, 0}},
				{{0, 10}, {0, 20}},
				{{0, 10}, {10, 0}},
				{{10, 0}, {0, 0}},
				{{0, 0}, {0, 10}},
			},
		},
		{
			Desc: "#04 set2",
			points: []geom.Point{
				{0, 20}, {10, 20}, {20, 20}, {20, 10}, {20, 0}, {10, 0},
			},
			edges: []geom.Line{
				{{10, 0}, {0, 20}},
				{{0, 20}, {10, 20}},
				{{10, 20}, {10, 0}},
				{{10, 0}, {10, 20}},
				{{20, 10}, {10, 0}},
				{{20, 10}, {10, 20}},
				{{10, 20}, {20, 20}},
				{{20, 20}, {20, 10}},
				{{20, 10}, {20, 0}},
				{{20, 0}, {10, 0}},
				{{10, 0}, {20, 10}},
			},
		},
		{
			Desc: "#04 set3",
			points: []geom.Point{
				{10, 20}, {20, 20}, {20, 10}, {20, 0}, {10, 0},
			},
			edges: []geom.Line{
				{{10, 0}, {10, 20}},
				{{20, 10}, {10, 0}},
				{{20, 10}, {10, 20}},
				{{10, 20}, {20, 20}},
				{{20, 20}, {20, 10}},
				{{20, 10}, {20, 0}},
				{{20, 0}, {10, 0}},
				{{10, 0}, {20, 10}},
			},
		},
		{
			Desc: "dup in a row",
			points: []geom.Point{
				{2680.390, 3431.154}, {2676.023, 3439.196}, {2676.023, 3439.196}, {2676.168, 3439.720},
			},
			edges: []geom.Line{
				{{2680.390, 3431.154}, {2676.023, 3439.196}},
				{{2676.023, 3439.196}, {2676.168, 3439.720}},
				{{2676.168, 3439.720}, {2680.390, 3431.154}},
			},
		},
		{
			Desc: "Natural Earth bad triangulation",
			points: []geom.Point{
				{2305, 1210},
				{2302, 1209},
				{2301, 1208},
				{2300, 1206},
				{2299, 1205},
				{2297, 1203},
				{2296, 1202},
				{2294, 1197},
				{2309, 1200},
				{2309, 1201},
				{2309, 1208},
			},
			edges: must.ParseLines([]byte(
				`MULTILINESTRING ((2305 1210,2302 1209),(2302 1209,2301 1208),(2301 1208,2305 1210),(2305 1210,2301 1208),(2301 1208,2300 1206),(2300 1206,2305 1210),(2305 1210,2300 1206),(2300 1206,2299 1205),(2299 1205,2305 1210),(2305 1210,2299 1205),(2299 1205,2297 1203),(2297 1203,2305 1210),(2297 1203,2296 1202),(2296 1202,2294 1197),(2294 1197,2297 1203),(2297 1203,2294 1197),(2294 1197,2309 1200),(2309 1200,2297 1203),(2297 1203,2309 1200),(2309 1200,2309 1201),(2309 1201,2297 1203),(2297 1203,2309 1201),(2305 1210,2297 1203),(2305 1210,2309 1201),(2309 1201,2309 1208),(2309 1208,2305 1210))`,
			)),
		},
	}
	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestTriangulateSubRings(t *testing.T) {
	type tcase struct {
		Desc    string
		opoints []geom.Point
		points  []geom.Point
		edges   []geom.Line
		err     error
	}
	order := winding.Order{
		YPositiveDown: true,
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			points, edges, err := triangulateSubRings(tc.opoints, order)
			if tc.err != err {
				t.Errorf("error, expected %v got %v", tc.err, err)
				return
			}
			if len(points) != len(tc.points) {
				t.Errorf("points, expected %v got %v", wkt.MustEncode(tc.points), wkt.MustEncode(points))
			}
			if len(edges) != len(tc.edges) {
				t.Errorf("edges, expected %v got %v", tc.edges, edges)
			}

		}
	}

	tests := []tcase{
		// Subtests
		{
			Desc: "empty",
		},
		{
			opoints: []geom.Point{
				{10, 20}, {20, 20}, {20, 10}, {20, 0}, {10, 0}, {0, 0}, {0, 10}, {0, 20},
			},
			points: []geom.Point{
				{10, 20}, {20, 20}, {20, 10}, {20, 0}, {10, 0}, {0, 0}, {0, 10}, {0, 20},
			},
		},
		{
			Desc: "dup in a row",
			opoints: []geom.Point{
				{2680.390, 3431.154}, {2676.023, 3439.196}, {2676.023, 3439.196}, {2676.168, 3439.720},
			},
			points: []geom.Point{
				{2680.390, 3431.154}, {2676.023, 3439.196}, {2676.168, 3439.720},
			},
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}
