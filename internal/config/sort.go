package config

import (
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/timeutil"
)

type Sort int

const (
	SortDefault Sort = iota
	SortSize
	SortNatural
	SortMtime
	SortBtime
	SortAtime
	SortCtime
	SortTimeVar
	SortIntVar
	SortStringVar
)

func (s Sort) String() string {
	return [...]string{"default", "size", "natural", timeutil.Mod, timeutil.Access, timeutil.Birth, timeutil.Change, "time_var", "int_var", "string_var"}[s]
}

func parseSortArg(arg string) (Sort, error) {
	arg = strings.TrimSpace(arg)

	switch arg {
	case "":
		return SortDefault, nil
	case SortDefault.String():
		return SortDefault, nil
	case SortSize.String():
		return SortSize, nil
	case SortNatural.String():
		return SortNatural, nil
	case SortMtime.String():
		return SortMtime, nil
	case SortBtime.String():
		return SortBtime, nil
	case SortAtime.String():
		return SortAtime, nil
	case SortCtime.String():
		return SortCtime, nil
	case SortTimeVar.String():
		return SortTimeVar, nil
	case SortIntVar.String():
		return SortIntVar, nil
	case SortStringVar.String():
		return SortStringVar, nil
	}

	return SortDefault, errInvalidSort.Fmt(arg)
}
