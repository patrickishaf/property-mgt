package config

import "fmt"

type RedisConfig struct {
	DB       int    `envconfig:"REDIS_DB"`
	Host     string `envconfig:"REDIS_HOST"`
	Password string `envconfig:"REDIS_PASSWORD"`
	Port     int    `envconfig:"REDIS_PORT"`
	Protocol int    `envconfig:"REDIS_PROTOCOL"`
	Url      string `envconfig:"REDIS_URL"`
	User     string `envconfig:"REDIS_USER"`
}

func (this *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", this.Host, this.Port)
}

func (cfg *RedisConfig) GetURL() string {
	if cfg.Url != "" {
		return cfg.Url
	}
	return fmt.Sprintf("redis://%s:%s@%s:%d/%d", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
}
