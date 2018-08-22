package kdtree

import (
	"encoding/json"
	"testing"

	"github.com/go-spatial/geom"
)

// TestInsertPoints is a very simple white-box test to validate the structure of the tree
func TestInsertPoints(t *testing.T) {
	type tcase struct {
		points []geom.Point
		eJSON  string
		err    error
	}

	fn := func(t *testing.T, tc tcase) {
		var err error

		uut := new(KdTree)

		for _, pt := range tc.points {
			if _, err = uut.Insert(pt); err != nil {
				break
			}
		}

		if tc.err == nil && err != nil {
			t.Errorf("insert points error, expected nil, got %v", err)
			return
		}

		if tc.err != nil {
			if err == nil || err.Error() != tc.err.Error() {
				t.Errorf("error, expected %f got %v", tc.err, err)
			}
			return
		}

		gJSON, err := json.Marshal(uut.root)
		if err != nil {
			t.Fatalf("converting to json error, expected nil, got %v", err)
			return
		}

		if tc.eJSON != string(gJSON) {
			t.Errorf("nearest neighbor iterator, expected %v got %v", tc.eJSON, string(gJSON))
		}
	}

	tests := map[string]tcase{
		"good": {
			points: []geom.Point{
				{0, 0},
				{1, 0},
				{1, 1},
				{-1, 0},
			},
			eJSON: `{"P":[0,0],"Left":{"P":[-1,0]},"Right":{"P":[1,0],"Right":{"P":[1,1]}}}`,
		},
		// insert 0,0 twice which should return an error
		"duplicate point": {
			points: []geom.Point{
				{0, 0},
				{1, 0},
				{1, 1},
				{0, 0},
				{-1, 0},
			},
			err: ErrDuplicateNode,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
