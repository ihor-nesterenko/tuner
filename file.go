package tuner

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	yamlExt = "yaml"
	jsonExt = "json"
)

func NewFileTuner(path string) (Reader, error) {
	if path == "" {
		return nil, errors.New("path must not be empty")
	}

	fileTuner := &fileTuner{
		path: path,
	}

	return fileTuner, nil
}

type fileTuner struct {
	path    string
	rawFile []byte
}

func (r *fileTuner) Read(target interface{}) error {
	if !isPointer(target) {
		return errors.New("target struct must be a pointer")
	}

	file, err := ioutil.ReadFile(r.path)
	if err != nil {
		return errors.Wrapf(err, "failed to read file with path=%s", r.path)
	}

	err = r.unmarshal(file, target)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal cfg file to target")
	}
	return nil
}

func (r fileTuner) unmarshal(source []byte, target interface{}) error {

	switch ext := r.getExtension(); ext {
	case yamlExt:
		if err := yaml.Unmarshal(r.rawFile, &target); err != nil {
			return errors.Errorf("failed to unmarshal file into yaml")
		}
	case jsonExt:
		if err := json.Unmarshal(r.rawFile, &target); err != nil {
			return errors.Errorf("failed to unmarshal file into json")
		}
	default:
		return errors.Errorf("wrong extension: %s", ext)
	}

	return nil
}

func (r fileTuner) getExtension() string {
	pathSplitted := strings.Split(r.path, ".")
	return pathSplitted[len(pathSplitted)-1]
}
