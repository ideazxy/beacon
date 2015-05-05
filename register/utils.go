package register

import (
	"fmt"
	"strings"
)

func appendPrefix(key, prefix string) string {
	if prefix != "" {
		return fmt.Sprintf("/%s%s", strings.Trim(prefix, "/"), key)
	}
	return key
}
