package config

import "github.com/zoobz-io/check"

// Redis holds configuration for the Redis cache.
type Redis struct {
	Addr     string `env:"APP_REDIS_ADDR" default:"localhost:6379"`
	Password string `env:"APP_REDIS_PASSWORD" default:""`
	DB       int    `env:"APP_REDIS_DB" default:"0"`
}

// Validate checks that the Redis configuration is valid.
func (r Redis) Validate() error {
	return check.All(
		check.Str(r.Addr, "addr").Required().V(),
	).Err()
}
