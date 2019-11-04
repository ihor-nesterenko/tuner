package tuner

import (
	"encoding/json"
	"io/ioutil"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Extension int8

const (
	Yaml Extension = iota
	Json
)

type FileTuner interface {
	Read() error
	Unmarshal(target interface{}, extension Extension) error
}

func NewFileTuner(path string) (FileTuner, error) {
	fileTuner := &fileTuner{
		path: path,
	}
	if err := fileTuner.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate input parameters")
	}

	return fileTuner, nil
}

type fileTuner struct {
	path    string
	rawFile []byte
}

func (r *fileTuner) Read() error {
	file, err := ioutil.ReadFile(r.path)
	if err != nil {
		return errors.Wrapf(err, "failed to read file with path=%s", r.path)
	}

	r.rawFile = file
	return nil
}

func (r fileTuner) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.path, validation.Required),
	)
}

func (r fileTuner) Unmarshal(target interface{}, extension Extension) error {
	switch extension {
	case Yaml:
		if err := yaml.Unmarshal(r.rawFile, &target); err != nil {
			return errors.Errorf("failed to unmarshal %d", Yaml)
		}
	case Json:
		if err := json.Unmarshal(r.rawFile, &target); err != nil {
			return errors.Errorf("failed to unmarshal %d", Json)
		}
	default:
		return errors.Errorf("wrong extension: %d", extension)
	}

	return nil
}
