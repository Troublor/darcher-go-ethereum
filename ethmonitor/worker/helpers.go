package worker

import (
	"fmt"
	"strings"
)

func PrettifyHash(hash string) string {
	offset := 0
	if strings.Index(hash, "0x") == 0 {
		offset = 2
	}
	return fmt.Sprintf("%s…%s", hash[offset:offset+6], hash[len(hash)-6:])
}
