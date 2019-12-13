package quadedge

// Operation defines QuadEdge Algebra Operation that can be done
type Operation uint8

const (
	Rot = Operation(iota)
	InvRot
	Sym
	ONext
	OPrev
	DNext
	DPrev
	LNext
	LPrev
	RNext
	RPrev
)

// Apply will return the edge after the given operation have been applied to the Edge
func (e *Edge) Apply(ops ...Operation) *Edge {
	if e == nil {
		return nil
	}
	ee := e
	for i := range ops {
		// if ee ever becomes nil,  we know all other ops will be nil as well.
		if ee == nil {
			return nil
		}
		switch ops[i] {
		case Rot:
			ee = ee.Rot()
		case InvRot:
			ee = ee.InvRot()
		case Sym:
			ee = ee.Sym()
		case ONext:
			ee = ee.ONext()
		case OPrev:
			ee = ee.OPrev()
		case DNext:
			ee = ee.DNext()
		case DPrev:
			ee = ee.DPrev()
		case LNext:
			ee = ee.LNext()
		case LPrev:
			ee = ee.LPrev()
		case RNext:
			ee = ee.RNext()
		case RPrev:
			ee = ee.RPrev()
		}
	}
	return ee
}
