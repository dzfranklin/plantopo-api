package testsupport

import (
	"context"
	"database/sql"
	"errors"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivermigrate"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // pgx driver

	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/common"
)

// New returns a [GolangMigrator], which is a pgtestdb.Migrator that
// uses golang-migrate to perform migrations.
//
// `migrationsDir` is the path to the directory containing migration files.
//
// You can configure the behavior of dbmate by passing Options:
//   - [WithFS] allows you to use an embedded filesystem.
func newMigrator(migrationsDir string, fs fs.FS) *GolangMigrator {
	return &GolangMigrator{
		MigrationsDir: migrationsDir,
		FS:            fs,
	}
}

// GolangMigrator is a pgtestdb.Migrator that uses golang-migrate to perform migrations.
//
// Because Hash() requires calculating a unique hash based on the contents of
// the migrations, database, this implementation only supports reading migration
// files from disk or an embedded filesystem.
//
// GolangMigrator does not perform any Verify() or Prepare() logic.
type GolangMigrator struct {
	// Where the migrations come from
	MigrationsDir string
	FS            fs.FS
}

func (gm *GolangMigrator) Hash() (string, error) {
	return common.HashDirs(gm.FS, "*.sql", gm.MigrationsDir)
}

func (gm *GolangMigrator) Migrate(
	ctx context.Context,
	db *sql.DB,
	templateConfig pgtestdb.Config,
) error {
	var err error

	var m *migrate.Migrate
	var d source.Driver
	if d, err = iofs.New(gm.FS, gm.MigrationsDir); err == nil {
		m, err = migrate.NewWithSourceInstance("iofs", d, templateConfig.URL())
	}
	if err != nil {
		return err
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			panic(err)
		}
	}(m)
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		err = nil
	}
	if err != nil {
		return err
	}

	rm := rivermigrate.New[*sql.Tx](riverdatabasesql.New(db), nil)
	_, err = rm.Migrate(ctx, rivermigrate.DirectionUp, nil)
	return err
}

// Prepare is a no-op method.
func (*GolangMigrator) Prepare(
	_ context.Context,
	_ *sql.DB,
	_ pgtestdb.Config,
) error {
	return nil
}

// Verify is a no-op method.
func (*GolangMigrator) Verify(
	_ context.Context,
	_ *sql.DB,
	_ pgtestdb.Config,
) error {
	return nil
}
