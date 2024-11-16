package flash

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"io"
	"os"
	"testing"
	"time"
)

func runTests(t *testing.T, test TestFn, driver Driver, tc *DriverTestConfig, cc *ClientConfig, execSql ExecSqlFunc) {

	/* ------------------------------------------- INITIALIZATION TEST-------------------------------*/
	test(t, "Can be initialized", func(t *testing.T) {
		defer driver.Close()

		err := driver.Init(cc)
		if err != nil {
			t.Error(err)
		}
	}, true)

	test(t, "Can be closed", func(t *testing.T) {
		_ = driver.Init(cc)
		if err := driver.Close(); err != nil {
			t.Error(err)
		}
	}, true)

	test(t, "Listen keep running at least 3 seconds without error when no listeners exists", func(t *testing.T) {
		defer driver.Close()
		_ = driver.Init(cc)

		errChan := make(chan error, 1)
		go func() {
			c := make(DatabaseEventsChan)
			errChan <- driver.Listen(&c)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("Listen returned an error: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}, true)

	/* ------------------------------------------- RUNTIME TEST-------------------------------*/
	_ = driver.Init(cc)
	go func() {
		eventChan := make(DatabaseEventsChan)
		_ = driver.Listen(&eventChan)
	}()
	defer driver.Close()

	type ListenerConfigTestMap struct {
		Name           string
		listenerConfig *ListenerConfig
	}
	for ti, testEntry := range []ListenerConfigTestMap{
		{Name: "All fields", listenerConfig: &ListenerConfig{Table: "posts"}},
		{Name: "Partial fields", listenerConfig: &ListenerConfig{Table: "posts", Fields: []string{"active", "slug"}}},
		{Name: "All fields with conditions", listenerConfig: &ListenerConfig{
			Table:      "posts",
			Conditions: []*ListenerCondition{{Column: "active", Value: true}},
		}},
		{Name: "Partial fields with conditions", listenerConfig: &ListenerConfig{
			Table:      "posts",
			Fields:     []string{"id", "active"},
			Conditions: []*ListenerCondition{{Column: "slug", Value: nil}},
		}},
	} {
		for _, operation := range []Operation{
			OperationInsert,
			OperationUpdate,
			OperationDelete,
			OperationTruncate,
		} {
			test(t, "HandleOperationListenStart - "+testEntry.Name+" - "+operation.String(), func(t *testing.T) {
				errChan := make(chan error, 1)
				go func() {
					errChan <- driver.HandleOperationListenStart(fmt.Sprintf(`uid-%d`, ti), testEntry.listenerConfig, operation)
				}()

				ctx, cancel := context.WithTimeout(context.Background(), tc.RegistrationTimeout)
				defer cancel()

				select {
				case err := <-errChan:
					if err != nil {
						t.Errorf("HandleOperationListenStart returned an error: %v", err)
					}
				case <-ctx.Done():
					t.Errorf("HandleOperationListenStart timed out")
				}
			}, false)
			test(t, "HandleOperationListenStop - "+testEntry.Name+" - "+operation.String(), func(t *testing.T) {
				lc := &ListenerConfig{Table: "posts"}
				errChan := make(chan error, 1)
				go func() {
					errChan <- driver.HandleOperationListenStop(fmt.Sprintf(`uid-%d`, ti), lc, operation)
				}()

				ctx, cancel := context.WithTimeout(context.Background(), tc.RegistrationTimeout)
				defer cancel()

				select {
				case err := <-errChan:
					if err != nil {
						t.Errorf("HandleOperationListenStop returned an error: %v", err)
					}
				case <-ctx.Done():
					t.Errorf("HandleOperationListenStop timed out")
				}
			}, false)
		}
	}

}

type DriverTestConfig struct {
	ImagesVersions []string `default:"postgres,flash"`

	Database string
	Username string
	Password string

	ContainerCustomizers []testcontainers.ContainerCustomizer

	PropagationTimeout  time.Duration // Delay for event propagated from the DB to the eventsChan
	RegistrationTimeout time.Duration // Delay for OperationListenStart / HandleOperationListenStop

	Parallel bool
}

var DefaultDriverTestConfig = &DriverTestConfig{
	ImagesVersions: []string{
		// Standard PostgreSQL
		"docker.io/postgres:14-alpine",
		"docker.io/postgres:15-alpine",
		"docker.io/postgres:16-alpine",

		// PgVector
		// "docker.io/pgvector/pgvector:pg14",
		// "docker.io/pgvector/pgvector:pg15",
		// "docker.io/pgvector/pgvector:pg16",

		// PostGIS
		// "docker.io/postgis/postgis:14-3.4-alpine",
		// "docker.io/postgis/postgis:15-3.4-alpine",
		// "docker.io/postgis/postgis:16-3.4-alpine",

		// TimescaleDB
		// "docker.io/timescale/timescaledb:latest-pg14",
		// "docker.io/timescale/timescaledb:latest-pg15",
		// "docker.io/timescale/timescaledb:latest-pg16",
	},

	Database: "testdb",
	Username: "testuser",
	Password: "testpasword",

	PropagationTimeout:  time.Second,
	RegistrationTimeout: time.Second,

	Parallel: false, // DO NOT WORK
}

type TestFn func(t *testing.T, name string, f func(t *testing.T), restore bool)

func RunFlashDriverTestCase[T Driver](t *testing.T, config *DriverTestConfig, getDriverCb func() T) {
	if config == nil {
		config = DefaultDriverTestConfig
	}
	for _, image := range config.ImagesVersions {
		t.Run(image, func(t *testing.T) {

			driverInstance := getDriverCb()

			t.Parallel()

			dbCnx, conn, container := startPostgresContainer(t, config, image)
			logger := zerolog.New(os.Stdout).Level(zerolog.FatalLevel).With().Caller().Stack().Timestamp().Logger()
			clientConfig := &ClientConfig{
				DatabaseCnx: dbCnx,
				Driver:      driverInstance,
				Logger:      &logger,
			}

			testFn := func(t *testing.T, name string, f func(t *testing.T), restore bool) {
				t.Run(name, func(t *testing.T) {
					if restore {
						t.Cleanup(func() {
							restoreSnapshot(t, container)
						})
					}
					// USE LOCK in parallel to avoid restore snapshot during

					//if config.Parallel {
					//	t.Parallel() TODO
					//}

					f(t)
				})
			}

			runTests(t, testFn, driverInstance, config, clientConfig, conn)
		})
	}
}

type ExecSqlFunc func(t *testing.T, sql string) string

func startPostgresContainer(t *testing.T, config *DriverTestConfig, image string) (string, ExecSqlFunc, *postgres.PostgresContainer) {
	ctx := context.Background()

	customizers := config.ContainerCustomizers
	customizers = append(customizers,
		postgres.WithDatabase(config.Database),
		postgres.WithUsername(config.Username),
		postgres.WithPassword(config.Password),
		postgres.BasicWaitStrategies(),
	)

	container, err := postgres.Run(ctx, image, customizers...)
	if err != nil {
		t.Error(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %s", err)
		}
	})

	// explicitly set sslmode=disable because the container is not configured to use TLS
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Error(err)
	}

	execSql := func(t *testing.T, sql string) string {
		code, r, err := container.Exec(context.Background(), []string{"psql", "-U", config.Username, "-d", config.Database, "-c", sql})
		if err != nil {
			t.Error(err)
		}
		bytes, err := io.ReadAll(r)
		if err != nil {
			t.Error(err)
		}

		if code != 0 {
			t.Error(string(bytes))
		}
		return string(bytes)
	}

	// Bootstrap DB with default table
	execSql(t, bootstrapSql)

	// Create restore point for later
	err = container.Snapshot(context.Background(), postgres.WithSnapshotName("db-snapshot"))
	if err != nil {
		t.Error(err)
	}

	return connStr, execSql, container
}

const bootstrapSql = `
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(255),
    active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_active ON posts (active);

INSERT INTO posts (slug, active) VALUES
('slug1', true),
('slug2', false),
('slug3', true),
('slug4', false),
('slug5', true),
('slug6', false),
('slug7', true),
('slug8', false),
('slug9', true),
('slug10', false),
('slug11', true),
('slug12', false),
('slug13', true),
('slug14', false),
('slug15', true),
('slug16', false),
('slug17', true),
('slug18', false),
('slug19', true),
(NULL, false)
`

func restoreSnapshot(t *testing.T, container *postgres.PostgresContainer) {
	ctx := context.Background()
	err := container.Restore(ctx)
	if err != nil {
		t.Fatalf("failed to restore snapshot: %v", err)
	}
}
