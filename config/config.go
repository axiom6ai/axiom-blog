package config

import (
	"axiom-blog/pkg/util"
	"errors"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v9"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
	"sync"
)

/**
  @author: xichencx@163.com
  @date:2024/3/25
  @description:
**/

var (
	Conf       = &Config{}
	once       sync.Once
	configName = flag.String("config", "config", "配置文件名称，默认 config")
)

func ConfInit() {
	once.Do(func() {
		if err := env.Parse(Conf); err != nil {
			panic(err)
		}
		viper.AddConfigPath(path.Join(inferRootDir(), "/config"))
		viper.SetConfigName(*configName)
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				// 配置文件未找到错误；如果需要可以忽略
				panic(fmt.Sprintf("viper未找到配置文件:%v", err))
			}

			panic(fmt.Sprintf("viper读取配置文件失败:%v", err))
		}

		if err := viper.Unmarshal(Conf); err != nil {
			panic(fmt.Sprintf("viper序列化到structure失败:%v", err))
		}
	})
}

// inferRootDir 递归推导项目根目录
func inferRootDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var infer func(d string) string
	infer = func(d string) string {
		if d == "/" {
			panic(fmt.Sprintf("请确保在项目根目录或子目录下运行程序，当前在:%v", cwd))
		}

		if util.Exist(d + "/config") {
			return d
		}

		return infer(filepath.Dir(d))
	}

	return infer(cwd)
}

type Config struct {
	Host string `mapstructure:"host" env:"APP_HOST" envDefault:"localhost"`
	Port string `mapstructure:"port" env:"APP_PORT" envDefault:":8000"`
	Name string `mapstructure:"name" env:"APP_NAME" envDefault:"axiom-blog"`
	Desc struct {
		Keyword string `mapstructure:"keyword" env:"APP_KEYWORD" envDefault:"blog"`
		Desc    string `mapstructure:"desc" env:"APP_DESC" envDefault:"axiom-blog"`
	}
	Env     string `mapstructure:"env" env:"APP_ENV" envDefault:"development"`
	Version string `mapstructure:"version" env:"APP_VERSION" envDefault:"0.0.1"`
	Stdout  bool   `mapstructure:"stdout" env:"APP_LOG_STDOUT" envDefault:"true"`

	Token struct {
		SignKey string `mapstructure:"" env:"APP_SIGN_KEY" envDefault:"ethan"`
		Expires int    `mapstructure:"" env:"APP_EXPIRES" envDefault:"1"`
		Issuer  string `mapstructure:"" env:"APP_ISSUER" envDefault:"axiom"`
		Subject string `mapstructure:"" env:"APP_SUBJECT" envDefault:"blog"`
	}

	Postgres struct {
		Host          string `env:"POSTGRES_HOST" envDefault:"localhost"`
		User          string `env:"POSTGRES_USER" envDefault:"axiom"`
		Password      string `env:"POSTGRES_PASSWORD" envDefault:"password"`
		Database      string `env:"POSTGRES_DB" envDefault:"axiom"`
		Port          int    `env:"POSTGRES_PORT" envDefault:"5432"`
		SlowThreshold int    `mapstructure:"slowThreshold" env:"POSTGRES_SLOW_THRESHOLD" envDefault:"3"`
		MaxIdle       int    `mapstructure:"maxIdle" env:"POSTGRES_MAX_IDLE" envDefault:"2"`
		MaxConn       int    `mapstructure:"maxConn" env:"POSTGRES_MAX_CONN" envDefault:"10"`
		LogLevel      int    `mapstructure:"logLevel" env:"POSTGRES_LOG_LEVEL" envDefault:"4"`
	}

	Redis struct {
		Host     string `mapstructure:"host" env:"REDIS_HOST" envDefault:"localhost"`
		Password string `mapstructure:"password" env:"REDIS_PASSWORD" envDefault:"password"`
		Port     int    `mapstructure:"port" env:"REDIS_PORT" envDefault:"5432"`
	}
}
