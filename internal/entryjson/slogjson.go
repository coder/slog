package entryjson

import (
	"regexp"
)

// Filter filters the json field f from j.
func Filter(j, f string) string {
	return regexp.MustCompile(`"`+f+`":[^,]+,`).ReplaceAllString(j, "")
}
