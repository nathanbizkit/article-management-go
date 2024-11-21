package server

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_Server(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")
	lct := test.NewLocalTestContainer(t)

	t.Run("Start", func(t *testing.T) {
		t.Cleanup(func() {
			viper.Reset()
			t.Setenv("APP_MODE", "")
			t.Setenv("APP_PORT", "")
			t.Setenv("APP_TLS_PORT", "")
			t.Setenv("TLS_CERT_FILE", "")
			t.Setenv("TLS_KEY_FILE", "")
			t.Setenv("CORS_ALLOWED_ORIGINS", "")
			t.Setenv("AUTH_JWT_SECRET_KEY", "")
			t.Setenv("AUTH_COOKIE_DOMAIN", "")
			t.Setenv("DB_USER", "")
			t.Setenv("DB_PASS", "")
			t.Setenv("DB_HOST", "")
			t.Setenv("DB_PORT", "")
			t.Setenv("DB_NAME", "")
		})

		t.Setenv("APP_MODE", lct.Environ().AppMode)
		t.Setenv("APP_PORT", lct.Environ().AppPort)
		t.Setenv("APP_TLS_PORT", lct.Environ().AppTLSPort)
		t.Setenv("TLS_CERT_FILE", lct.Environ().TLSCertFile)
		t.Setenv("TLS_KEY_FILE", lct.Environ().TLSKeyFile)
		t.Setenv("CORS_ALLOWED_ORIGINS", "*")
		t.Setenv("AUTH_JWT_SECRET_KEY", lct.Environ().AuthJWTSecretKey)
		t.Setenv("AUTH_COOKIE_DOMAIN", lct.Environ().AuthCookieDomain)
		t.Setenv("DB_USER", lct.Environ().DBUser)
		t.Setenv("DB_PASS", lct.Environ().DBPass)
		t.Setenv("DB_HOST", lct.Environ().DBHost)
		t.Setenv("DB_PORT", lct.Environ().DBPort)
		t.Setenv("DB_NAME", lct.Environ().DBName)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		panicChan := make(chan bool, 0)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicChan <- true
					return
				}
			}()

			go Start()

			for {
				select {
				case <-ctx.Done():
					panicChan <- false

					// send interrupt signal to Start()
					syscall.Kill(os.Getpid(), syscall.SIGINT)
					return
				}
			}
		}()

		select {
		case actual := <-panicChan:
			// wait for server to shutdown
			time.Sleep(5 * time.Second)

			assert.False(t, actual, "Start() must not panic")
		}
	})
}
