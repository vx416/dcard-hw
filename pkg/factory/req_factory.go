package factory

import (
	apiv1 "github.com/vx416/dcard-work/pkg/api/v1"
	gofactory "github.com/vx416/gogo-factory"
	"github.com/vx416/gogo-factory/attr"
	"github.com/vx416/gogo-factory/genutil"
)

var GuardianAnimalRequest = gofactory.New(
	&apiv1.GetGuardianAnimalRequest{},
	attr.Str("Name", genutil.RandName(3)),
)
