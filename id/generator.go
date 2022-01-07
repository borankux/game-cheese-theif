package id

import (
	"github.com/goombaio/namegenerator"
	"time"
)

type NameGenerator struct {
	Core namegenerator.Generator
}

func(g NameGenerator) NewID() string {
	return g.Core.Generate()
}

func NewNameGenerator() NameGenerator {
	seed := time.Now().UTC().UnixNano()
	return NameGenerator{
		Core: namegenerator.NewNameGenerator(seed),
	}
}
