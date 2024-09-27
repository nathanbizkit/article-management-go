package env

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type env struct {
	appMode            string
	appPort            string
	corsAllowedOrigins []string
	authSecretKey      string
	authCookieDomain   string
	dbUser             string
	dbPass             string
	dbHost             string
	dbPort             string
	dbName             string
}

type ENVer interface {
	AppMode() string
	AppPort() string
	CorsAllowedOrigins() []string
	AuthSecretKey() string
	AuthCookieDomain() string
	DBUser() string
	DBPass() string
	DBHost() string
	DBPort() string
	DBName() string
}

func Load(filePath string) (ENVer, error) {
	e := &env{}

	viper.SetConfigFile(filePath)

	if err := viper.ReadInConfig(); err != nil {
		viper.AutomaticEnv()
	}

	e.appMode = parseToStringFallback("APP_MODE", "develop")
	e.appPort = parseToStringFallback("APP_PORT", "8000")
	e.authCookieDomain = parseToStringFallback("AUTH_COOKIE_DOMAIN", "localhost")
	e.dbHost = parseToStringFallback("DB_HOST", "localhost")
	e.dbPort = parseToStringFallback("DB_PORT", "5432")
	e.corsAllowedOrigins = make([]string, 0)

	corsAllowedOrigins := parseToStringFallback("CORS_ALLOWED_ORIGINS", "", "*")
	if len(corsAllowedOrigins) > 0 {
		e.corsAllowedOrigins = strings.Split(strings.TrimSpace(corsAllowedOrigins), ",")
	}

	var err error
	e.authSecretKey, err = parseToString("AUTH_SECRET_KEY")
	if err != nil {
		return nil, err
	}

	e.dbUser, err = parseToString("DB_USER")
	if err != nil {
		return nil, err
	}

	e.dbPass, err = parseToString("DB_PASS")
	if err != nil {
		return nil, err
	}

	e.dbName, err = parseToString("DB_NAME")
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *env) AppMode() string {
	return e.appMode
}

func (e *env) AppPort() string {
	return e.appPort
}

func (e *env) CorsAllowedOrigins() []string {
	return e.corsAllowedOrigins
}

func (e *env) AuthSecretKey() string {
	return e.authSecretKey
}

func (e *env) AuthCookieDomain() string {
	return e.authCookieDomain
}

func (e *env) DBUser() string {
	return e.dbUser
}

func (e *env) DBPass() string {
	return e.dbPass
}

func (e *env) DBHost() string {
	return e.dbHost
}

func (e *env) DBPort() string {
	return e.dbPort
}

func (e *env) DBName() string {
	return e.dbName
}

func parseToString(key string) (string, error) {
	value, ok := viper.Get(key).(string)
	if !ok || value == "" {
		return "", fmt.Errorf("$%s is not set", key)
	}
	return value, nil
}

func parseToStringFallback(key, fallback string, ignoreValues ...string) string {
	value, ok := viper.Get(key).(string)

	var isIgnored bool
	if len(ignoreValues) > 0 {
		for _, v := range ignoreValues {
			if value == v {
				isIgnored = true
				break
			}
		}
	}

	if !ok || value == "" || isIgnored {
		return fallback
	}

	return value
}
