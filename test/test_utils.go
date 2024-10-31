package test

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/nathanbizkit/article-management/test/container"
)

const englishCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var ltc *container.LocalTestContainer
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetLocalTestContainer returns a local test container
func GetLocalTestContainer(t *testing.T) *container.LocalTestContainer {
	t.Helper()

	if testing.Short() {
		return nil
	}

	ltc, err := container.NewLocalTestContainer()
	if err != nil {
		log.Fatal(err)
	}

	t.Cleanup(func() {
		ltc.Close()
	})

	return ltc
}

// RandomString returns a random string with x length in English and Numbers
func RandomString(t *testing.T, length int) string {
	t.Helper()

	b := make([]byte, length)
	for i := range b {
		b[i] = englishCharset[seededRand.Intn(len(englishCharset))]
	}

	return string(b)
}
