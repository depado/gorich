package table

import (
	"reflect"
	"testing"
)

func TestRatioResolve(t *testing.T) {
	tests := []struct {
		name  string
		total int
		edges []Edge
		want  []int
	}{
		{
			name:  "three equal ratios",
			total: 110,
			edges: []Edge{
				{Ratio: 1, MinimumSize: 1},
				{Ratio: 1, MinimumSize: 1},
				{Ratio: 1, MinimumSize: 1},
			},
			want: []int{37, 36, 37},
		},
		{
			name:  "fixed size",
			total: 100,
			edges: []Edge{
				{Size: 20},
				{Ratio: 1, MinimumSize: 1},
				{Ratio: 1, MinimumSize: 1},
			},
			want: []int{20, 40, 40},
		},
		{
			name:  "ratio 2:1",
			total: 60,
			edges: []Edge{
				{Ratio: 2, MinimumSize: 1},
				{Ratio: 1, MinimumSize: 1},
			},
			want: []int{40, 20},
		},
		{
			name:  "minimum size enforced",
			total: 10,
			edges: []Edge{
				{Ratio: 1, MinimumSize: 20},
				{Ratio: 1, MinimumSize: 1},
			},
			want: []int{20, 1},
		},
		{
			name:  "insufficient total, all min",
			total: 5,
			edges: []Edge{
				{Ratio: 1, MinimumSize: 10},
				{Ratio: 1, MinimumSize: 10},
			},
			want: []int{10, 10},
		},
		{
			name:  "single edge",
			total: 42,
			edges: []Edge{
				{Ratio: 1, MinimumSize: 1},
			},
			want: []int{42},
		},
		{
			name:  "edge with size 0 treated as unset",
			total: 30,
			edges: []Edge{
				{Size: 0, Ratio: 2, MinimumSize: 1},
				{Size: 0, Ratio: 1, MinimumSize: 1},
			},
			want: []int{20, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ratioResolve(tt.total, tt.edges)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ratioResolve(%d, %+v) = %v, want %v", tt.total, tt.edges, got, tt.want)
			}
		})
	}
}

func TestRatioDistribute(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		ratios   []int
		minimums []int
		want     []int
		wantSum  int
	}{
		{
			name:    "equal distribution",
			total:   100,
			ratios:  []int{1, 1},
			want:    []int{50, 50},
			wantSum: 100,
		},
		{
			name:    "ratio 3:1",
			total:   100,
			ratios:  []int{3, 1},
			want:    []int{75, 25},
			wantSum: 100,
		},
		{
			name:    "ratio 2:1:1",
			total:   100,
			ratios:  []int{2, 1, 1},
			want:    []int{50, 25, 25},
			wantSum: 100,
		},
		{
			name:     "with minimums",
			total:    100,
			ratios:   []int{1, 1, 1},
			minimums: []int{10, 0, 0},
			want:     []int{100, 0, 0},
			wantSum:  100,
		},
		{
			name:     "all with minimums",
			total:    100,
			ratios:   []int{1, 1, 1},
			minimums: []int{10, 10, 10},
			want:     []int{34, 33, 33},
			wantSum:  100,
		},
		{
			name:     "with minimums, ratio zeroed",
			total:    100,
			ratios:   []int{0, 1, 1},
			minimums: []int{10, 0, 0},
			want:     []int{100, 0, 0},
			wantSum:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ratioDistribute(tt.total, tt.ratios, tt.minimums)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ratioDistribute(%d, %v, %v) = %v, want %v",
					tt.total, tt.ratios, tt.minimums, got, tt.want)
			}
			sum := 0
			for _, v := range got {
				sum += v
			}
			if sum != tt.wantSum {
				t.Errorf("ratioDistribute sum = %d, want %d", sum, tt.wantSum)
			}
		})
	}
}

func TestRatioReduce(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		ratios   []int
		maximums []int
		values   []int
		want     []int
	}{
		{
			name:     "basic reduction",
			total:    50,
			ratios:   []int{1, 1},
			maximums: []int{100, 100},
			values:   []int{60, 40},
			want:     []int{35, 15},
		},
		{
			name:     "zero ratio no reduction",
			total:    50,
			ratios:   []int{0, 1},
			maximums: []int{100, 100},
			values:   []int{30, 40},
			want:     []int{30, -10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ratioReduce(tt.total, tt.ratios, tt.maximums, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ratioReduce(%d, %v, %v, %v) = %v, want %v",
					tt.total, tt.ratios, tt.maximums, tt.values, got, tt.want)
			}
		})
	}
}

func TestRatioResolveSum(t *testing.T) {
	// Verify that ratio_resolve sums to total when possible
	tests := []struct {
		total int
		edges []Edge
	}{
		{110, []Edge{{Ratio: 1, MinimumSize: 1}, {Ratio: 1, MinimumSize: 1}, {Ratio: 1, MinimumSize: 1}}},
		{100, []Edge{{Size: 20}, {Ratio: 1, MinimumSize: 1}, {Ratio: 1, MinimumSize: 1}}},
		{60, []Edge{{Ratio: 2, MinimumSize: 1}, {Ratio: 1, MinimumSize: 1}}},
		{30, []Edge{{Ratio: 1, MinimumSize: 1}}},
	}
	for _, tt := range tests {
		got := ratioResolve(tt.total, tt.edges)
		sum := 0
		for _, v := range got {
			sum += v
		}
		if sum != tt.total {
			t.Errorf("ratioResolve(%d, ...) sum = %d, want %d", tt.total, sum, tt.total)
		}
	}
}
