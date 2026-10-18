package config

type CpaasConfig struct {
	APIBaseURL    string `envconfig:"CPAAS_API_BASE_URL"`
	APIKey        string `envconfig:"CPAAS_API_KEY"`
	SenderAddress string `envconfig:"CPAAS_SENDER_ADDR"`
}
