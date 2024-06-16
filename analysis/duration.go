package analysis

import (
	"github.com/paulmach/orb/geojson"
	"time"
)

// TrackDuration calculates the duration of a track in seconds.
func TrackDuration(feature geojson.Feature) (int, bool) {
	coordProps := feature.Properties.CoordinateProperties()

	times, ok := coordProps["times"].([]interface{})
	if !ok {
		return 0, false
	}
	if len(times) < 2 {
		return 0, true
	}

	var foundFirst bool
	var first time.Time
	for i := range times {
		t, ok := ParseSloppyRecentTime(times[i])
		if ok {
			foundFirst = true
			first = t
			break
		}
	}
	var foundLast bool
	var last time.Time
	for i := len(times) - 1; i >= 0; i-- {
		t, ok := ParseSloppyRecentTime(times[i])
		if ok {
			foundLast = true
			last = t
			break
		}
	}
	if !foundFirst || !foundLast {
		return 0, false
	}

	if first.After(last) {
		return 0, false
	}

	return int(last.Sub(first).Seconds()), true
}
