package table

import "math"

// Edge describes a slot with size, ratio, and minimum constraints.
type Edge struct {
	Size        int
	Ratio       int
	MinimumSize int
}

// ratioResolve divides total space to satisfy size, ratio, and minimum_size constraints.
//
// It returns a list of integers that should add up to total in most cases.
// If it's impossible to satisfy, the returned list may be greater than total.
func ratioResolve(total int, edges []Edge) []int {
	sizes := make([]*int, len(edges))
	for i, e := range edges {
		if e.Size > 0 {
			v := e.Size
			sizes[i] = &v
		}
	}

	for {
		done := true
		for _, s := range sizes {
			if s == nil {
				done = false
				break
			}
		}
		if done {
			break
		}

		remaining := total
		for _, s := range sizes {
			if s != nil {
				remaining -= *s
			}
		}

		if remaining <= 0 {
			result := make([]int, len(edges))
			for i := range edges {
				if sizes[i] == nil {
					result[i] = maxInt(edges[i].MinimumSize, 1)
				} else {
					result[i] = *sizes[i]
				}
			}
			return result
		}

		// Collect flexible edges
		type flex struct {
			idx  int
			edge Edge
		}
		var flexible []flex
		for i, s := range sizes {
			if s == nil {
				flexible = append(flexible, flex{i, edges[i]})
			}
		}

		totalRatio := 0
		for _, f := range flexible {
			r := f.edge.Ratio
			if r < 1 {
				r = 1
			}
			totalRatio += r
		}

		portion := float64(remaining) / float64(totalRatio)

		adjusted := false
		for _, f := range flexible {
			r := f.edge.Ratio
			if r < 1 {
				r = 1
			}
			if int(portion*float64(r)) <= f.edge.MinimumSize {
				v := f.edge.MinimumSize
				sizes[f.idx] = &v
				adjusted = true
				break
			}
		}
		if adjusted {
			continue
		}

		remainder := 0.0
		for _, f := range flexible {
			r := f.edge.Ratio
			if r < 1 {
				r = 1
			}
			size := portion*float64(r) + remainder
			intSize := int(math.Round(size))
			remainder = size - float64(intSize)
			sizes[f.idx] = &intSize
		}
		break
	}

	result := make([]int, len(edges))
	for i, s := range sizes {
		if s != nil {
			result[i] = *s
		} else {
			result[i] = maxInt(edges[i].MinimumSize, 1)
		}
	}
	return result
}

// ratioReduce divides an integer total into parts based on ratios, reducing from current values.
func ratioReduce(total int, ratios, maximums, values []int) []int {
	adjRatios := make([]int, len(ratios))
	for i := range ratios {
		if maximums[i] > 0 {
			adjRatios[i] = ratios[i]
		}
	}
	totalRatio := sumInts(adjRatios)
	if totalRatio == 0 {
		result := make([]int, len(values))
		copy(result, values)
		return result
	}
	totalRemaining := total
	result := make([]int, len(values))
	for i := range adjRatios {
		if adjRatios[i] > 0 && totalRatio > 0 {
			distributed := minInt(maximums[i], int(math.Round(float64(adjRatios[i]*totalRemaining)/float64(totalRatio))))
			result[i] = values[i] - distributed
			totalRemaining -= distributed
			totalRatio -= adjRatios[i]
		} else {
			result[i] = values[i]
		}
	}
	return result
}

// ratioDistribute distributes an integer total into parts based on ratios.
func ratioDistribute(total int, ratios []int, minimums []int) []int {
	adjRatios := make([]int, len(ratios))
	if minimums != nil {
		for i := range adjRatios {
			if minimums[i] > 0 {
				adjRatios[i] = ratios[i]
			}
		}
	} else {
		copy(adjRatios, ratios)
	}

	totalRatio := sumInts(adjRatios)

	_mins := minimums
	if _mins == nil {
		_mins = make([]int, len(ratios))
	}

	totalRemaining := total
	result := make([]int, len(ratios))
	for i := range adjRatios {
		if totalRatio > 0 {
			distributed := int(math.Ceil(float64(adjRatios[i]*totalRemaining) / float64(totalRatio)))
			distributed = maxInt(distributed, _mins[i])
			result[i] = distributed
			totalRatio -= adjRatios[i]
			totalRemaining -= distributed
		} else {
			result[i] = totalRemaining
			totalRemaining = 0
		}
	}
	return result
}

func sumInts(vals []int) int {
	s := 0
	for _, v := range vals {
		s += v
	}
	return s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
