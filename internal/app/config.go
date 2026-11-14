package app

import (
	"os"

	"github.com/samber/lo"
)

// Config contains process configuration; modules define their own options.
type Config struct {
	HTTPAddr string
}

func ConfigFromEnv() Config {
	return Config{
		HTTPAddr: lo.CoalesceOrEmpty(os.Getenv("HTTP_ADDR"), "127.0.0.1:8080"),
	}
}
