package geom

type MultiLineString [][][2]float64

func (mls MultiLineString) LineStrings() [][][2]float64 {
	return mls
}
