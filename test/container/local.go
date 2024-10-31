package container

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
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
		err = fmt.Errorf("could not construct pool: %w", err)
		return nil, err
	}

	// network
	networkName := "article-management-app"
	network, err := createNetwork(networkName, pool)
	if err != nil {
		err = fmt.Errorf("could not create network: %w", err)
		return nil, err
	}

	// db
	dbResource, err := createPostgresDB(pool, network)
	if err != nil {
		err = fmt.Errorf("could not create db container: %w", err)
		return nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		err = fmt.Errorf("could not connect to Docker: %w", err)
		return nil, err
	}

	hostAndPort := dbResource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://root:password@%s/app_test?sslmode=disable", hostAndPort)

	log.Println("connecting to database on url: ", databaseUrl)

	testDB, err := testDBConnectivity(pool, databaseUrl)
	if err != nil {
		err = fmt.Errorf("could not connect to database: %w", err)
		return nil, err
	}

	log.Printf("db container: %s", dbResource.Container.Name)

	// db migration
	tempDir, err := os.MkdirTemp("", "_migrations")
	if err != nil {
		err = fmt.Errorf("could not create temp dir: %w", err)
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	copyDir("./db/script", tempDir)

	migrationResource, err := createMigration(pool, network, databaseUrl, tempDir)
	if err != nil {
		err = fmt.Errorf("could not create db migration container: %w", err)
		return nil, err
	}

	err = migrateDB(pool, databaseUrl, migrationResource)
	if err != nil {
		err = fmt.Errorf("could not create migrate: %w", err)
		return nil, err
	}

	log.Printf("migration container: %s", migrationResource.Container.Name)

	return &LocalTestContainer{
		network:              networkName,
		pool:                 pool,
		db:                   testDB,
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
		err = fmt.Errorf("could not purge db resource from test, please remove manually")
		return err
	}

	err = l.pool.Client.RemoveNetwork(l.network)
	if err != nil {
		err = fmt.Errorf("Could not remove network: %s", err)
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

func createPostgresDB(pool *dockertest.Pool, network *docker.Network) (*dockertest.Resource, error) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       "db_test",
		Repository: "postgres",
		Tag:        "16",
		Env: []string{
			"POSTGRES_USER=root",
			"POSTGRES_PASSWORD=password",
			"POSTGRES_DB=app_test",
			"TZ=UTC",
			"PGTZ=UTC",
			"listen_addresses = '*'",
		},
		NetworkID: network.ID,
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	return resource, err
}

func createMigration(pool *dockertest.Pool, network *docker.Network, dbUrl string, tempDir string) (*dockertest.Resource, error) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "migrate/migrate",
		Tag:        "latest",
		NetworkID:  network.ID,
		Cmd: []string{"-path", "/migrations",
			"-database", dbUrl,
			"-verbose", "up", "2"},
		Mounts: []string{
			fmt.Sprintf("%s:/migrations", tempDir),
		},
	}, func(config *docker.HostConfig) {
		config.NetworkMode = "bridge"
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	return resource, err
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

func testDBConnectivity(pool *dockertest.Pool, dbUrl string) (*sql.DB, error) {
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

func migrateDB(pool *dockertest.Pool, dbUrl string, migration *dockertest.Resource) error {
	pool.MaxWait = 120 * time.Second
	err := pool.Retry(func() error {
		_, err := migration.Exec(
			[]string{"migrate", "-path", "/migrations", "-database", dbUrl, "up", "2"},
			dockertest.ExecOptions{})
		return err
	})

	return err
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
