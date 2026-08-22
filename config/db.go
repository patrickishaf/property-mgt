package config

import "fmt"

type DBConfig struct {
	Host                  string `envconfig:"DB_HOST" default:"localhost"`
	MaxIdleConnections    int    `envconfig:"DB_MAX_IDLE_CONNECTIONS" default:"45"`
	MaxIdleConnectionTime string `envconfig:"DB_MAX_IDLE_CONNECTION_TIME" default:"45"`
	MaxOpenConnections    int    `envconfig:"DB_MAX_OPEN_CONNECTIONS"`
	MigrationsDir         string `envconfig:"DB_MIGRATIONS_DIR" default:"internal/db/migrations"`
	Name                  string `envconfig:"DB_NAME"`
	Password              string `envconfig:"DB_PASSWORD"`
	Port                  int    `envconfig:"DB_PORT" default:"5432"`
	SSLMode               string `envconfig:"SSL_MODE" default:"disable"` // possible values: disable, require, verify-full
	Url                   string `envconfig:"DB_URL"`
	User                  string `envconfig:"DB_USER"`
}

func (this *DBConfig) GetConnectionString() string {
	if this.Url != "" {
		return this.Url
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", this.User, this.Password, this.Host, this.Port, this.Name, this.SSLMode)
}

func (this *DBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", this.Host, this.User, this.Password, this.Name, this.Port, this.SSLMode)
}
