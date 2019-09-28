package mvt

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-spatial/geom/testing/must"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
)

var dumpSolution = flag.Bool("dump.solution", false, "Dump the solution to the test")

func TestEncodePolygon(t *testing.T) {
	type tcase struct {
		x, y    int64
		Polygon geom.Polygon
		g       []uint32
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			c := cursor{
				x: tc.x,
				y: tc.y,
			}

			g := c.encodePolygon(tc.Polygon)
			if len(g) != len(tc.g) {
				t.Errorf("g length, expected %v, got %v", len(tc.g), len(g))
				if *dumpSolution {
					dumpFilename := filepath.Join("testdata", "dump", t.Name()+".json")
					dumpDir := filepath.Dir(dumpFilename)
					os.MkdirAll(dumpDir, os.ModePerm)
					t.Logf("dumping got to %v", dumpFilename)
					f, err := os.Create(dumpFilename)
					if err != nil {
						t.Logf("unable to create dumpfile: %v", err)
						return
					}
					bytes, err := json.Marshal(g)
					if err != nil {
						t.Logf("failed to marshal to json: %v", err)
					}
					_, err = f.Write(bytes)
					if err != nil {
						t.Logf("failed to write to dumpfile: %v", err)
					}
				}
				return
			}

			// pad index
			iPad := int(math.Log10(float64(len(tc.g)))) + 1
			// pad expected
			ePad := 0
			// pad got
			gPad := 0

			for i := range tc.g {
				v := int(math.Log10(float64(tc.g[i]))) + 1
				if ePad < v {
					ePad = v
				}

				v = int(math.Log10(float64(g[i]))) + 1
				if gPad < v {
					gPad = v
				}
			}

			for i := range tc.g {
				if tc.g[i] != g[i] {
					t.Errorf("value not correct for %*d, expected %*d got %*d",
						iPad, i, ePad, tc.g[i], gPad, g[i])
				}
			}
		}
	}

	testForFile := func(file string) tcase {

		var sol []uint32
		filename := filepath.Join("testdata", file)
		f, err := ioutil.ReadFile(filename + ".wkt")
		if err != nil {
			panic(fmt.Sprintf("error opening file (%v.wkt): %v", filename, err))
		}
		poly := must.AsPolygon(must.Decode(wkt.DecodeBytes(f)))

		if info, err := os.Stat(filename + ".json"); !(os.IsNotExist(err) || info.IsDir()) {
			f, err = ioutil.ReadFile(filename + ".json")
			if err != nil {
				panic(fmt.Sprintf("error opening file (%v.json): %v", filename, err))
			}
			if err = json.Unmarshal(f, &sol); err != nil {
				panic(fmt.Sprintf("error un-marshaling file (%v.json): %v", filename, err))
			}
		}

		return tcase{
			Polygon: poly,
			g:       sol,
		}
	}

	tests := []string{"florida_keys"}

	for _, name := range tests {
		t.Run(name, fn(testForFile(name)))
	}
}
