package jump

type Scale interface {
	Scale() int
}

func IndexOfScales(scales []Scale, key uint64) int {
	bucket := 0
	for _, s := range scales {
		bucket += s.Scale()
	}

	idx := Hash(key, bucket)

	p := 0
	for i, s := range scales {
		scale := s.Scale()
		if idx >= p && idx < p+scale {
			return i
		}
		p += scale
	}
	return -1
}
