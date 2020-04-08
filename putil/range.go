package putil

type Int64Range struct {
	Min int64
	Max int64
}

func IsOverlapWithTwoInt64Ranges(a, b Int64Range) bool {
	return a.Min <= b.Max && b.Min <= a.Max
}

func FirstOverlapIndexesWithInt64Ranges(ranges []Int64Range) (int, int) {
	for idx, r := range ranges {
		for i := idx + 1; i < len(ranges); i++ {
			if IsOverlapWithTwoInt64Ranges(r, ranges[i]) {
				return idx, i
			}
		}
	}
	return -1, -1
}
