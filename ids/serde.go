package ids

import (
	"fmt"
	"strconv"
	"strings"
)

func Marshal(prefix string, id int64) string {
	return fmt.Sprintf("%s_%d", prefix, id)
}

func MarshalNullable(prefix string, id *int64) string {
	if id == nil {
		return ""
	}
	return Marshal(prefix, *id)
}

func Unmarshal(prefix, id string) (int64, error) {
	if !strings.HasPrefix(id, prefix+"_") {
		return 0, fmt.Errorf("invalid id")
	}
	parsed, err := strconv.ParseInt(id[len(prefix)+1:], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id")
	}
	return parsed, nil
}
