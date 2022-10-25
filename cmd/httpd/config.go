package main

import (
	"fmt"
	"strings"

	"github.com/ardanlabs/conf/v3"
)

// config: defaults < config file < environment < command line options
// defaults: struct
// config file: YAML, TOML
// environment: os.Getenv
// command line: flag

// outside: viper + cobra

type Config struct {
	Addr string `conf:"default::8080,env:ADDR"`
	DSN  string `conf:"default:host=localhost user=postgres password=s3cr3t sslmode=disable,env:DSN"`
}

func loadConfig() (Config, error) {
	var c Config
	if _, err := conf.Parse("UNTER", &c); err != nil {
		return Config{}, err
	}

	if err := c.Validate(); err != nil {
		return Config{}, err
	}

	return c, nil
}

func (c Config) Validate() error {
	if err := validAddr(c.Addr); err != nil {
		return fmt.Errorf("bad port: %s", err)
	}

	if c.DSN == "" {
		return fmt.Errorf("missing DSN")
	}

	return nil
}

func validAddr(addr string) error {
	i := strings.Index(addr, ":")
	if i == -1 {
		return fmt.Errorf("missing ':' in address")
	}

	var port int
	if _, err := fmt.Sscanf(addr[i+1:], "%d", &port); err != nil {
		return fmt.Errorf("bad port")
	}

	const maxPort = 65535
	if port < 0 || port > maxPort {
		return fmt.Errorf("port %d our of range [0,%d]", port, maxPort)
	}

	return nil
}
