package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/vx416/dcard-work/pkg/limiter"
	"github.com/vx416/dcard-work/pkg/logging"
	"github.com/vx416/dcard-work/pkg/server"
	"go.uber.org/fx"
	"gopkg.in/yaml.v2"
)

var _config = &Config{}

// Get get global config
func Get() *Config {
	return _config
}

var once sync.Once

// Init init a config
func Init() (*Config, error) {
	var err error
	once.Do(func() {
		configPath := os.Getenv("CONFIG_PATH")
		if configPath == "" {
			_, f, _, _ := runtime.Caller(0)
			dir := filepath.Dir(f)
			configPath = filepath.Join(dir, "../../configs/")
		}
		configFile := os.Getenv("CONFIG_FILE")
		if configFile == "" {
			configFile = "app.yaml"
		}
		configPath = filepath.Join(configPath, configFile)

		_, fileErr := os.Stat(configPath)
		if fileErr == nil {
			err = initFromYaml(configPath)
			if os.Getenv("DATA_PATH") != "" {
				_config.DataPath = os.Getenv("DATA_PATH")
			}
			fmt.Println("TESTINGINGINGING", os.Getenv("DATA_PATH"))
			return
		} else if os.IsNotExist(fileErr) {
			err = initFromEnv()
			return
		}
		err = fileErr
	})
	if err != nil {
		return nil, err
	}
	return _config, nil
}

func initFromYaml(cfgPath string) error {
	file, err := os.Open(cfgPath)
	if err != nil {
		return err
	}

	cfgData, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(cfgData, &_config)
}

func initFromEnv() error {
	return nil
}

type Config struct {
	fx.Out

	DataPath string          `yaml:"data_path" env:"DATA_PATH"`
	Limiter  *limiter.Config `yaml:"limiter" env:",prefix=LIMITER_"`
	Log      *logging.Config `yaml:"log" env:",prefix=LOG_"`
	Server   *server.Config  `yaml:"server" env:",prefix=SERVER_"`
}

func (c *Config) ProvideInfra() fx.Option {
	return fx.Options(
		fx.Supply(*c),
		fx.Provide(
			limiter.New,
			logging.New,
			server.RunWith,
		),
	)
}
