package config

type BachsConfig struct {
	APIBaseURL    string `envconfig:"BACHS_API_BASE_URL"`
	APIKey        string `envconfig:"BACHS_API_KEY"`
	WebhookSecret string `envconfig:"BACHS_WH_SECRET"`
}
