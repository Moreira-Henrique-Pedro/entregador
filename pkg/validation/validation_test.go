package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errInvalid = errors.New("invalid")

func TestRequired(t *testing.T) {
	assert.NoError(t, Required(errInvalid, "name", "Ana"))

	for _, value := range []string{"", "   "} {
		err := Required(errInvalid, "name", value)
		assert.ErrorIs(t, err, errInvalid)
		assert.EqualError(t, err, "invalid: name is required")
	}
}

func TestFirst(t *testing.T) {
	first, second := errors.New("first"), errors.New("second")

	assert.NoError(t, First())
	assert.NoError(t, First(nil, nil))
	assert.Equal(t, first, First(nil, first, second))
}
