package testsupport

import (
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"os"
	"testing"
)

func NewDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("--- Skipping integration tests as RUN_INTEGRATION_TESTS environment variable not set ---")
	}

	gm := newMigrator("migrations", db.MigrationsFS)
	config := pgtestdb.Custom(t, pgtestdb.Config{
		DriverName: "pgx",
		User:       "postgres",
		Password:   "password",
		Host:       "localhost",
		Port:       "5433",
		Options:    "sslmode=disable",
	}, gm)
	pool, err := db.NewPool(config.URL())
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}
