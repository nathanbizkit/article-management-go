package container

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/nathanbizkit/article-management/util"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type LocalTestContainer struct {
	network              string
	pool                 *dockertest.Pool
	db                   *sql.DB
	dbName               string
	dbContainer          *dockertest.Resource
	dbMigrationContainer *dockertest.Resource
}

// NewLocalTestContainer creates a new local test container
func NewLocalTestContainer() (*LocalTestContainer, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		err = fmt.Errorf("failed to construct pool: %w", err)
		return nil, err
	}

	// network
	networkName := "article-management-app"
	network, err := createNetwork(networkName, pool)
	if err != nil {
		err = fmt.Errorf("failed to create network: %w", err)
		return nil, err
	}

	// db
	dbResource, err := createPostgresDB(pool, network)
	if err != nil {
		pool.Client.RemoveNetwork(networkName)
		err = fmt.Errorf("failed to create db container: %w", err)
		return nil, err
	}

	closeResources := func() {
		dbResource.Close()
		pool.Client.RemoveNetwork(networkName)
	}

	err = pool.Client.Ping()
	if err != nil {
		closeResources()
		err = fmt.Errorf("failed to connect to Docker: %w", err)
		return nil, err
	}

	db, err := getDBConnectionPool(
		pool,
		fmt.Sprintf(
			"postgres://root:password@%s/app_test?sslmode=disable",
			dbResource.GetHostPort("5432/tcp"),
		),
	)
	if err != nil {
		closeResources()
		err = fmt.Errorf("failed to connect to database: %w", err)
		return nil, err
	}

	log.Printf("db container: %s\n", dbResource.Container.Name)

	// db migration
	dbUrl := fmt.Sprintf(
		"postgres://root:password@%s:%s/app_test?sslmode=disable",
		strings.Trim(dbResource.Container.Name, "/"), "5432",
	)

	tempDir, err := os.MkdirTemp("", "migrations")
	if err != nil {
		closeResources()
		err = fmt.Errorf("failed to create temp dir: %w", err)
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	err = copyDir(filepath.Join(util.Root, "./db/migrations"), tempDir)
	if err != nil {
		closeResources()
		err = fmt.Errorf("failed to copy files to migrations folder: %w", err)
		return nil, err
	}

	migrationResource, err := createMigration(pool, network, dbUrl, tempDir)
	if err != nil {
		closeResources()
		err = fmt.Errorf("failed to create db migration container: %w", err)
		return nil, err
	}

	err = migrateDB(pool, migrationResource, dbUrl)
	if err != nil {
		closeResources()
		err = fmt.Errorf("failed to migrate db: %w", err)
		return nil, err
	}

	log.Printf("migration container: %s\n", migrationResource.Container.Name)

	// set db schema
	_, err = db.Exec(`SET search_path TO article_management`)
	if err != nil {
		closeResources()
		err = fmt.Errorf("failed to set db schema: %w", err)
		return nil, err
	}

	return &LocalTestContainer{
		network:              networkName,
		pool:                 pool,
		db:                   db,
		dbName:               dbResource.Container.Name,
		dbContainer:          dbResource,
		dbMigrationContainer: migrationResource,
	}, nil
}

// DB returns test database connection pool
func (l *LocalTestContainer) DB() *sql.DB {
	return l.db
}

// Close purges and closes all test containers
func (l *LocalTestContainer) Close() error {
	err := l.dbContainer.Close()
	if err != nil {
		err = fmt.Errorf("failed to purge db resource, please remove container manually")
		return err
	}

	err = l.pool.Client.RemoveNetwork(l.network)
	if err != nil {
		err = fmt.Errorf("failed to remove network: %s", err)
		return err
	}

	return nil
}

func createNetwork(name string, pool *dockertest.Pool) (*docker.Network, error) {
	network, err := findNetwork(name, pool)
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

func findNetwork(name string, pool *dockertest.Pool) (*docker.Network, error) {
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
			"POSTGRES_USER=root",
			"POSTGRES_PASSWORD=password",
			"POSTGRES_DB=app_test",
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
