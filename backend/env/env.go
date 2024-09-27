package env

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type env struct {
	loadOnce           sync.Once
	filePath           string
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
	Load() (ENVer, error)
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

func New(filePath string) ENVer {
	return &env{filePath: filePath}
}

func (o *env) Load() (ENVer, error) {
	var err error

	o.loadOnce.Do(func() {
		viper.SetConfigFile(o.filePath)

		if e := viper.ReadInConfig(); e != nil {
			viper.AutomaticEnv()
		}

		o.appMode = parseToStringFallback("APP_MODE", "develop")
		o.appPort = parseToStringFallback("APP_PORT", "8000")
		o.authCookieDomain = parseToStringFallback("AUTH_COOKIE_DOMAIN", "localhost")
		o.dbHost = parseToStringFallback("DB_HOST", "localhost")
		o.dbPort = parseToStringFallback("DB_PORT", "5432")
		o.corsAllowedOrigins = make([]string, 0)

		corsAllowedOrigins := parseToStringFallback("CORS_ALLOWED_ORIGINS", "", "*")
		if len(corsAllowedOrigins) > 0 {
			o.corsAllowedOrigins = strings.Split(strings.TrimSpace(corsAllowedOrigins), ",")
		}

		var e error
		o.authSecretKey, e = parseToString("AUTH_SECRET_KEY")
		if e != nil {
			err = e
			return
		}

		o.dbUser, e = parseToString("DB_USER")
		if e != nil {
			err = e
			return
		}

		o.dbPass, e = parseToString("DB_PASS")
		if e != nil {
			err = e
			return
		}

		o.dbName, e = parseToString("DB_NAME")
		if e != nil {
			err = e
			return
		}
	})

	return o, err
}

func (o *env) AppMode() string {
	return o.appMode
}

func (o *env) AppPort() string {
	return o.appPort
}

func (o *env) CorsAllowedOrigins() []string {
	return o.corsAllowedOrigins
}

func (o *env) AuthSecretKey() string {
	return o.authSecretKey
}

func (o *env) AuthCookieDomain() string {
	return o.authCookieDomain
}

func (o *env) DBUser() string {
	return o.dbUser
}

func (o *env) DBPass() string {
	return o.dbPass
}

func (o *env) DBHost() string {
	return o.dbHost
}

func (o *env) DBPort() string {
	return o.dbPort
}

func (o *env) DBName() string {
	return o.dbName
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
