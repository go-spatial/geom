package makevalid

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/internal/test/must"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

// makevalidCases encapsulates the various parts of the Tegola make valid algorithm
type makevalidCase struct {
	// Description is a simple description of the test
	Description string
	// The MultiPolyon describing the test case
	MultiPolygon *geom.MultiPolygon
	// The expected valid MultiPolygon for the Multipolygon
	ExpectedMultiPolygon *geom.MultiPolygon
}

// Segments returns the flattened segments of the MultiPolygon, on an error it will panic.
func (mvc makevalidCase) Segments() (segments []geom.Line) {
	if debug {
		log.Printf("MakeValidTestCase Polygon: %+v", mvc.MultiPolygon)
	}
	segs, err := Destructure(context.Background(), cmp, nil, mvc.MultiPolygon)
	if err != nil {
		panic(err)
	}
	if debug {
		log.Printf("MakeValidTestCase Polygon Segments: %+v", segs)
	}
	return segs
}

func (mvc makevalidCase) Hitmap(clp *geom.Extent) (hm planar.HitMapper) {
	var err error
	if hm, err = hitmap.New(clp, mvc.MultiPolygon); err != nil {
		panic("Hitmap gave error!")
	}
	return hm
}

var makevalidTestCases = func() []makevalidCase {

	f, err := os.Open(path.Join("testdata", "testcases"))

	if err != nil {
		panic(fmt.Sprintf("opening testcase dir: %v ", err))
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		panic(fmt.Sprintf("reading testcase dir: %v ", err))
	}
	// 0 is input filename, 1 is expected filename
	cases := make(map[string][2]string)
	var keys []string
	for i := range list {
		if list[i].IsDir() {
			continue
		}
		name := strings.ToLower(list[i].Name())
		if !strings.HasSuffix(name, ".wkt") {
			continue
		}
		if !strings.HasPrefix(name, "multipolygon_") {
			continue
		}
		key := bytes.TrimPrefix([]byte(list[i].Name()), []byte("multipolygon_"))

		// is it input?
		pos := 0
		switch {
		case bytes.HasSuffix(key, []byte("_input.wkt")):
			key = bytes.ReplaceAll(
				bytes.TrimSuffix(key, []byte("_input.wkt")),
				[]byte("_"),
				[]byte(" "),
			)
			pos = 0
		case bytes.HasSuffix(key, []byte("_expected.wkt")):
			key = bytes.ReplaceAll(
				bytes.TrimSuffix(key, []byte("_expected.wkt")),
				[]byte("_"),
				[]byte(" "),
			)
			pos = 1
		default:
			continue
		}
		casename, ok := cases[string(key)]
		if !ok {
			keys = append(keys, string(key))
		}
		casename[pos] = list[i].Name()
		cases[string(key)] = casename
	}
	sort.Strings(keys)
	tcases := make([]makevalidCase, len(keys))
	for i := range keys {
		var input, expected *geom.MultiPolygon
		if cases[keys[i]][0] != "" {
			fn := path.Join("testdata", "testcases", cases[keys[i]][0])
			input = must.MPPointer(must.ReadMultiPolygon(fn))
		}
		if cases[keys[i]][1] != "" {
			fn := path.Join("testdata", "testcases", cases[keys[i]][1])
			expected = must.MPPointer(must.ReadMultiPolygon(fn))
		}
		tcases[i] = makevalidCase{
			Description:          keys[i],
			MultiPolygon:         input,
			ExpectedMultiPolygon: expected,
		}
	}
	return tcases
}()
