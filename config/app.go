package config

type AppConfig struct {
	AccessTokenExpirationMinutes int `envconfig:"JWT_AT_EXPIRATION_MINS"`
	// AllowedOrigins lists the origins allowed to call the API from a browser,
	// comma-separated (e.g. "https://app.example.com,http://localhost:8082").
	AllowedOrigins              string `envconfig:"ALLOWED_ORIGINS"`
	AppName                     string `envconfig:"APP_NAME"`
	CheckoutExpiryMinutes       int    `envconfig:"CHECKOUT_EXPIRY_MINUTES" default:"30"`
	Env                         string `envconfig:"APP_ENV" default:"development"`
	JobBackoffTimeSeconds       int    `envconfig:"JOB_BACKOFF_TIME_SECONDS" default:"10"`
	JWTSecretKey                string `envconfig:"JWT_SECRET_KEY"`
	MaxJobAttemptCount          int    `envconfig:"MAX_JOB_ATTEMPT_COUNT" default:"4"`
	OTPTimeoutSeconds           int    `envconfig:"OTP_TIMEOUT_SECONDS" default:"180"`
	PaymentCurrency             string `envconfig:"PAYMENT_CURRENCY" default:"NGN"`
	PwdResetExpirationSeconds   int    `envconfig:"PASSWORD_RESET_EXPIRATION_SECONDS"`
	Port                        int    `envconfig:"PORT"`
	RefreshTokenExpirationHours int    `envconfig:"JWT_RT_EXPIRATION_HOURS"`
	WebAppURL                   string `envconfig:"WEB_APP_URL"`
	WorkerPoolCapacity          int    `envconfig:"WORKER_POOL_CAPACITY" default:"100"`
}
