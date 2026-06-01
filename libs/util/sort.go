package util

import (
	"cmp"
	"slices"
)

func SortBy[T any, K cmp.Ordered](items []T, key func(T) K, desc bool) []T {
	out := slices.Clone(items)
	slices.SortFunc(out, func(a, b T) int {
		if desc {
			return cmp.Compare(key(b), key(a))
		}
		return cmp.Compare(key(a), key(b))
	})
	return out
}
