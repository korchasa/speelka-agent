package utils

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/ghodss/yaml"
)

func SDump(v interface{}) string {
	ym, _ := yaml.Marshal(v)
	ys := "\n" + string(ym)
	return strings.Replace(ys, "\n", "\n    ", -1)
}

// Now returns the current time
func Now() time.Time {
	return time.Now()
}

// SinceMs returns the elapsed milliseconds since t
func SinceMs(t time.Time) int64 {
	return time.Since(t).Milliseconds()
}

// CompactJson returns a compact JSON string for a map (or any interface{})
func CompactJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Join joins a slice of strings with the given separator
func Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}
