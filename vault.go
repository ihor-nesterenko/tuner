package tuner

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const vaultTag = "vault"

// VaultConfig is used to read secrets from Vault into given struct
type VaultConfig struct {
	*TLSConfig
	Address *string        `json:"address" yaml:"address"`
	Timeout *time.Duration `json:"timeout" yaml:"timeout"`

	Path  string `json:"path" yaml:"path"`
	Token string `json:"token" yaml:"token"`
}

func (v VaultConfig) validate() error {
	if v.Token == "" {
		return errors.New("vault token must not be empty")
	}
	if v.Path == "" {
		return errors.New("vault path to secrets must not be empty")
	}

	return nil
}

type TLSConfig struct {
	CACert        string `json:"ca_cert" yaml:"ca_cert"`
	CAPath        string `json:"ca_path" yaml:"ca_path"`
	ClientCert    string `json:"client_cert" yaml:"client_cert"`
	ClientKey     string `json:"client_key" yaml:"client_key" `
	TLSServerName string `json:"tls_server_name" yaml:"tls_server_name"`
	Insecure      bool   `json:"insecure" yaml:"insecure"`
}

func NewVaultTuner(vaultCfg VaultConfig) (Reader, error) {
	err := vaultCfg.validate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate vault configuration")
	}

	cfg := api.DefaultConfig()
	if cfg.Error != nil {
		return nil, errors.Wrap(cfg.Error, "failed to read vault config")
	}

	err = initCfg(cfg, vaultCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init vault config")
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(cfg.Error, "failed to create vault client")
	}

	client.SetToken(vaultCfg.Token)

	token := client.Auth().Token()
	_, err = token.LookupSelf()
	if err != nil {
		return nil, errors.Wrap(err, "wrong auth token")
	}

	return vaultTuner{
		client: client,
		path:   vaultCfg.Path,
	}, nil
}

func initCfg(clientCfg *api.Config, vaultCfg VaultConfig) error {
	if vaultCfg.TLSConfig != nil {
		err := clientCfg.ConfigureTLS(&api.TLSConfig{
			CACert:        vaultCfg.TLSConfig.CACert,
			CAPath:        vaultCfg.TLSConfig.CAPath,
			ClientCert:    vaultCfg.TLSConfig.ClientCert,
			ClientKey:     vaultCfg.TLSConfig.ClientKey,
			TLSServerName: vaultCfg.TLSConfig.TLSServerName,
			Insecure:      vaultCfg.TLSConfig.Insecure,
		})
		if err != nil {
			return errors.Wrap(err, "failed to configure TLS")
		}
	}

	if vaultCfg.Address != nil {
		clientCfg.Address = *vaultCfg.Address
	}

	if vaultCfg.Timeout != nil {
		clientCfg.Timeout = *vaultCfg.Timeout
	}

	return nil
}

type vaultTuner struct {
	client *api.Client
	path   string
}

func (v vaultTuner) Read(target interface{}) error {
	if !isPointer(target) {
		return errors.New("target struct must be a pointer")
	}

	secrets, err := v.getSecrets()
	if err != nil {
		return errors.Wrap(err, "failed to get secrets")
	}

	err = unmarshal(secrets, target)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal secrets to target")
	}
	return nil
}

func (v vaultTuner) getSecrets() (map[string]interface{}, error) {
	logical := v.client.Logical()
	response, err := logical.Read(v.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read secrets from ")
	}

	return response.Data, nil
}

func unmarshal(source map[string]interface{}, target interface{}) error {
	if len(source) == 0 {
		return nil
	}

	targetType := reflect.TypeOf(target).Elem()
	targetValue := reflect.ValueOf(target).Elem()

	for i := 0; i < targetType.NumField(); i++ {
		tag := targetType.Field(i).Tag.Get(vaultTag)
		if tag == "" {
			continue
		}

		rawValue, ok := source[tag]
		if !ok {
			continue
		}

		err := setField(targetType.Field(i).Type.Kind(), targetValue.Field(i), rawValue)
		if err != nil {
			return errors.Wrapf(err, "failed to set %s target field", tag)
		}
	}

	return nil
}

func setField(targetKind reflect.Kind, targetValue reflect.Value, value interface{}) error {
	//TODO: finish for all types
	//TODO: must return error if int/float value > than targetKind
	switch targetKind {
	case reflect.String:
		res, ok := value.(string)
		if !ok {
			return errors.New("failed to convert value to string")
		}
		targetValue.SetString(res)
	case reflect.Bool:
		res, ok := value.(bool)
		if !ok {
			return errors.New("failed to convert value to bool")
		}
		targetValue.SetBool(res)
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		raw, ok := value.(json.Number)
		if !ok {
			return errors.New("failed to convert value to json.Number")
		}
		res, err := raw.Int64()
		if err != nil {
			return errors.New("failed to convert json.Number to int64")
		}
		targetValue.SetInt(res)
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		raw, ok := value.(json.Number)
		if !ok {
			return errors.New("failed to convert value to json.Number")
		}
		res, err := raw.Float64()
		if err != nil {
			return errors.New("failed to convert json.Number to float64")
		}
		targetValue.SetFloat(res)
	default:
		return errors.New("i'm lazy, won't work with structures")
	}

	return nil
}
