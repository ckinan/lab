package util_test

import (
	"testing"

	"github.com/ckinan/lab/internal/util"
)

func TestSortBy(t *testing.T) {
	type item struct{ val int }

	tests := []struct {
		name  string
		input []item
		desc  bool
		want  []int
	}{
		{
			name:  "ascending",
			input: []item{{3}, {1}, {2}},
			desc:  false,
			want:  []int{1, 2, 3},
		},
		{
			name:  "descending",
			input: []item{{3}, {1}, {2}},
			desc:  true,
			want:  []int{3, 2, 1},
		},
		{
			name:  "empty slice",
			input: []item{},
			desc:  false,
			want:  []int{},
		},
		{
			name:  "single item",
			input: []item{{42}},
			desc:  false,
			want:  []int{42},
		},
		{
			name:  "does not mutate original",
			input: []item{{3}, {1}, {2}},
			desc:  true,
			want:  []int{3, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := make([]item, len(tt.input))
			copy(original, tt.input)

			got := util.SortBy(tt.input, func(i item) int { return i.val }, tt.desc)

			// verify result order
			for i, want := range tt.want {
				if got[i].val != want {
					t.Errorf("index %d: got %d, want %d", i, got[i].val, want)
				}
			}

			// verify original was not mutated (SortBy clones the slice)
			for i, orig := range original {
				if tt.input[i] != orig {
					t.Errorf("original slice was mutated at index %d", i)
				}
			}
		})
	}
}
