package app

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/vx416/dcard-work/pkg/config"
	"github.com/vx416/dcard-work/pkg/logging"
)

// GuardianAnimal animal
type GuardianAnimal struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func loadAnimals() (map[string]GuardianAnimal, error) {
	var (
		animals    []GuardianAnimal
		animalsMap = make(map[string]GuardianAnimal)
	)

	dataPath := config.Get().DataPath

	logging.Get().Debugf("load animals file from %s", dataPath)

	file, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}

	animalData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(animalData, &animals)
	if err != nil {
		return nil, err
	}

	for _, a := range animals {
		animalsMap[a.Name] = a
	}

	return animalsMap, nil
}
