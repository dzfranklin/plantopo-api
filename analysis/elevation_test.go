package analysis

import (
	"context"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestElevationServiceLocalSmokeTest(t *testing.T) {
	t.Skip("skipping smoke test")
	ctx := context.Background()
	svc := NewElevationService("http://localhost:3000")
	res, err := svc.QueryElevations(ctx, orb.LineString{{-105.6732, 40.0968}})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Greater(t, res[0], int32(2900))
	require.Less(t, res[0], int32(3000))
}

func TestElevationServiceClusterSmokeTest(t *testing.T) {
	t.Skip("skipping smoke test")
	ctx := context.Background()
	svc := NewElevationService("http://elevation.pt.svc.cluster.local")
	res, err := svc.QueryElevations(ctx, orb.LineString{{-105.6732, 40.0968}})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Greater(t, res[0], int32(2900))
	require.Less(t, res[0], int32(3000))
}
