package main

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

type PostgresConfig struct {
	Host        string `envconfig:"PGHOST"`
	Name        string `envconfig:"PGDATABASE"`
	Port        int    `envconfig:"PGPORT"`
	User        string `envconfig:"PGUSER"`
	Password    string `envconfig:"PGPASSWORD"`
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

type MailgunConfig struct {
	APIKey           string `envconfig:"MAILGUN_API_KEY"`
	PublicAPIKey     string `envconfig:"MAILGUN_PUBLIC_API_KEY"`
	Domain           string `envconfig:"MAILGUN_DOMAIN"`
	ElisEmailAddress string `envconfig:"ELIS_EMAIL_ADDRESS"`
}

type Config struct {
	Port          int
	Env           string
	Pepper        string
	HMACKey       string `split_words:"true"`
	Database      PostgresConfig
	Mailgun       MailgunConfig
	Bucket        string `envconfig:"CLOUD_BUCKET"`
	CredsFilePath string
}

func NewConfig(configRequired bool) Config {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatal(err)
	}

	return c
}

func DefaultConfig() Config {
	return Config{
		Port:          3000,
		Env:           "dev",
		Pepper:        "secret-random-string",
		HMACKey:       "secret-hmac-key",
		Database:      DefaultPostgresConfig(),
		Bucket:        "not-the-real-bucket",
		CredsFilePath: "./creds.json",
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
	// return c.DatabaseURL + "?sslmode=require" 	// this randomly just stopped working and IDK why.
	sslMode := "disable"
	if c.Host != "localhost" {
		// heroku requires ssl but localhost doesnt
		sslMode = "require"
	}

	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", c.Host, c.Port, c.User, c.Name, sslMode)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.Host, c.Port, c.User, c.Password, c.Name, sslMode)
}
