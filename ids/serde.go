package ids

import (
	"encoding/hex"
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

func MarshalHash(prefix string, hash []byte) string {
	return fmt.Sprintf("%s_%x", prefix, hash)
}

func MarshalNullableHash(prefix string, hash []byte) string {
	if hash == nil {
		return ""
	}
	return MarshalHash(prefix, hash)
}

func MarshalNullableString(prefix string, id *string) string {
	if id == nil {
		return ""
	}
	return fmt.Sprintf("%s_%s", prefix, *id)
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

func UnmarshalHash(prefix, id string) ([]byte, error) {
	if !strings.HasPrefix(id, prefix+"_") {
		return nil, fmt.Errorf("invalid id")
	}
	base := id[len(prefix)+1:]
	hash, err := hex.DecodeString(base)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}
	return hash, nil
}
