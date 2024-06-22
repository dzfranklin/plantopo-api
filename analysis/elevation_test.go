package analysis

import (
	"context"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestElevationServiceClusterSmokeTest(t *testing.T) {
	t.Skip("skipping smoke test")
	ctx := context.Background()
	svc := NewElevationService("https://elevation.dfranklin.dev/api/v1")
	res, err := svc.QueryElevations(ctx, orb.LineString{{-105.6732, 40.0968}})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Greater(t, res[0], 2900.0)
	require.Less(t, res[0], 3000.0)
}
