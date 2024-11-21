package env

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/viper"
)

// ENV definition
type ENV struct {
	AppMode            string   `mapstructure:"APP_MODE"`
	AppPort            string   `mapstructure:"APP_PORT"`
	AppTLSPort         string   `mapstructure:"APP_TLS_PORT"`
	TLSCertFile        string   `mapstructure:"TLS_CERT_FILE"`
	TLSKeyFile         string   `mapstructure:"TLS_KEY_FILE"`
	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	AuthJWTSecretKey   string   `mapstructure:"AUTH_JWT_SECRET_KEY"`
	AuthCookieDomain   string   `mapstructure:"AUTH_COOKIE_DOMAIN"`
	DBUser             string   `mapstructure:"DB_USER"`
	DBPass             string   `mapstructure:"DB_PASS"`
	DBHost             string   `mapstructure:"DB_HOST"`
	DBPort             string   `mapstructure:"DB_PORT"`
	DBName             string   `mapstructure:"DB_NAME"`
	TLSEnabled         bool
	IsDevelopment      bool
}

// Parse loads environment variables either from .env or environment directly and returns a new env
func Parse(envFile string) (*ENV, error) {
	viper.SetConfigType("env")

	if envFile == "" {
		viper.AutomaticEnv()
	} else {
		viper.SetConfigFile(envFile)

		err := viper.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}

	// viper does not read from environment directly
	// it needs to call SetDefault to allow it to bind
	// directly from environment even if it must not be empty
	// which we can check later
	viper.SetDefault("APP_MODE", "develop")
	viper.SetDefault("APP_PORT", "8000")
	viper.SetDefault("APP_TLS_PORT", "8443")
	viper.SetDefault("TLS_CERT_FILE", "")
	viper.SetDefault("TLS_KEY_FILE", "")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	viper.SetDefault("AUTH_JWT_SECRET_KEY", "")
	viper.SetDefault("AUTH_COOKIE_DOMAIN", "localhost")
	viper.SetDefault("DB_USER", "")
	viper.SetDefault("DB_PASS", "")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_NAME", "")

	environ := ENV{}
	err := viper.Unmarshal(&environ)
	if err != nil {
		return nil, err
	}

	err = validation.ValidateStruct(&environ,
		validation.Field(
			&environ.AppMode,
			validation.In("test", "testing", "dev", "develop", "prod", "production"),
		),
		validation.Field(
			&environ.AppPort,
			is.Digit,
		),
		validation.Field(
			&environ.AppTLSPort,
			is.Digit,
		),
		validation.Field(
			&environ.AuthJWTSecretKey,
			validation.Required,
		),
		validation.Field(
			&environ.DBUser,
			validation.Required,
		),
		validation.Field(
			&environ.DBPass,
			validation.Required,
		),
		validation.Field(
			&environ.DBName,
			validation.Required,
		),
	)
	if err != nil {
		return nil, err
	}

	if len(environ.CORSAllowedOrigins) != 0 {
		var allowedAllOrigins bool
		for _, origin := range environ.CORSAllowedOrigins {
			if origin == "*" {
				allowedAllOrigins = true
				break
			}
		}

		if allowedAllOrigins {
			environ.CORSAllowedOrigins = []string{}
		}
	}

	environ.TLSEnabled = environ.TLSCertFile != "" && environ.TLSKeyFile != ""

	environ.IsDevelopment = environ.AppMode == "dev" || environ.AppMode == "develop" ||
		environ.AppMode == "test" || environ.AppMode == "testing"

	return &environ, nil
}
