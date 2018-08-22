package clip

import (
	"context"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

var testExtents = [...]*geom.Extent{
	/* 000 */ geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
	/* 001 */ geom.NewExtent([2]float64{2, 2}, [2]float64{9, 9}),
	/* 002 */ geom.NewExtent([2]float64{0, 0}, [2]float64{11, 11}),
	/* 003 */ geom.NewExtent([2]float64{-2, -2}, [2]float64{12, 12}),
	/* 004 */ geom.NewExtent([2]float64{-3, -3}, [2]float64{13, 13}),

	/* 005 */ geom.NewExtent([2]float64{-4, -4}, [2]float64{14, 14}),
	/* 006 */ geom.NewExtent([2]float64{5, 1}, [2]float64{7, 3}),
	/* 007 */ geom.NewExtent([2]float64{0, 5}, [2]float64{2, 7}),
	/* 008 */ geom.NewExtent([2]float64{0, 5}, [2]float64{2, 7}),
	/* 009 */ geom.NewExtent([2]float64{5, 2}, [2]float64{11, 9}),

	/* 010 */ geom.NewExtent([2]float64{-1, -1}, [2]float64{11, 11}),
	/* 011 */ geom.NewExtent([2]float64{0, 0}, [2]float64{4096, 4096}),
}

func TestClipLineString(t *testing.T) {
	type tcase struct {
		extent   *geom.Extent
		linestr  geom.LineString
		expected geom.MultiLineString
		err      error
	}

	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		ctx := context.Background()
		mls, err := lineString(ctx, tc.linestr, tc.extent)
		// Check the error values first.
		switch {
		case tc.err != nil && err == nil:
			t.Errorf("unexpected error, expected %v, got nil", tc.err)
			return
		case tc.err == nil && err != nil:
			t.Errorf("unexpected error, expected nil, got %v", err)
			return
		case tc.err == err && tc.err != nil && tc.err.Error() != err.Error():
			t.Errorf("unexpected error, expected %v, got %v", tc.err, err)
			return
		case tc.err != nil:
			// we are expecting an error. And if it got to this point, then the error is the expected error.
			// nothing more to do.
			return
		}
		if len(tc.expected) != len(mls) {
			t.Errorf("number of lines, expected %v got %v", len(tc.expected), len(mls))
			t.Errorf("\texpected: %v", tc.expected)
			t.Errorf("\tgot     : %v", mls)
			return
		}

		for i := range tc.expected {
			if !cmp.LineStringEqual(tc.expected[i], mls[i]) {
				t.Errorf("line %v, \n\tExpected %v\n\tgot     %v", i, tc.expected[i], mls[i])
			}
		}

	}

	tests := [...]tcase{
		{ /* 000 */
			extent:  testExtents[0],
			linestr: [][2]float64{{-2, 1}, {2, 1}, {2, 2}, {-1, 2}, {-1, 11}, {2, 11}, {2, 4}, {4, 4}, {4, 13}, {-2, 13}},
			expected: [][][2]float64{
				{{0, 1}, {2, 1}, {2, 2}, {0, 2}},
				{{2, 10}, {2, 4}, {4, 4}, {4, 10}},
			},
		},
		{ /* 001 */
			extent:  testExtents[0],
			linestr: [][2]float64{{-2, 1}, {12, 1}, {12, 2}, {-1, 2}, {-1, 11}, {2, 11}, {2, 4}, {4, 4}, {4, 13}, {-2, 13}},
			expected: geom.MultiLineString{
				[][2]float64{{0, 1}, {10, 1}},
				[][2]float64{{10, 2}, {0, 2}},
				[][2]float64{{2, 10}, {2, 4}, {4, 4}, {4, 10}},
			},
		},
		{ /* 002 */
			extent:  testExtents[0],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{0, 9}, {10, 9}},
				[][2]float64{{10, 2}, {5, 2}, {5, 8}, {0, 8}},
				[][2]float64{{0, 4}, {3, 4}, {3, 1}},
			},
		},
		{ /* 003 */
			extent:  testExtents[1],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{2, 9}, {9, 9}},
				[][2]float64{{9, 2}, {5, 2}, {5, 8}, {2, 8}},
				[][2]float64{{2, 4}, {3, 4}, {3, 2}},
			},
		},
		{ /* 004 */
			extent:  testExtents[2],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{0, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {0, 8}},
				[][2]float64{{0, 4}, {3, 4}, {3, 1}},
			},
		},
		{ /* 005 */
			extent:  testExtents[3],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{-2, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			},
		},
		{ /* 006 */
			extent:  testExtents[4],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			},
		},
		{ /* 007 */
			extent:  testExtents[5],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			},
		},
		{ /* 008 */
			extent:  testExtents[6],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{7, 2}, {5, 2}, {5, 3}},
			},
		},
		{ /* 009 */
			extent:   testExtents[7],
			linestr:  [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: nil,
		},
		{ /* 010 */
			extent:   testExtents[8],
			linestr:  [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: nil,
		},
		{ /* 011 */
			extent:  testExtents[9],
			linestr: [][2]float64{{-3, 1}, {-3, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: geom.MultiLineString{
				[][2]float64{{5, 9}, {11, 9}, {11, 2}, {5, 2}, {5, 8}},
			},
		},
		{ /* 012 */
			extent:   testExtents[9],
			linestr:  [][2]float64{{-3, 1}, {-3, 10}, {12, 10}, {12, 1}, {4, 1}, {4, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 1}},
			expected: nil,
		},
		{ /* 013 */
			extent:  testExtents[0],
			linestr: [][2]float64{{-3, -3}, {-3, 10}, {12, 10}, {12, 1}, {4, 1}, {4, 8}, {-1, 8}, {-1, 4}, {3, 4}, {3, 3}},
			expected: geom.MultiLineString{
				[][2]float64{{0, 10}, {10, 10}},
				[][2]float64{{10, 1}, {4, 1}, {4, 8}, {0, 8}},
				[][2]float64{{0, 4}, {3, 4}, {3, 3}},
			},
		},
		{ /* 014 */
			extent:  testExtents[10],
			linestr: [][2]float64{{-1, -1}, {12, -1}, {12, 12}, {-1, 12}},
			expected: geom.MultiLineString{
				[][2]float64{{-1, -1}, {11, -1}},
			},
		},
		{ /* 015 */
			extent: testExtents[11],

			linestr: [][2]float64{{7848, 19609}, {7340, 18835}, {6524, 17314}, {6433, 17163}, {5178, 15057}, {5147, 15006}, {4680, 14226}, {3861, 12766}, {2471, 10524}, {2277, 10029}, {1741, 8281}, {1655, 8017}, {1629, 7930}, {1437, 7368}, {973, 5481}, {325, 4339}, {-497, 3233}, {-1060, 2745}, {-1646, 2326}, {-1883, 2156}, {-2002, 2102}, {-2719, 1774}, {-3638, 1382}, {-3795, 1320}, {-5225, 938}, {-6972, 295}, {-7672, -88}, {-8243, -564}, {-8715, -1112}, {-9019, -1573}, {-9235, -2067}, {-9293, -2193}, {-9408, -2570}, {-9823, -4630}, {-10118, -5927}, {-10478, -7353}, {-10909, -8587}, {-11555, -9743}, {-11837, -10005}, {-12277, -10360}, {-13748, -11189}, {-14853, -12102}, {-15806, -12853}, {-16711, -13414}},
			expected: geom.MultiLineString{
				[][2]float64{{144.397830, 4096}, {-0, 3901.712895}},
			},
		},
		{ /* 016 */
			extent:   testExtents[11],
			linestr:  [][2]float64{},
			expected: nil,
		},
		{ /* 017 */
			extent:   testExtents[11],
			linestr:  [][2]float64{{-1, 1}, {1, -1}},
			expected: nil,
		},
		{ /* 018 */
			extent:  nil,
			linestr: [][2]float64{{-1, 1}, {1, -1}},
			expected: geom.MultiLineString{
				[][2]float64{{-1, 1}, {1, -1}},
			},
		},
		{ /* 019 */
			extent:  testExtents[11],
			linestr: [][2]float64{{-1, 1}},
			err:     geom.ErrInvalidLineString,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}
}
