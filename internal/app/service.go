package app

import (
	"context"

	"github.com/vx416/dcard-work/pkg/api/errors"
	"github.com/vx416/dcard-work/pkg/logging"
)

// Ringer consistent hash ring
type Ringer interface {
	AddNodes(nodesKey ...string) error
	GetNode(key string) (nodeKey string, err error)
	Nodes() []string
	String() string
}

// Servicer service layer interface
type Servicer interface {
	GetAnimal(ctx context.Context, base string) (GuardianAnimal, error)
}

type service struct {
	ring    Ringer
	animals map[string]GuardianAnimal
}

func (svc service) GetAnimal(ctx context.Context, base string) (GuardianAnimal, error) {
	var (
		animal     = GuardianAnimal{}
		animalName string
		ok         bool
		err        error
	)

	animalName, err = svc.ring.GetNode(base)
	if err != nil {
		return animal, errors.WithMessagef(errors.ErrInternalServerError, "svc: GetNode from ring failed, base:%s, err:%+v", base, err)
	}
	animal, ok = svc.animals[animalName]
	if !ok {
		return animal, errors.WithNewMsgf(errors.ErrInternalServerError, "service load incorrect animal, animal(%s) not in the memory", animalName)
	}

	return animal, nil
}

// New new servicer
func New(ringer Ringer) (Servicer, error) {
	animals, err := loadAnimals()
	if err != nil {
		return nil, err
	}

	if len(animals) == 0 {
		return nil, errors.New("animals data is empty")
	}

	for animalName := range animals {
		err = ringer.AddNodes(animalName)
		if err != nil {
			return nil, err
		}
	}
	logging.Get().Debugf("ring:%s", ringer.String())

	svc := service{
		ring:    ringer,
		animals: animals,
	}

	return svc, nil
}
