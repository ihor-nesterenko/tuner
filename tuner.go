package tuner

import "github.com/pkg/errors"

type Tuner interface {
	Reader
	FromVault(vaultCfg VaultConfig) error
	FromEnv()
	FromFile(path string) error
}

type defaultTuner struct {
	vaultTuner Reader
	envTuner   Reader
	fileTuner  Reader
}

func NewTuner() Tuner {
	return new(defaultTuner)
}

func (d defaultTuner) Read(target interface{}) (err error) {
	if !isPointer(target) {
		return errors.New("target struct must be a pointer")
	}

	if d.fileTuner != nil {
		err = d.fileTuner.Read(target)
		if err != nil {
			return errors.Wrap(err, "failed to read config from file")
		}
	}

	if d.envTuner != nil {
		err = d.envTuner.Read(target)
		if err != nil {
			return errors.Wrap(err, "failed to read config from environment")
		}

	}

	if d.vaultTuner != nil {
		err = d.vaultTuner.Read(target)
		if err != nil {
			return errors.Wrap(err, "failed to read config from vault")
		}
	}

	return
}

func (d *defaultTuner) FromVault(vaultCfg VaultConfig) error {
	vaultTuner, err := NewVaultTuner(vaultCfg)
	if err != nil {
		return err
	}

	d.vaultTuner = vaultTuner
	return nil
}

func (d *defaultTuner) FromEnv() {
	d.envTuner = NewEnvTuner()
}

func (d *defaultTuner) FromFile(path string) error {
	fileTuner, err := NewFileTuner(path)
	if err != nil {
		return err
	}

	d.fileTuner = fileTuner
	return nil
}
