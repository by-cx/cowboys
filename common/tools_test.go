package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCowboyLoader(t *testing.T) {
	myself, enemies, err := CowboyLoader("../cowboys.js", "Bill")
	assert.Nil(t, err)
	assert.Equal(t, "Bill", myself.Name)
	assert.Equal(t, 4, len(enemies))

	_, enemies, err = CowboyLoader("../cowboys.js", "")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(enemies))
}
