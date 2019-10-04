package wkt

import (
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/go-spatial/geom"
)

func Encode(w io.Writer, geo geom.Geometry) error {
	return NewEncoder(w).Encode(geo, false)
}

func EncodeBytes(geo geom.Geometry) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := Encode(buf, geo)
	return buf.Bytes(), err
}

func EncodeString(geo geom.Geometry) (string, error) {
	byt, err := EncodeBytes(geo)
	return string(byt), err
}

func MustEncode(geo geom.Geometry) string {
	str, err := EncodeString(geo)
	if err != nil {
		log.Println(err)
		//panic(err)
		return ""
	}

	return str
}

func Decode(r io.Reader) (geo geom.Geometry, err error) {
	return NewDecoder(r).Decode()
}

func DecodeBytes(b []byte) (geo geom.Geometry, err error) {
	return Decode(bytes.NewReader(b))
}

func DecodeString(s string) (geo geom.Geometry, err error) {
	return Decode(strings.NewReader(s))
}
