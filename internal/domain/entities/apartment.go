package entities

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
)

func SortApartments(apartments []string) []string {
	sorted := slices.Clone(apartments)
	slices.SortFunc(sorted, compareApartments)
	return slices.Compact(sorted)
}

func compareApartments(a, b string) int {
	numberA, restA, hasNumberA := splitLeadingNumber(a)
	numberB, restB, hasNumberB := splitLeadingNumber(b)

	switch {
	case hasNumberA && hasNumberB:
		return cmp.Or(cmp.Compare(numberA, numberB), strings.Compare(restA, restB))
	case hasNumberA:
		return -1
	case hasNumberB:
		return 1
	}
	return strings.Compare(a, b)
}

func splitLeadingNumber(apartment string) (int, string, bool) {
	end := strings.IndexFunc(apartment, func(r rune) bool { return r < '0' || r > '9' })
	if end == -1 {
		end = len(apartment)
	}
	number, err := strconv.Atoi(apartment[:end])
	if err != nil {
		return 0, apartment, false
	}
	return number, apartment[end:], true
}
