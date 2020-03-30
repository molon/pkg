package jump

type Scaler interface {
	GetScale() int
}

func IndexOfScalers(scales []Scaler, key uint64) int {
	bucket := 0
	for _, s := range scales {
		bucket += s.GetScale()
	}

	idx := Hash(key, bucket)

	p := 0
	for i, s := range scales {
		scale := s.GetScale()
		if idx >= int32(p) && idx < int32(p+scale) {
			return i
		}
		p += scale
	}
	return -1
}
