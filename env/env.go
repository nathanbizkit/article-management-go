package env

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/viper"
)

type ENV struct {
	AppMode            string   `mapstructure:"APP_MODE"`
	AppPort            string   `mapstructure:"APP_PORT"`
	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	AuthJWTSecretKey   string   `mapstructure:"AUTH_JWT_SECRET_KEY"`
	AuthCookieDomain   string   `mapstructure:"AUTH_COOKIE_DOMAIN"`
	DBUser             string   `mapstructure:"DB_USER"`
	DBPass             string   `mapstructure:"DB_PASS"`
	DBHost             string   `mapstructure:"DB_HOST"`
	DBPort             string   `mapstructure:"DB_PORT"`
	DBName             string   `mapstructure:"DB_NAME"`
}

// Parse loads environment variables either from .env or environment directly and returns a new env
func Parse() (*ENV, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// viper does not read from environment directly
	// it needs to call SetDefault to allow it to bind
	// directly from environment even if it must not be empty
	// which we can check later
	viper.SetDefault("APP_MODE", "develop")
	viper.SetDefault("APP_PORT", "8000")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	viper.SetDefault("AUTH_JWT_SECRET_KEY", "")
	viper.SetDefault("AUTH_COOKIE_DOMAIN", "localhost")
	viper.SetDefault("DB_USER", "")
	viper.SetDefault("DB_PASS", "")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_NAME", "")

	e := &ENV{}
	if err := viper.Unmarshal(e); err != nil {
		return nil, err
	}

	err := validation.ValidateStruct(e,
		validation.Field(
			&e.AppMode,
			validation.In("dev", "develop", "prod", "production"),
		),
		validation.Field(
			&e.AppPort,
			is.Digit,
		),
		validation.Field(
			&e.AuthJWTSecretKey,
			validation.Required,
		),
		validation.Field(
			&e.DBUser,
			validation.Required,
		),
		validation.Field(
			&e.DBPass,
			validation.Required,
		),
		validation.Field(
			&e.DBName,
			validation.Required,
		),
	)

	return e, err
}
