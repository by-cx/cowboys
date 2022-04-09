package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Loads cowboy specs from the persistent storage based on his name.
// Return representation of myself, my enemies and error if there is any.
func CowboyLoader(path string, name string) (Cowboy, Cowboys, error) {
	cowboys := Cowboys{}

	body, err := ioutil.ReadFile(path)
	if err != nil {
		return Cowboy{}, Cowboys{}, fmt.Errorf("loading cowboys.js error: %v", err)
	}

	err = json.Unmarshal(body, &cowboys)
	if err != nil {
		return Cowboy{}, Cowboys{}, fmt.Errorf("parsing cowboys.js error: %v", err)
	}

	enemies := make(Cowboys)
	var myself Cowboy

	for _, cowboy := range cowboys {
		if cowboy.Name == name {
			myself = cowboy
		} else {
			enemies[cowboy.Name] = cowboy
		}
	}

	return myself, enemies, nil
}
