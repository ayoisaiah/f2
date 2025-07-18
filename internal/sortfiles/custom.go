package sortfiles

import (
	"cmp"
	"slices"
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// ByTimeVar sorts changes by user-specified time variable.
func ByTimeVar(
	changes file.Changes,
	conf *config.Config,
) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		if a.PrimaryPair != nil {
			a.SortCriterion.TimeVar = a.PrimaryPair.SortCriterion.TimeVar
		}

		if b.PrimaryPair != nil {
			b.SortCriterion.TimeVar = b.PrimaryPair.SortCriterion.TimeVar
		}

		timeA := a.SortCriterion.TimeVar
		timeB := b.SortCriterion.TimeVar

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
			a.SortCriterion.StringVar = a.PrimaryPair.SortCriterion.StringVar
		}

		if b.PrimaryPair != nil {
			b.SortCriterion.StringVar = b.PrimaryPair.SortCriterion.StringVar
		}

		strA := a.SortCriterion.StringVar
		strB := b.SortCriterion.StringVar

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
			a.SortCriterion.IntVar = a.PrimaryPair.SortCriterion.IntVar
		}

		if b.PrimaryPair != nil {
			b.SortCriterion.IntVar = b.PrimaryPair.SortCriterion.IntVar
		}

		intA := a.SortCriterion.IntVar
		intB := b.SortCriterion.IntVar

		if conf.SortPerDir && a.BaseDir != b.BaseDir {
			return 0
		}

		if conf.ReverseSort {
			return cmp.Compare(intB, intA)
		}

		return cmp.Compare(intA, intB)
	})
}
