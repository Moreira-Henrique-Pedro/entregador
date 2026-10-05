package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortApartments(t *testing.T) {
	assert.Equal(t, []string{"2", "63", "101", "101A", "B2"}, SortApartments([]string{"101", "63", "B2", "101A", "2", "63"}))
	assert.Empty(t, SortApartments(nil))
}
