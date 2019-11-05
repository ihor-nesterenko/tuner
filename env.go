package tuner

import (
	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
)

type envTuner struct {
}

func NewEnvTuner() Reader {
	return new(envTuner)
}

func (e envTuner) Read(target interface{}) error {
	if !isPointer(target) {
		return errors.New("target struct must be a pointer")
	}
	return env.Parse(target)
}
