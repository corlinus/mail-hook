package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel        logrus.Level      `yaml:"log_level" default:"info"`
	Addr            string            `yaml:"addr" default:":25"`
	Domain          string            `yaml:"domain" default:"localhost"`
	ReadTimeout     int               `yaml:"read_timeout" default:"10"`
	WriteTimeout    int               `yaml:"write_timeout" default:"10"`
	MaxMessageBytes int               `yaml:"max_message_bytes" default:"10485760"`
	MaxRecipients   int               `yaml:"max_recipients" default:"50"`
	AllowDomains    map[string]string `yaml:"allow_domains"`
	SpoolDir        string            `yaml:"spool_dir" default:"spool"`
	SendReties      int               `yaml:"send_retries" default:"5"`
	SpoolThreads    int               `yaml:"spool_threads" default:"5"`

	wg     *sync.WaitGroup    `yaml:"-"`
	ctx    context.Context    `yaml:"-"`
	cancel context.CancelFunc `yaml:"-"`
}

func LoadCfg(fname string) (*Config, error) {
	cfg := &Config{wg: &sync.WaitGroup{}}
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	// expand environment variables
	data = []byte(os.ExpandEnv(string(data)))

	err = setDefaults(cfg)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// setDefaults - sets default values for struct fields interpolating value of tag `default`. This could be skepped by setting noDefalt tag to `+`
func setDefaults(ptr any) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("Not a pointer")
	}

	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()
	var err error
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}

		f := v.Field(i)
		ftag := t.Field(i).Tag

		if ftag.Get("noDefault") == "+" {
			continue
		} else if f.Kind() == reflect.Struct {
			err = setDefaults(f.Addr().Interface())
			if err != nil {
				return err
			}
		} else if f.Kind() == reflect.Pointer && reflect.ValueOf(f).Kind() == reflect.Struct {
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}
			err = setDefaults(f.Interface())
			if err != nil {
				return err
			}
		} else if defaultVal, ok := ftag.Lookup("default"); ok {
			v := reflect.New(f.Type())
			err = yaml.Unmarshal([]byte(defaultVal), v.Interface())
			if err != nil {
				return err
			}
			f.Set(v.Elem())
		}
	}
	return nil
}
