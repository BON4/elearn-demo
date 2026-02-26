package config

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/caarlos0/env/v11"
	"gopkg.in/ini.v1"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT" default:"8080"`
	DBUrl    string `env:"DATABASE_URL"`
	RBBMQUrl string `env:"REBBITMQ_URL"`
}

func (c *Config) Validate() error {
	if c.DBUrl == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.RBBMQUrl == "" {
		return errors.New("REBBITMQ_URL is required")
	}

	if _, err := url.ParseRequestURI(c.DBUrl); err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	if _, err := url.ParseRequestURI(c.RBBMQUrl); err != nil {
		return fmt.Errorf("invalid REBBITMQ_URL: %w", err)
	}

	if c.HTTPPort == "" {
		return errors.New("HTTP_PORT is required")
	}

	return nil
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func LoadFromINI(path string) (*Config, error) {
	file, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		HTTPPort: file.Section("").Key("HTTP_PORT").MustString("8080"),
		DBUrl:    file.Section("").Key("DATABASE_URL").String(),
		RBBMQUrl: file.Section("").Key("REBBITMQ_URL").String(),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
