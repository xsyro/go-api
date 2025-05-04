package config

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	Debug      bool `env:"DEBUG,required"`
	Db         dbConf
	Redis      redis
	AuthConfig authConfig
}

// Config defines the options that are used when connecting to a PostgreSQL instance
type dbConf struct {
	Host        string `env:"DB_HOST,required"`
	Port        string `env:"DB_PORT,required"`
	User        string `env:"DB_USER,required"`
	Pass        string `env:"DB_PASS,required"`
	DbName      string `env:"DB_NAME,required"`
	SSLMode     string `env:"SSL_MODE"`
	SSLCert     string `env:"SSL_CERT"`
	SSLKey      string `env:"SSL_KEY"`
	SSLRootCert string `env:"SSL_ROOT_CERT"`
}

type redis struct {
	Host string `env:"REDIS_HOST,required"`
	Port string `env:"REDIS_PORT,required"`
}

type authConfig struct {
	AccessSecret  string `env:"ACCESS_SECRET,required"`
	RefreshSecret string `env:"REFRESH_SECRET,required"`
}

func AppConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
