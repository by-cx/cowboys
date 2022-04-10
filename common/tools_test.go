package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCowboyLoader(t *testing.T) {
	myself, enemies, err := CowboyLoader("../cowboys.json", "Bill")
	assert.Nil(t, err)
	assert.Equal(t, "Bill", myself.Name)
	assert.Equal(t, 4, len(enemies))

	_, enemies, err = CowboyLoader("../cowboys.json", "")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(enemies))
}
