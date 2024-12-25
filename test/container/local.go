package container

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/nathanbizkit/article-management-go/env"
	"github.com/nathanbizkit/article-management-go/util"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	networkName = "article-management-testing-app"
	dbSchema    = "article_management"
	dbUser      = "root"
	dbPass      = "password"
	dbName      = "app_test"
)

type LocalTestContainer struct {
	network              string
	pool                 *dockertest.Pool
	environ              *env.ENV
	db                   *sql.DB
	dbName               string
	dbContainer          *dockertest.Resource
	dbMigrationContainer *dockertest.Resource
}

// NewLocalTestContainer creates a new local test container
func NewLocalTestContainer() (*LocalTestContainer, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to construct pool: %w", err)
	}

	// network
	network, err := createNetwork(pool, networkName)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	// db
	dbResource, err := createPostgresDB(pool, network)
	if err != nil {
		pool.Client.RemoveNetwork(networkName)
		return nil, fmt.Errorf("failed to create db container: %w", err)
	}

	closeResources := func() {
		dbResource.Close()
		pool.Client.RemoveNetwork(networkName)
	}

	err = pool.Client.Ping()
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	db, err := getDBConnectionPool(
		pool,
		fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			dbUser, dbPass, dbResource.GetHostPort("5432/tcp"), dbName,
		),
	)
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("db container: %s\n", dbResource.Container.Name)

	// db migration
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, strings.Trim(dbResource.Container.Name, "/"), "5432", dbName,
	)

	tempDir, err := os.MkdirTemp("", "migrations")
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	err = copyDir(filepath.Join(util.Root, "./db/migrations"), tempDir)
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to copy files to migrations folder: %w", err)
	}

	migrationResource, err := createMigration(pool, network, dbUrl, tempDir)
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to create db migration container: %w", err)
	}

	err = migrateDB(pool, migrationResource, dbUrl)
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	log.Printf("migration container: %s\n", migrationResource.Container.Name)

	// set db schema
	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", dbSchema))
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to set db schema: %w", err)
	}

	appPort, err := getFreePort()
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to get free app port: %w", err)
	}

	appTLSPort, err := getFreePort()
	if err != nil {
		closeResources()
		return nil, fmt.Errorf("failed to get free app tls port: %w", err)
	}

	// set env
	dbHostPort := strings.Split(dbResource.GetHostPort("5432/tcp"), ":")
	environ := &env.ENV{
		AppMode:          "test",
		AppPort:          strconv.Itoa(appPort),
		AppTLSPort:       strconv.Itoa(appTLSPort),
		AuthJWTSecretKey: "secretKey",
		DBUser:           dbUser,
		DBPass:           dbPass,
		DBHost:           dbHostPort[0],
		DBPort:           dbHostPort[1],
		DBName:           dbName,
		IsDevelopment:    true,
	}

	return &LocalTestContainer{
		network:              networkName,
		pool:                 pool,
		environ:              environ,
		db:                   db,
		dbName:               dbResource.Container.Name,
		dbContainer:          dbResource,
		dbMigrationContainer: migrationResource,
	}, nil
}

// Environ returns test env
func (l *LocalTestContainer) Environ() *env.ENV {
	return l.environ
}

// DB returns test database connection pool
func (l *LocalTestContainer) DB() *sql.DB {
	return l.db
}

// Close purges and closes all test containers
func (l *LocalTestContainer) Close() error {
	err := l.dbContainer.Close()
	if err != nil {
		return fmt.Errorf("failed to purge db resource, please remove container manually: %w", err)
	}

	err = l.pool.Client.RemoveNetwork(l.network)
	if err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}

	return nil
}

func createNetwork(pool *dockertest.Pool, name string) (*docker.Network, error) {
	network, err := findNetwork(pool, name)
	if err != nil {
		return nil, err
	}

	if network == nil {
		network, err = pool.Client.CreateNetwork(docker.CreateNetworkOptions{
			Name:           name,
			Driver:         "bridge",
			CheckDuplicate: true,
		})
		if err != nil {
			return nil, err
		}
	}

	return network, err
}

func findNetwork(pool *dockertest.Pool, name string) (*docker.Network, error) {
	networks, err := pool.Client.ListNetworks()
	if err != nil {
		return nil, err
	}

	for _, net := range networks {
		if net.Name == name {
			return &net, nil
		}
	}

	return nil, nil
}

func getDBConnectionPool(pool *dockertest.Pool, dbUrl string) (*sql.DB, error) {
	var db *sql.DB

	pool.MaxWait = 120 * time.Second
	err := pool.Retry(func() error {
		var err error

		db, err = sql.Open("postgres", dbUrl)
		if err != nil {
			return err
		}

		return db.Ping()
	})

	return db, err
}

func createPostgresDB(pool *dockertest.Pool, network *docker.Network) (*dockertest.Resource, error) {
	return pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16",
		NetworkID:  network.ID,
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPass),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
			"TZ=UTC",
			"PGTZ=UTC",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.NetworkMode = "bridge"
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
}

func createMigration(pool *dockertest.Pool, network *docker.Network, dbUrl, tempDir string) (*dockertest.Resource, error) {
	return pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "migrate/migrate",
		Tag:        "latest",
		NetworkID:  network.ID,
		Cmd:        []string{"-database", dbUrl, "-path", "/migrations", "up"},
	}, func(config *docker.HostConfig) {
		config.Mounts = []docker.HostMount{
			{
				Target: "/migrations",
				Source: tempDir,
				Type:   "bind",
			},
		}
		config.NetworkMode = "bridge"
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
}

func migrateDB(pool *dockertest.Pool, migration *dockertest.Resource, dbUrl string) error {
	pool.MaxWait = 120 * time.Second
	return pool.Retry(func() error {
		_, err := migration.Exec(
			[]string{"migrate", "-database", dbUrl, "-path", "/migrations", "up"},
			dockertest.ExecOptions{},
		)
		return err
	})
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, entry.Type()); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}

func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
