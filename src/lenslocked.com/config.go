package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/kelseyhightower/envconfig"
)

type PostgresConfig struct {
	Host        string `envconfig:"PGHOST"`
	Port        int    `envconfig:"PGPORT"`
	User        string `envconfig:"PGUSER"`
	Password    string `envconfig:"PGPASSWORD"`
	Name        string `envconfig:"PGNAME"`
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

type Config struct {
	Port     int
	Env      string
	Pepper   string
	HMACKey  string `split_words:"true"`
	Database PostgresConfig
}

func NewConfig(configRequired bool) Config {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatal(err)
	}

	spew.Dump(c)
	return c
}

func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "dev",
		Pepper:   "secret-random-string",
		HMACKey:  "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "eitah",
		Password: "your-password",
		Name:     "lenslocked_dev",
	}
}

func (c PostgresConfig) Dialect() string {
	return "postgres"
}

func (c PostgresConfig) ConnectionInfo() string {
	return c.DatabaseURL
	// if c.Password == "" {
	// 	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Name)
	// }
	// return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Name)
}
