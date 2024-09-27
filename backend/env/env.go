package env

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type env struct {
	loadOnce sync.Once
	isLoaded bool
	filePath string
	values   Values
}

type ENVer interface {
	Load() error
	Values() Values
}

func New(file string) ENVer {
	return &env{filePath: file}
}

func (o *env) Load() error {
	var err error

	o.loadOnce.Do(func() {
		o.isLoaded = true

		viper.SetConfigFile(o.filePath)

		if e := viper.ReadInConfig(); e != nil {
			viper.AutomaticEnv()
		}

		corsAllowedOrigins := parseToStringFallback("CORS_ALLOWED_ORIGINS", "", "*")
		corsAllowedOriginsArr := make([]string, 0)
		if len(corsAllowedOrigins) > 0 {
			corsAllowedOriginsArr = strings.Split(strings.TrimSpace(corsAllowedOrigins), ",")
		}

		authSecretKey, e := parseToString("AUTH_SECRET_KEY")
		if e != nil {
			err = e
			return
		}

		dbUser, e := parseToString("DB_USER")
		if e != nil {
			err = e
			return
		}

		dbPass, e := parseToString("DB_PASS")
		if e != nil {
			err = e
			return
		}

		dbHost, e := parseToString("DB_HOST")
		if e != nil {
			err = e
			return
		}

		dbName, e := parseToString("DB_NAME")
		if e != nil {
			err = e
			return
		}

		o.values = &values{
			appMode:            parseToStringFallback("APP_MODE", "develop"),
			appPort:            parseToStringFallback("APP_PORT", "8000"),
			corsAllowedOrigins: corsAllowedOriginsArr,
			authSecretKey:      authSecretKey,
			authCookieDomain:   parseToStringFallback("AUTH_COOKIE_DOMAIN", "localhost"),
			dbUser:             dbUser,
			dbPass:             dbPass,
			dbHost:             dbHost,
			dbName:             dbName,
		}
	})

	return err
}

func (o *env) Values() Values {
	if !o.isLoaded {
		o.Load()
	}
	return o.values
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
