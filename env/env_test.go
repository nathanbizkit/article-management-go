package env

import (
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ENV(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	t.Run("Parse", func(t *testing.T) {
		t.Cleanup(func() {
			resetEnvironment(t)
		})

		tempDir, err := os.MkdirTemp("", "env")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		envContent := []byte(`
		APP_MODE=dev
		APP_PORT=8000
		APP_TLS_PORT=8443
		TLS_CERT_FILE="/certs/localCA.pem"
		TLS_KEY_FILE="/certs/localCA_unencrypted.key"
		CORS_ALLOWED_ORIGINS=http://localhost:8000,https://localhost:8443
		AUTH_JWT_SECRET_KEY=secret
		DB_USER=root
		DB_PASS=password
		DB_HOST=db
		DB_PORT=5432
		DB_NAME=app`)

		envFile := path.Join(tempDir, "local.env")
		err = os.WriteFile(envFile, envContent, 0o644)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			title            string
			envFile          string
			setEnvironmentFn func(t *testing.T)
			expectedEnviron  *ENV
			hasError         bool
		}{
			{
				"parse with env file: success",
				envFile,
				func(t *testing.T) {},
				&ENV{
					AppMode:     "dev",
					AppPort:     "8000",
					AppTLSPort:  "8443",
					TLSCertFile: "/certs/localCA.pem",
					TLSKeyFile:  "/certs/localCA_unencrypted.key",
					CORSAllowedOrigins: []string{
						"http://localhost:8000",
						"https://localhost:8443",
					},
					AuthJWTSecretKey: "secret",
					DBUser:           "root",
					DBPass:           "password",
					DBHost:           "db",
					DBPort:           "5432",
					DBName:           "app",
					TLSEnabled:       true,
					IsDevelopment:    true,
				},
				false,
			},
			{
				"parse without env file: success",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				&ENV{
					AppMode:     "dev",
					AppPort:     "8000",
					AppTLSPort:  "8443",
					TLSCertFile: "/certs/localCA.pem",
					TLSKeyFile:  "/certs/localCA_unencrypted.key",
					CORSAllowedOrigins: []string{
						"http://localhost:8000",
						"https://localhost:8443",
					},
					AuthJWTSecretKey: "secret",
					DBUser:           "root",
					DBPass:           "password",
					DBHost:           "db",
					DBPort:           "5432",
					DBName:           "app",
					TLSEnabled:       true,
					IsDevelopment:    true,
				},
				false,
			},
			{
				"parse: cors allowed origins contains star symbol",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv("CORS_ALLOWED_ORIGINS", "*")
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				&ENV{
					AppMode:            "dev",
					AppPort:            "8000",
					AppTLSPort:         "8443",
					TLSCertFile:        "/certs/localCA.pem",
					TLSKeyFile:         "/certs/localCA_unencrypted.key",
					CORSAllowedOrigins: []string{},
					AuthJWTSecretKey:   "secret",
					DBUser:             "root",
					DBPass:             "password",
					DBHost:             "db",
					DBPort:             "5432",
					DBName:             "app",
					TLSEnabled:         true,
					IsDevelopment:      true,
				},
				false,
			},
			{
				"parse: invalid app mode",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "wrong_mode")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				nil,
				true,
			},
			{
				"parse: invalid app port",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "wrong_port")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				nil,
				true,
			},
			{
				"parse: invalid app tls port",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "wrong_port")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				nil,
				true,
			},
			{
				"parse: no auth jwt secret key",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				nil,
				true,
			},
			{
				"parse: no db user",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				nil,
				true,
			},
			{
				"parse: no db pass",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "app")
				},
				nil,
				true,
			},
			{
				"parse: no db name",
				"",
				func(t *testing.T) {
					t.Setenv("APP_MODE", "dev")
					t.Setenv("APP_PORT", "8000")
					t.Setenv("APP_TLS_PORT", "8443")
					t.Setenv("TLS_CERT_FILE", "/certs/localCA.pem")
					t.Setenv("TLS_KEY_FILE", "/certs/localCA_unencrypted.key")
					t.Setenv(
						"CORS_ALLOWED_ORIGINS",
						"http://localhost:8000,https://localhost:8443",
					)
					t.Setenv("AUTH_JWT_SECRET_KEY", "secret")
					t.Setenv("DB_USER", "root")
					t.Setenv("DB_PASS", "password")
					t.Setenv("DB_HOST", "db")
					t.Setenv("DB_PORT", "5432")
					t.Setenv("DB_NAME", "")
				},
				nil,
				true,
			},
		}

		for _, tt := range tests {
			resetEnvironment(t)
			tt.setEnvironmentFn(t)

			actualEnviron, err := Parse(tt.envFile)

			assert.Equal(t, tt.expectedEnviron, actualEnviron, tt.title)

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}
		}
	})
}

func resetEnvironment(t *testing.T) {
	t.Helper()

	viper.Reset()
	t.Setenv("APP_MODE", "")
	t.Setenv("APP_PORT", "")
	t.Setenv("APP_TLS_PORT", "")
	t.Setenv("TLS_CERT_FILE", "")
	t.Setenv("TLS_KEY_FILE", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("AUTH_JWT_SECRET_KEY", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASS", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_NAME", "")
}
