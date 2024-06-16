package analysis

import "time"

func ParseSloppyRecentTime(v interface{}) (time.Time, bool) {
	if n, ok := v.(int64); ok {
		return parseSloppyRecentTimeInt(n)
	}
	if n, ok := v.(int32); ok {
		return parseSloppyRecentTimeInt(int64(n))
	}
	if n, ok := v.(int); ok {
		return parseSloppyRecentTimeInt(int64(n))
	}
	if n, ok := v.(float64); ok {
		return parseSloppyRecentTimeInt(int64(n))
	}
	if n, ok := v.(float32); ok {
		return parseSloppyRecentTimeInt(int64(n))
	}
	if s, ok := v.(string); ok {
		return parseSloppyRecentTimeString(s)
	}
	return time.Time{}, false
}

func parseSloppyRecentTimeInt(v int64) (time.Time, bool) {
	t := time.Unix(v, 0).UTC()
	if isRecent(t) {
		return t, true
	} else {
		return time.Time{}, false
	}
}

func parseSloppyRecentTimeString(v string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339, time.RFC3339Nano, time.RFC1123, time.RFC1123Z, time.RFC822, time.RFC822Z, time.RFC850,
		time.UnixDate, time.RubyDate, time.DateTime, time.ANSIC,
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, v)
		if err == nil && isRecent(t) {
			return t, true
		}
	}
	return time.Time{}, false
}

func isRecent(t time.Time) bool {
	currentYear := time.Now().Year()
	return t.Year() > currentYear-200 && t.Year() < currentYear+200
}
