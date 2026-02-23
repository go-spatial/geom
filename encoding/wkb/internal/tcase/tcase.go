package tcase

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkb/internal/tcase/token"
	"github.com/go-spatial/geom/internal/parsing"
)

var ErrMissingDesc = fmt.Errorf("missing desc field")

type Type uint8

const (
	TypeNone   Type = 0
	TypeEncode Type = 1
	TypeDecode Type = 2
	TypeBoth   Type = 3
)

func (ot Type) Is(t Type) bool { return ot&t == t }

type C struct {
	Filename      string
	Desc          string
	DescPosition  parsing.Position
	BOM           binary.ByteOrder
	Skip          Type
	Expected      any
	DecodeError   string
	EncodeError   string
	Bytes         []byte
	BytesPosition parsing.Position
	SRID          uint32
}

func (c C) HasErrorFor(t Type) bool {
	switch t {
	case TypeEncode:
		return c.EncodeError != ""
	case TypeDecode:
		return c.DecodeError != ""
	}
	return false
}
func (c C) ErrorFor(t Type) string {
	switch t {
	case TypeEncode:
		return c.EncodeError
	case TypeDecode:
		return c.DecodeError
	}
	return ""
}
func (c C) DoesErrorMatch(t Type, e error) bool {
	if !c.HasErrorFor(t) {
		return e == nil
	}
	return e.Error() == c.ErrorFor(t)
}

func (c C) Geometry() geom.Geometry {
	if c.SRID == 0 {
		return c.Expected
	}
	switch g := c.Expected.(type) {
	default:
		return g
	case geom.Pointer:
		return geom.PointS{
			Srid: geom.Srid(c.SRID),
			Xy:   g.XY(),
		}
	case geom.MultiPointer:
		return geom.MultiPointS{
			Srid: geom.Srid(c.SRID),
			Mp:   g.Points(),
		}
	case geom.LineStringer:
		return geom.LineStringS{
			Srid: geom.Srid(c.SRID),
			Ls:   g.Vertices(),
		}
	case geom.MultiLineStringer:
		return geom.MultiLineStringS{
			Srid: geom.Srid(c.SRID),
			Mls:  g.LineStrings(),
		}
	case geom.Polygoner:
		return geom.PolygonS{
			Srid: geom.Srid(c.SRID),
			Pol:  g.LinearRings(),
		}
	case geom.MultiPolygoner:
		return geom.MultiPolygonS{
			Srid:         geom.Srid(c.SRID),
			MultiPolygon: g.Polygons(),
		}
	case geom.Collectioner:
		return geom.CollectionS{
			Srid:       geom.Srid(c.SRID),
			Collection: g.Geometries(),
		}
	}
}

func parse(r io.Reader, filename string) (cases []C, err error) {
	t := token.NewT(r)
	var cC *C
	for !t.AtEnd() {
		t.EatCommentsAndSpaces()
		if t.AtEnd() {
			break
		}
		label, err := t.ParseLabel()
		if err != nil {
			log.Printf("error trying to get label %#v", cC)
			return nil, err
		}

		switch strings.ToLower(string(label)) {

		case "bom":
			if cC == nil {
				return cases, ErrMissingDesc
			}
			bom := strings.ToLower(strings.TrimSpace(string(t.ParseTillEndIgnoreComments())))
			switch bom {
			case "little":
				cC.BOM = binary.LittleEndian
			case "big":
				cC.BOM = binary.BigEndian
			default:
				pos := t.Position()
				return cases, fmt.Errorf("invalid bom(%v) at %s, expect “little” or “big”", bom, pos.String())
			}

		case "bytes":
			if cC == nil {
				return cases, ErrMissingDesc
			}
			bin, err := t.ParseBinaryField()
			if err != nil {
				return cases, err
			}
			cC.BytesPosition = t.Position()
			cC.Bytes = bin

		case "desc":
			if cC != nil {
				cases = append(cases, *cC)
			}
			cC = new(C)
			cC.Filename = filename
			cC.DescPosition = t.Position()
			cC.Desc = strings.TrimSpace(string(t.ParseTillEndIgnoreComments()))

		case "decode_error":
			if cC == nil {
				return cases, ErrMissingDesc
			}
			cC.DecodeError = strings.TrimSpace(string(t.ParseTillEndIgnoreComments()))

		case "encode_error":
			if cC == nil {
				return cases, ErrMissingDesc
			}
			cC.EncodeError = strings.TrimSpace(string(t.ParseTillEndIgnoreComments()))

		case "expected", "geometry":
			if cC == nil {
				return cases, ErrMissingDesc
			}
			geom, err := t.ParseExpectedField()
			if err != nil {
				return cases, err
			}
			cC.Expected = geom

		case "skip":
			if cC == nil {
				return cases, ErrMissingDesc
			}
			val := strings.TrimSpace(string(t.ParseTillEndIgnoreComments()))
			strings.ToLower(val)
			strings.TrimSpace(val)
			switch val {
			case "encode":
				cC.Skip = TypeEncode
			case "decode":
				cC.Skip = TypeDecode
			case "both":
				cC.Skip = TypeBoth
			}
		case "srid":
			if cC == nil {
				return cases, ErrMissingDesc
			}
			sridStr := strings.ToLower(strings.TrimSpace(string(t.ParseTillEndIgnoreComments())))
			srid, err := strconv.ParseUint(sridStr, 10, 32)
			if err != nil {
				return cases, parsing.ErrAt{
					Pos: t.Position(),
					Err: fmt.Errorf("failed to parse srid[%s]: %w", sridStr, err),
				}
			}
			cC.SRID = uint32(srid)

		default:
			return cases, fmt.Errorf("unknown label:%v", string(label))

		}
	}
	cases = append(cases, *cC)
	return cases, nil
}

func Parse(r io.Reader, filename string) ([]C, error) {
	return parse(r, filename)
}
func ParseFile(filename string) ([]C, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parse(file, filename)
}

var isolatedFilenames = flag.String("tcase.Files", "", "List of comma separated file name to grab the test cases from; instead of all the files in the directory.")

func GetFiles(dir string) (filenames []string, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var ffiles []string
	if *isolatedFilenames != "" {
		ffiles = strings.Split(*isolatedFilenames, ",")
		for i := range ffiles {
			ffiles[i] = strings.TrimSpace(ffiles[i])
		}
	}
LOOP_FILES:
	for _, f := range files {
		fname := f.Name()
		fext := strings.ToLower(filepath.Ext(fname))
		if fext != ".tcase" {
			continue
		}
		if len(ffiles) != 0 {
			// need to filter out filenames.
			for i := range ffiles {
				if ffiles[i] == fname {
					goto ADD_FILE
				}
			}
			// We did not find a file matching this file so skip it.
			continue LOOP_FILES
		}
	ADD_FILE:
		filenames = append(filenames, filepath.Join(dir, fname))
	}
	return filenames, nil
}

func SprintBinary(bytes []byte, prefix string) (out string) {
	out = prefix + "//01 02 03 04 05 06 07 08"
	for i, b := range bytes {
		if i%8 == 0 {
			out += "\n" + prefix + "  "
		}
		out += fmt.Sprintf("%02x ", b)
	}
	return out
}
