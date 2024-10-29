package sortfiles

import (
	"cmp"
	"slices"
	"strings"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
)

// ByTimeVar sorts changes by user-specified time variable.
func ByTimeVar(
	changes file.Changes,
	conf *config.Config,
) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		if a.PrimaryPair != nil {
			a.CustomSort.Time = a.PrimaryPair.CustomSort.Time
		}

		if b.PrimaryPair != nil {
			b.CustomSort.Time = b.PrimaryPair.CustomSort.Time
		}

		timeA := a.CustomSort.Time
		timeB := b.CustomSort.Time

		if conf.SortPerDir && a.BaseDir != b.BaseDir {
			return 0
		}

		if conf.ReverseSort {
			return -cmp.Compare(timeA.UnixNano(), timeB.UnixNano())
		}

		return cmp.Compare(timeA.UnixNano(), timeB.UnixNano())
	})
}

// ByStringVar sorts changes by user-specified string variable
// (lexicographically).
func ByStringVar(
	changes file.Changes,
	conf *config.Config,
) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		if a.PrimaryPair != nil {
			a.CustomSort.String = a.PrimaryPair.CustomSort.String
		}

		if b.PrimaryPair != nil {
			b.CustomSort.String = b.PrimaryPair.CustomSort.String
		}

		strA := a.CustomSort.String
		strB := b.CustomSort.String

		if conf.SortPerDir && a.BaseDir != b.BaseDir {
			return 0
		}

		if conf.ReverseSort {
			return strings.Compare(strB, strA)
		}

		return strings.Compare(strA, strB)
	})
}

// ByIntVar sorts changes by user-specified integer variable.
func ByIntVar(
	changes file.Changes,
	conf *config.Config,
) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		if a.PrimaryPair != nil {
			a.CustomSort.Int = a.PrimaryPair.CustomSort.Int
		}

		if b.PrimaryPair != nil {
			b.CustomSort.Int = b.PrimaryPair.CustomSort.Int
		}

		intA := a.CustomSort.Int
		intB := b.CustomSort.Int

		if conf.SortPerDir && a.BaseDir != b.BaseDir {
			return 0
		}

		if conf.ReverseSort {
			return cmp.Compare(intB, intA)
		}

		return cmp.Compare(intA, intB)
	})
}
