package tracks

import (
	"context"
	"github.com/dzfranklin/plantopo-api/testsupport"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func newSubject(t *testing.T) *Repo {
	t.Helper()
	_, r := newSubjectWithDriver(t)
	return r
}

func newSubjectWithDriver(t *testing.T) (*riverpgxv5.Driver, *Repo) {
	t.Helper()

	pool := testsupport.NewDB(t)
	t.Cleanup(func() {
		pool.Close()
	})

	driver := riverpgxv5.New(pool)

	riverClient, err := river.NewClient[pgx.Tx](driver, &river.Config{})
	if err != nil {
		t.Fatal(err)
	}

	return driver, NewRepo(pool, riverClient)
}

func TestListOrderByTimeEmpty(t *testing.T) {
	r := newSubject(t)
	tracks, err := r.ListOrderByTime(context.Background(), "user_1")
	require.NoError(t, err)
	assert.Equal(t, 0, len(tracks))
}

func TestImportEnqueues(t *testing.T) {
	ctx := context.Background()
	driver, r := newSubjectWithDriver(t)

	id, err := r.Import(ctx, "user_1", "file.gpx", []byte("data"))
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	idInt, err := strconv.ParseInt(id[len("trackimport_"):], 10, 64)
	require.NoError(t, err)

	rivertest.RequireInserted(ctx, t, driver, &ImportWorkerArgs{Id: idInt}, nil)
}
