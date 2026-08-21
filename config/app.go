package config

type AppConfig struct {
	AccessTokenExpirationMinutes int    `envconfig:"JWT_AT_EXPIRATION_MINS"`
	AppName                      string `envconfig:"APP_NAME"`
	GeohashPrecision             uint   `envconfig:"GEOHASH_PRECISION" default:"9"`
	JobBackoffTimeSeconds        int    `envconfig:"JOB_BACKOFF_TIME_SECONDS" default:"10"`
	JWTSecretKey                 string `envconfig:"JWT_SECRET_KEY"`
	MaxJobAttemptCount           int    `envconfig:"MAX_JOB_ATTEMPT_COUNT" default:"4"`
	OTPTimeoutSeconds            int    `envconfig:"OTP_TIMEOUT_SECONDS" default:"180"`
	PwdResetExpirationSeconds    int    `envconfig:"PASSWORD_RESET_EXPIRATION_SECONDS"`
	Port                         int    `envconfig:"PORT"`
	RefreshTokenExpirationHours  int    `envconfig:"JWT_RT_EXPIRATION_HOURS"`
	SocketReadBufferSize         int    `envconfig:"SOCKET_READ_BUFFER_SIZE"`
	SocketWriteBufferSize        int    `envconfig:"SOCKET_WRITE_BUFFER_SIZE"`
	WebAppURL                    string `envconfig:"WEB_APP_URL"`
	WebAppPasswordResetRoute     string `envconfig:"WEB_APP_PASSWORD_RESET_ROUTE"`
}
