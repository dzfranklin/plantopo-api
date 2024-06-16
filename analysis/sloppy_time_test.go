package analysis

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseSloppyRecentTime(t *testing.T) {
	cases := []struct {
		name     string
		time     interface{}
		expected string
	}{
		{"int", 1718545009, "2024-06-16T13:36:49Z"},
		{"float", 1718545009.0, "2024-06-16T13:36:49Z"},
		{"RFC822", "16 Jun 24 13:36 UTC", "2024-06-16T13:36:00Z"},
		{"nil", nil, ""},
		{"bool", true, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := ParseSloppyRecentTime(c.time)
			if c.expected == "" {
				assert.False(t, ok, "expected to fail")
				return
			}
			assert.True(t, ok, "expected to succeed")
			assert.Equal(t, c.expected, got.Format(time.RFC3339))
		})
	}
}
