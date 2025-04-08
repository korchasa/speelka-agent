package utils

import (
	"strings"

	"github.com/ghodss/yaml"
)

func SDump(v interface{}) string {
	ym, _ := yaml.Marshal(v)
	ys := "\n" + string(ym)
	return strings.Replace(ys, "\n", "\n    ", -1)
}
