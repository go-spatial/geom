package kdtree

import (
	"encoding/json"
	//"fmt"
	"testing"

	"github.com/go-spatial/geom"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func toJson(t *testing.T, v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return string(b)
}

/*
Very simple white-box test to validate the structure of the tree
*/
func TestInsertPoints(t *testing.T) {
	uut := NewKdTree()

	uut.Insert(geom.Point{0, 0})
	uut.Insert(geom.Point{1, 0})
	// duplicate points are not allowed
	_, err := uut.Insert(geom.Point{0, 0})
	assertEqual(t, err != nil, true)
	uut.Insert(geom.Point{1, 1})
	uut.Insert(geom.Point{-1, 0})

	assertEqual(t, `{"P":[0,0],"Left":{"P":[-1,0]},"Right":{"P":[1,0],"Right":{"P":[1,1]}}}`, toJson(t, uut.root))
}
