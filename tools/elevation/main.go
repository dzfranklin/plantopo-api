package main

import (
	"context"
	"fmt"
	"github.com/dzfranklin/plantopo-api/analysis"
	"github.com/paulmach/orb"
	"os"
	"strconv"
)

const endpoint = "http://elevation.pt.svc.cluster.local"

func usage() {
	println("Usage: elevation <longitude> <latitude>")
	os.Exit(1)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	lng, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		usage()
	}
	lat, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		usage()
	}

	svc := analysis.NewElevationService(endpoint)

	elevations, err := svc.QueryElevations(context.Background(), orb.LineString{{lng, lat}})
	if err != nil {
		println("Error querying elevations:", err.Error())
		os.Exit(1)
	}
	elevation := elevations[0]

	fmt.Printf("%d\n", elevation)
}
