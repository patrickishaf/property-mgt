package config

import "fmt"

type DBConfig struct {
	Host          string `envconfig:"DB_HOST" default:"localhost"`
	MigrationsDir string `envconfig:"DB_MIGRATIONS_DIR" default:"internal/db/migrations"`
	Name          string `envconfig:"DB_NAME"`
	Password      string `envconfig:"DB_PASSWORD"`
	Port          int    `envconfig:"DB_PORT" default:"5432"`
	SSLMode       string `envconfig:"SSL_MODE" default:"disable"` // possible values: disable, require, verify-full
	Url           string `envconfig:"DB_URL"`
	User          string `envconfig:"DB_USER"`
}

func (this *DBConfig) GetConnectionString() string {
	if this.Url != "" {
		return this.Url
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", this.User, this.Password, this.Host, this.Port, this.Name, this.SSLMode)
}
